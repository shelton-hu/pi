package micro

import (
	"context"
	"strings"
	"time"

	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/web"

	"github.com/shelton-hu/pi/config"
)

func NewWebService(ctx context.Context, registryConfig config.Registry, opts ...web.Option) web.Service {
	opt := registry.Option(func(opts *registry.Options) {
		opts.Addrs = strings.Split(registryConfig.Address, ",")
	})
	registryOpt := etcd.NewRegistry(opt)

	defaultOpts := []web.Option{
		web.Name(registryConfig.Name),                                              // 微服务名称
		web.Version(registryConfig.Version),                                        // 微服务版本
		web.Registry(registryOpt),                                                  // 注册微服务
		web.RegisterTTL(time.Second * time.Duration(registryConfig.Ttl)),           // 微服务发现组件中的节点存活时间
		web.RegisterInterval(time.Second * time.Duration(registryConfig.Interval)), // 微服务发现组件中的节点刷新间隔
		web.Metadata(registryConfig.MetaData),                                      // 微服务自身元数据，会上报给服务发现组件。是一个 key-value 列表
		web.Flags(&cli.StringFlag{
			Name:  "env",
			Value: "dev",
			Usage: "app runtime environment. the default value is dev",
		}, &cli.StringFlag{
			Name:  "localConfig",
			Value: "",
			Usage: "app config file path. the default value is get by env",
		}, &cli.StringFlag{
			Name:  "etcdAddrs",
			Value: _DefaultDevRegistry,
			Usage: "etcd address. the default value is for env",
		}),
	}
	//合并选项
	opts = append(defaultOpts, opts...)
	service := web.NewService(opts...)

	return service
}
