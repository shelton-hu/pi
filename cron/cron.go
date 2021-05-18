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

// shutdownSignal is the os signal what's need exit the exec gracefully.
var shutdownSignal = []os.Signal{
	syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL,
}

// Cron ...
type Cron struct {
	// cr is the cron instance.
	cr *cron.Cron

	// jobs is the task list of the cron.
	jobs map[string]*Job

	// cfg is the config if the cron.
	cfg *config.SystemConfig

	// context ...
	ctx context.Context
}

type Job struct {
	// spec of the job.
	spec *config.CronSpec

	// fn is the function of the job.
	fn JobFn

	// entryIds list of the job.
	entryIds []cron.EntryID
}

// JobFn is the function of the job.
type JobFn func(ctx context.Context)

// NewCron ...
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

// Register ...
func (c *Cron) Register(jobName string, jobFn JobFn) {
	c.jobs[jobName] = &Job{
		spec: &config.CronSpec{},
		fn:   jobFn,
	}
}

// Run run the cron. It has completed watching c.ctx.Done() and four os signals what are
// syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL for exiting gracfully
func (c *Cron) Run() {
	go c.run()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, shutdownSignal...)
	select {
	case <-ch:
	case <-c.ctx.Done():
	}

	<-c.cr.Stop().Done()
}

// run ...
func (c *Cron) run() {
	for {
		newSpecs := c.getSpecsFromCfg()

		for k, v := range c.jobs {
			name, job := k, v
			newSpec := newSpecs[name]

			if !c.verifySpec(newSpec) {
				continue
			}
			if c.equalSpec(job.spec, newSpec) {
				continue
			}

			for _, id := range job.entryIds {
				c.cr.Remove(id)
			}
			job.entryIds = []cron.EntryID{}

			if newSpec.Suspend {
				job.spec = newSpec
				continue
			}

			cmd := func() {
				ctx, span, _ := microTracing.StartSpanFromContext(context.Background(), opentracing.GlobalTracer(), "cron: "+name)
				defer span.Finish()
				sepc, _ := json.Marshal(job.spec)
				span.LogKV("spec", string(sepc))

				defer func() {
					if e := recover(); e != nil {
						span.SetTag("job.result", "fail")
						ext.Error.Set(span, true)
						span.LogKV("err", e.(string))
					}
				}()

				job.fn(ctx)

				span.SetTag("job.result", "success")
			}

			for i := 0; i < newSpec.Parallelism; i++ {
				entryId, err := c.cr.AddFunc(newSpec.Schedule, cmd)
				if err != nil {
					continue
				}
				job.entryIds = append(job.entryIds, entryId)
			}

			job.spec = newSpec

			c.cr.Start()
		}

		time.Sleep(1 * time.Second)
	}
}

// getSpecsFromCfg ...
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

// verifySpec returns false when spec is nil or schedule is empty or scheduleis invalid.
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

// equalSpec returns false when old spec is not equal to new spec.
func (c *Cron) equalSpec(oldSpec, newSpec *config.CronSpec) bool {
	return oldSpec.Schedule == newSpec.Schedule && oldSpec.Suspend == newSpec.Suspend && oldSpec.Parallelism == newSpec.Parallelism
}
