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

// NewWebService ...
func NewWebService(ctx context.Context, registryConfig config.Registry, opts ...web.Option) web.Service {
	opt := registry.Option(func(opts *registry.Options) {
		opts.Addrs = strings.Split(registryConfig.Address, ",")
	})
	registryOpt := etcd.NewRegistry(opt)

	defaultOpts := []web.Option{
		web.Name(registryConfig.Name),
		web.Version(registryConfig.Version),
		web.Registry(registryOpt),
		web.RegisterTTL(time.Second * time.Duration(registryConfig.Ttl)),
		web.RegisterInterval(time.Second * time.Duration(registryConfig.Interval)),
		web.Metadata(registryConfig.MetaData),
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
