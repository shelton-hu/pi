package daemon

import (
	"context"
	"math"
	"time"

	"github.com/shelton-hu/nb/goroutine"
)

type Daemon struct {
	// 重试间隔基数
	// 即一次协程死掉后，到该协程再次被拉起前之间的时间间隔
	// 重试间隔 = 重试间隔基数 * ( 2 ^ 重试间隔指数 )
	baseDuration time.Duration

	// 最大重试间隔指数
	// 无论重试次数如何增长，重试间隔指数不能超过此设定值
	maxRetryIndex uint

	// 重新计数时间间隔
	// 当注册的函数运行超过此时间后，重试次数重新开始计数
	recountDuration time.Duration

	// 注册到此实例的函数列表
	daemonFuncs []*DaemonFunc

	// 上下文
	ctx context.Context
}

type DaemonFunc struct {
	// 注册的函数
	// 该函数会在`*Daemon.Run`或者`*Daemon.RunWithNoBlock`中被执行
	fn interface{}

	// 注册的函数的入参
	args []interface{}
}

const (
	// 默认的重试间隔基数
	BaseDuration = 125 * time.Millisecond

	// 默认的最大重试间隔指数
	MaxRetryIndex = 7

	// 默认的重新计数间隔
	RecountDuration = 30 * time.Second
)

// 新建一个Daemon实例，该实例用于协程守护
func NewDaemon(ctx context.Context) *Daemon {
	return &Daemon{
		baseDuration:    BaseDuration,
		maxRetryIndex:   MaxRetryIndex,
		recountDuration: RecountDuration,

		ctx: ctx,
	}
}

// 设置重试间隔基数
func (d *Daemon) SetBaseDuration(baseDuration time.Duration) {
	d.baseDuration = baseDuration
}

// 设置最大重试间隔指数
func (d *Daemon) SetMaxRetryIndex(i uint) {
	d.maxRetryIndex = i
}

// 设置重新计数间隔
func (d *Daemon) SetRecountDuration(recountDuration time.Duration) {
	d.recountDuration = recountDuration
}

// 注册函数
func (d *Daemon) Register(function interface{}, args ...interface{}) {
	d.daemonFuncs = append(d.daemonFuncs, &DaemonFunc{
		fn:   function,
		args: args,
	})
}

// 执行协程守护，并且会一直阻塞
func (d *Daemon) Run() {
	d.run()
	select {}
}

// 执行协程守护，不会阻塞
func (d *Daemon) RunWithoutBlock() {
	d.run()
}

// 执行协程守护
func (d *Daemon) run() {
	// 每个注册的函数起一个协程
	for i := range d.daemonFuncs {

		go func(d *Daemon, i int) {

			// 取出注册的函数
			daemonFunc := d.daemonFuncs[i]

			// 定义重试计数器
			var count int

			for {
				// 执行注册的函数，并且记录执行时间
				begin := time.Now()
				goroutine.Go(daemonFunc.fn, daemonFunc.args...)
				end := time.Now()
				duration := end.Sub(begin)

				// 当执行时间大于重新计数时间间隔时，清空重试计数器
				if duration > d.recountDuration {
					count = 0
				}

				// 计算协程睡眠时间，并睡眠
				sleepDuration := d.baseDuration * time.Duration(math.Pow(2, math.Min(float64(count), float64(d.maxRetryIndex))))
				time.Sleep(sleepDuration)

				// 重试计数器自增1
				count++
			}
		}(d, i)
	}
}
