package redis

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/shelton-hu/logger"
)

type Handler func(channel string, message []byte) error

/**
 * 发布
 */
func (r *Redis) Publish(channel string, message string) (int, error) {
	n, err := redis.Int(r.do("PUBLISH", channel, message))
	if err != nil {
		return 0, err
	}
	return n, nil
}

/**
 * 订阅
 */
func (r *Redis) Subscribe(channels []string, handlers map[string]Handler) error {
	defer r.conn.Close()

	if err := r.conn.Subscribe(redis.Args{}.AddFlat(channels)...); err != nil {
		return err
	}
	defer func() {
		if err := r.conn.Unsubscribe(); err != nil {
			logger.Error(r.ctx, err.Error())
		}
	}()

	go func() {
		<-r.ctx.Done()
		if err := r.conn.Unsubscribe(); err != nil {
			logger.Error(r.ctx, err.Error())
		}
	}()

	done := make(chan error, 1)
	// start a new goroutine to receive message
	logger.Info(r.ctx, "start a new goroutine to reveive message, channel: %v", channels)

	go func() {
		for {
			switch msg := r.conn.Receive().(type) {
			case error:
				done <- fmt.Errorf("redis pubsub receive err: %v", msg)
				//fmt.Printf("redis pubsub receive err: %v", msg)
				return
			case redis.Message:
				if handle, ok := handlers[msg.Channel]; ok {
					if err := handle(msg.Channel, msg.Data); err != nil {
						delete(handlers, msg.Channel)
						_ = r.conn.Unsubscribe(msg.Channel)
						done <- err
						return
					}
				}
			case redis.Subscription:
				if msg.Count == 0 {
					// all channels are unsubscribed
					done <- nil
					return
				}
			}
		}
	}()

	// health check
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		select {
		case err := <-done:
			return err
		case <-tick.C:
			if err := r.conn.Ping(""); err != nil {
				return err
			}
		}
	}
}
