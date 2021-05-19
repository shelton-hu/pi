package redis

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/shelton-hu/logger"
)

// LockResourceFunc is function which will be done after lock successfully by redis.
type LockResourceFunc func(ctx context.Context, args ...interface{}) ([]interface{}, error)

// LockOptions ...
type LockOptions func(*LockOption)

// LockOption ...
type LockOption struct {
	autoExpire                  ExpireTime
	maxExpire                   ExpireTime
	retryTimes                  int
	firstrRetryIntervalDuration time.Duration
}

// newLockOption ...
func newLockOption() *LockOption {
	return &LockOption{
		autoExpire:                  60,
		maxExpire:                   4 * ExpireTImeMinute,
		retryTimes:                  10,
		firstrRetryIntervalDuration: 1 * time.Second,
	}
}

// TryLock is a function that which provides thread safety lock method by redis.
// There are three in parameters:
// 		key        uniquely identifies of the resource
// 		lrfn       the function which will be called after lock successfully
// 		lrfnIn     in parameters of lrfn
// 		opts       the options parameters of the TryLock function
// There are two out parameters:
//		irfnOut    out parameters of lrfn
//		err        error
func (r *Redis) TryLock(key string, lrfn LockResourceFunc, lrfnIn []interface{}, opts ...LockOptions) (lrfnOut []interface{}, err error) {
	l := newLockOption()
	for _, opt := range opts {
		opt(l)
	}

	for i := 0; i < l.retryTimes; i++ {
		reply, err := r.SetNX(key, 1, l.autoExpire)
		if err != nil || !reply {
			time.Sleep(l.firstrRetryIntervalDuration * time.Duration(math.Pow(2.0, float64(i))))
			continue
		}

		done := make(chan struct{}, 1)
		go func() {
			timer := time.NewTimer(time.Duration(l.maxExpire) * time.Second)
			for i := 0; ; i++ {
				select {
				case <-done:
					return
				case <-timer.C:
					return
				}
			}
		}()
		defer func() {
			done <- struct{}{}
			if err := r.Delete(key); err != nil {
				logger.Error(r.ctx, err.Error())
			}
		}()

		return lrfn(r.ctx, lrfnIn...)
	}

	return nil, errors.New("try lock failed")
}

// SetLockAutoExpire ...
func SetLockAutoExpire(d ExpireTime) LockOptions {
	return func(l *LockOption) {
		l.autoExpire = d
	}
}

// SetMaxExpire ...
func SetMaxExpire(d ExpireTime) LockOptions {
	return func(l *LockOption) {
		l.maxExpire = d
	}
}

// SetLockRetryTimes ...
func SetLockRetryTimes(times int) LockOptions {
	return func(l *LockOption) {
		l.retryTimes = times
	}
}

// SetLockFirstRetryIntervalDuration ...
func SetLockFirstRetryIntervalDuration(d time.Duration) LockOptions {
	return func(l *LockOption) {
		l.firstrRetryIntervalDuration = d
	}
}
