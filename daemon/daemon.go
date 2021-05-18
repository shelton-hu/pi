package daemon

import (
	"context"
	"math"
	"time"

	"github.com/shelton-hu/nb/goroutine"
)

const (
	// _BaseDuration is the default value of baseDuration.
	_BaseDuration = 125 * time.Millisecond

	// _MaxRetryIndex is the default value of maxRetryIndex.
	_MaxRetryIndex = 7

	// _RecountDuration is the default value of recountDuration.
	_RecountDuration = 30 * time.Second
)

// Daemon ...
type Daemon struct {
	// baseDuration is the base duration between daemon be killed and startted.
	// retryDuration = baseDuration * ( 2 ^ retryIndex).
	baseDuration time.Duration

	// maxRetryIndex is the max value of retryIndex.
	maxRetryIndex uint

	// recountDuration ...
	// When daemonFuncs run longer then recountDuration, retryIndex will be recount.
	recountDuration time.Duration

	// daemonFuncs is the functions which registered into here.
	daemonFuncs []*DaemonFunc

	// context ...
	ctx context.Context
}

// DaemonFunc ...
type DaemonFunc struct {
	// fn is the function body of DaemonFunc.
	fn interface{}

	// args is the function in args of DaemonFunc.
	args []interface{}
}

// NewDaemon ...
func NewDaemon(ctx context.Context) *Daemon {
	return &Daemon{
		baseDuration:    _BaseDuration,
		maxRetryIndex:   _MaxRetryIndex,
		recountDuration: _RecountDuration,

		ctx: ctx,
	}
}

// SetBaseDuration ...
func (d *Daemon) SetBaseDuration(baseDuration time.Duration) {
	d.baseDuration = baseDuration
}

// SetMaxRetryIndex ...
func (d *Daemon) SetMaxRetryIndex(i uint) {
	d.maxRetryIndex = i
}

// SetRecountDuration ...
func (d *Daemon) SetRecountDuration(recountDuration time.Duration) {
	d.recountDuration = recountDuration
}

// Register ...
func (d *Daemon) Register(function interface{}, args ...interface{}) {
	d.daemonFuncs = append(d.daemonFuncs, &DaemonFunc{
		fn:   function,
		args: args,
	})
}

// Run ...
// Which will block.
func (d *Daemon) Run() {
	d.run()
	select {}
}

// RunWithoutBlock ...
// Which will not block.
func (d *Daemon) RunWithoutBlock() {
	d.run()
}

// run ...
func (d *Daemon) run() {
	for i := range d.daemonFuncs {
		go func(d *Daemon, i int) {
			daemonFunc := d.daemonFuncs[i]

			var count int
			for {
				begin := time.Now()
				goroutine.Go(daemonFunc.fn, daemonFunc.args...)
				end := time.Now()
				duration := end.Sub(begin)

				if duration > d.recountDuration {
					count = 0
				}

				sleepDuration := d.baseDuration * time.Duration(math.Pow(2, math.Min(float64(count), float64(d.maxRetryIndex))))
				time.Sleep(sleepDuration)

				count++
			}
		}(d, i)
	}
}
