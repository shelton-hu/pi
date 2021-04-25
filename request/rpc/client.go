package service

import (
	"strings"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/micro/go-micro/v2/client/selector"
	clientRegistry "github.com/micro/go-micro/v2/client/selector/registry"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	wrapperTracing "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	opentracing "github.com/opentracing/opentracing-go"

	"github.com/shelton-hu/pi/config"
	"github.com/shelton-hu/pi/micro/wrapper"
)

func NewClient(key string, registryConfig config.Registry, opts ...client.Option) client.Client {
	opt := registry.Option(func(opts *registry.Options) {
		opts.Addrs = strings.Split(registryConfig.Address, ",")
	})
	registryOpt := etcd.NewRegistry(opt)

	defaultOpts := []client.Option{
		client.Registry(registryOpt),
		client.Selector(clientRegistry.NewSelector(selector.Registry(registryOpt))),
		client.WrapCall(
			wrapper.RecoverCallWrapper(), //必须放在第一个
			wrapperTracing.NewCallWrapper(opentracing.GlobalTracer()),
			wrapper.TracerCallWrapper(),
		),
		client.Retries(0),
	}

	//合并选项
	opts = append(defaultOpts, opts...)

	return grpc.NewClient(opts...)
}
