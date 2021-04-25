package service

import (
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
	"github.com/shelton-hu/pi/micro/wrapper"
)

func NewRpcService(registryConfig config.Registry, opts ...micro.Option) micro.Service {
	opt := registry.Option(func(opts *registry.Options) {
		opts.Addrs = strings.Split(registryConfig.Address, ",")
	})
	registryOpt := etcd.NewRegistry(opt)

	defaultOpts := []micro.Option{
		micro.Name(registryConfig.Name),                                              //微服务名称
		micro.Version(registryConfig.Version),                                        //微服务版本
		micro.Registry(registryOpt),                                                  //注册微服务
		micro.RegisterTTL(time.Second * time.Duration(registryConfig.Ttl)),           //微服务发现组件中的节点存活时间
		micro.RegisterInterval(time.Second * time.Duration(registryConfig.Interval)), //微服务发现组件中的节点刷新间隔
		micro.Metadata(registryConfig.MetaData),                                      //微服务自身元数据，会上报给服务发现组件。是一个 key-value 列表
		micro.Flags(&cli.StringFlag{ //自定义命令行参数，必须要有，否则无法接收启动时的参数
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
			wrapper.RecoverHandlerWrapper(), //必须放在第一个
			wrapperTracing.NewHandlerWrapper(opentracing.GlobalTracer()),
			wrapper.TracerHandlerWrapper(),
			prometheus.NewHandlerWrapper(),
		),
		micro.WrapSubscriber(wrapperTracing.NewSubscriberWrapper(opentracing.GlobalTracer())),
	}

	opts = append(defaultOpts, opts...)

	service := micro.NewService(opts...)

	return service
}
