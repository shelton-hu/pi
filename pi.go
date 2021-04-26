package pi

import (
	"context"
	"sync"

	goMicro "github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/web"

	"github.com/shelton-hu/pi/config"
	"github.com/shelton-hu/pi/kafka"
	"github.com/shelton-hu/pi/micro"
	"github.com/shelton-hu/pi/mysql"
	"github.com/shelton-hu/pi/redis"
)

var global *Pi
var once sync.Once
var mu sync.Mutex

type Pi struct {
	namespace string
	appName   string

	microRpcService goMicro.Service
	microWebService web.Service
}

type options func(*Option)

type Option struct {
	microRpcOpts []goMicro.Option
	microWebOpts []web.Option
}

func (o *Option) applyOpts(opts ...options) {
	for _, opt := range opts {
		opt(o)
	}
}

func SetMicroRpcOptions(opts ...goMicro.Option) options {
	return func(o *Option) {
		o.microRpcOpts = append(o.microRpcOpts, opts...)
	}
}

func SetMicroWebOptions(opts ...web.Option) options {
	return func(o *Option) {
		o.microWebOpts = append(o.microWebOpts, opts...)
	}
}

type closeFunc struct {
	fns []func()
}

func (c *closeFunc) Close() {
	for _, fn := range c.fns {
		fn()
	}
}

func SetGlobal(ctx context.Context, etcdAddresses []string, namespace, appname string, opts ...options) (closes *closeFunc) {
	o := new(Option)
	o.applyOpts(opts...)

	mu.Lock()
	defer mu.Unlock()

	once.Do(func() {
		global = &Pi{
			namespace: namespace,
			appName:   appname,
		}

		config.InitConfig(ctx, etcdAddresses, global.namespace, global.appName)

		mysql.ConnectMysql(ctx, global.SysConf().Mysql)
		closes.fns = append(closes.fns, func() { mysql.CloseMysql(ctx) })

		redis.ConnectRedis(ctx, global.SysConf().Redis)
		closes.fns = append(closes.fns, func() { redis.CloseRedis(ctx) })

		kafka.ConnectKafka(ctx, global.SysConf().Kafka)
		closes.fns = append(closes.fns, func() { kafka.CloseKafka(ctx) })

		global.microRpcService = micro.NewRpcService(ctx, global.SysConf().Registry, o.microRpcOpts...)
		global.microWebService = micro.NewWebService(ctx, global.SysConf().Registry, o.microWebOpts...)
	})

	return closes
}

func G() *Pi {
	if global == nil {
		panic("pi is not init")
	}
	return global
}

func (p *Pi) Namespace() string {
	return p.namespace
}

func (p *Pi) AppName() string {
	return p.appName
}

func (p *Pi) SysConf() *config.SystemConfig {
	return config.SysConf()
}

func (p *Pi) CusConf() *config.CustomConfig {
	return config.CusConf()
}

func (p *Pi) Mysql(ctx context.Context, name ...string) *mysql.Mysql {
	return mysql.GetConnect(ctx, name...)
}

func (p *Pi) Redis(ctx context.Context) *redis.Redis {
	return redis.GetConnect(ctx)
}

func (p *Pi) Kafka(ctx context.Context) *kafka.Kafka {
	return kafka.GetConnect(ctx)
}

func (p *Pi) MicroRpcService(ctx context.Context) goMicro.Service {
	return p.microRpcService
}

func (p *Pi) MicroWebService(ctx context.Context) web.Service {
	return p.microWebService
}
