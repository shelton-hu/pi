package micro

import (
	"context"
	"strings"
	"time"

	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-plugins/wrapper/monitoring/prometheus/v2"
	wrapperTracing "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	opentracing "github.com/opentracing/opentracing-go"

	"github.com/shelton-hu/pi/config"
)

// NewRpcService ...
func NewRpcService(ctx context.Context, registryConfig config.Registry, opts ...micro.Option) micro.Service {
	opt := registry.Option(func(opts *registry.Options) {
		opts.Addrs = strings.Split(registryConfig.Address, ",")
	})
	registryOpt := etcd.NewRegistry(opt)

	defaultOpts := []micro.Option{
		micro.Name(registryConfig.Name),
		micro.Version(registryConfig.Version),
		micro.Registry(registryOpt),
		micro.RegisterTTL(time.Second * time.Duration(registryConfig.Ttl)),
		micro.RegisterInterval(time.Second * time.Duration(registryConfig.Interval)),
		micro.Metadata(registryConfig.MetaData),
		micro.Flags(&cli.StringFlag{
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
			Usage: "etcd address. the default valu is for env",
		}),
		micro.WrapHandler(
			recoverHandlerWrapper(),
			wrapperTracing.NewHandlerWrapper(opentracing.GlobalTracer()),
			traceHandlerWrapper(),
			prometheus.NewHandlerWrapper(),
		),
		micro.WrapSubscriber(wrapperTracing.NewSubscriberWrapper(opentracing.GlobalTracer())),
	}

	opts = append(defaultOpts, opts...)
	service := micro.NewService(opts...)

	return service
}
