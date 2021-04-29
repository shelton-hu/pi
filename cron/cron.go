package cron

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	microTracing "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	cron "github.com/robfig/cron/v3"

	"github.com/shelton-hu/pi/config"
)

type Cron struct {
	cr   *cron.Cron           // cron实例
	jobs map[string]*Job      // 任务列表
	cfg  *config.SystemConfig // 配置文件

	ctx context.Context // 上下文
}

type Job struct {
	spec     *config.CronSpec // 定时任务配置
	fn       JobFn            // 定时任务函数
	entryIds []cron.EntryID   // 定时任务entryId列表
}

type JobFn func(ctx context.Context)

var shutdownSignal = []os.Signal{
	syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL,
}

// 新建一个Cron实例，该实例用于定时任务
// 注意这里要传配置文件的指针，以供函数监听配置
func NewCron(ctx context.Context, cfg *config.SystemConfig) *Cron {
	return &Cron{
		cr: cron.New(cron.WithChain(
			//注入此函数是为了防并发
			cron.SkipIfStillRunning(cron.DefaultLogger),
		)),
		jobs: make(map[string]*Job),
		cfg:  cfg,

		ctx: ctx,
	}
}

// 注册定时任务
func (c *Cron) Register(jobName string, jobFn JobFn) {
	c.jobs[jobName] = &Job{
		spec: &config.CronSpec{},
		fn:   jobFn,
	}
}

// 启动定时任务
// 已经实现了监听c.ctx.Done()信号，以及
// syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL
// 4个系统信号实现优雅地退出
// 即等待当前所有的任务执行完以后再返回
func (c *Cron) Run() {
	// 开启协程启动Cron实例
	go c.run()

	// 监听退出信号
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, shutdownSignal...)
	select {
	case <-ch:
	case <-c.ctx.Done():
	}

	// 调用Stop函数，并监听返回的ctx的退出信号
	<-c.cr.Stop().Done()
}

// 启动定时任务
func (c *Cron) run() {
	for {
		// 从配置文件中获取所有任务配置
		newSpecs := c.getSpecsFromCfg()

		// 每个循环，都遍历jobs列表
		for k, v := range c.jobs {

			// 安全地将指针赋值给新的变量
			name, job := k, v

			// 从配置中心获取任务配置
			newSpec := newSpecs[name]

			// 校验配置，不合法(包括schedule不合法，或者配置为空)直接continue
			if !c.verifySpec(newSpec) {
				continue
			}

			// 比较新老配置，相同则直接continue
			if c.equalSpec(job.spec, newSpec) {
				continue
			}

			// 先移除对应job下已注册的func
			for _, id := range job.entryIds {
				c.cr.Remove(id)
			}
			job.entryIds = []cron.EntryID{}

			// 当`job.spec.Suspend`为true时(代表该任务需要暂停)，则直接continue
			if newSpec.Suspend {
				job.spec = newSpec
				continue
			}

			// 定义cmd，供注册func时使用
			cmd := func() {

				// 开启trace
				ctx, span, _ := microTracing.StartSpanFromContext(context.Background(), opentracing.GlobalTracer(), "cron: "+name)
				defer span.Finish()
				sepc, _ := json.Marshal(job.spec)
				span.LogKV("spec", string(sepc))

				// recover以防止panic
				defer func() {
					if e := recover(); e != nil {
						// 记录trace
						span.SetTag("job.result", "fail")
						ext.Error.Set(span, true)
						span.LogKV("err", e.(string))
					}
				}()

				// 执行任务函数
				job.fn(ctx)

				// 记录trace
				span.SetTag("job.result", "success")
			}

			// 注册func
			for i := 0; i < newSpec.Parallelism; i++ {
				entryId, err := c.cr.AddFunc(newSpec.Schedule, cmd)
				if err != nil {
					continue
				}
				job.entryIds = append(job.entryIds, entryId)
			}

			// 为了保证任务一定被注册，最后更新任务配置
			job.spec = newSpec

			// 开启定时任务
			c.cr.Start()
		}

		// 每个循环，sleep 1s
		time.Sleep(1 * time.Second)
	}
}

// 从配置文件读取配置，并转化为map
func (c *Cron) getSpecsFromCfg() map[string]*config.CronSpec {
	specs := make(map[string]*config.CronSpec)
	for _, v := range c.cfg.Cron.Spec {
		spec := &config.CronSpec{
			Name:        v.Name,
			Schedule:    v.Schedule,
			Suspend:     v.Suspend,
			Parallelism: v.Parallelism,
		}
		if spec.Parallelism < 1 {
			spec.Parallelism = 1
		}
		specs[spec.Name] = spec
	}
	return specs
}

// 校验spec，当spec为nil，schedule为空，或者schedule不合法，则返回false
func (c *Cron) verifySpec(spec *config.CronSpec) (pass bool) {
	if spec == nil {
		return false
	}
	if spec.Schedule == "" {
		return false
	}
	if _, err := cron.ParseStandard(spec.Schedule); err != nil {
		return false
	}
	return true
}

// 比较新老配置，当新老配置不一致时，返回false
func (c *Cron) equalSpec(oldSpec, newSpec *config.CronSpec) bool {
	return oldSpec.Schedule == newSpec.Schedule && oldSpec.Suspend == newSpec.Suspend && oldSpec.Parallelism == newSpec.Parallelism
}
