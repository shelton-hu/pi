package nb

import (
	"context"
	"sync"

	"github.com/shelton-hu/pi/config"
	"github.com/shelton-hu/pi/mysql"
	"github.com/shelton-hu/pi/redis"
)

type Pi struct {
	namespace string
	appName   string
}

var global *Pi
var once sync.Once
var mu sync.Mutex

type closeFunc struct {
	fns []func()
}

func (c *closeFunc) Close() {
	for _, fn := range c.fns {
		fn()
	}
}

func SetGlobal(ctx context.Context, etcdAddresses []string, namespace, appname string) (closes *closeFunc) {
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
	})

	return closes
}

func Global() *Pi {
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
