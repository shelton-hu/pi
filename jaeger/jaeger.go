package jaeger

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/uber/jaeger-client-go"
	jaegerConfigure "github.com/uber/jaeger-client-go/config"

	"github.com/shelton-hu/logger"

	"github.com/shelton-hu/pi/config"
)

var j *Jaeger

type Jaeger struct {
	close io.Closer
}

// 获取jaeger实例
func ConnectJaeger(ctx context.Context, jaegerConfig config.Jaeger, opts ...jaegerConfigure.Option) {
	address := fmt.Sprintf("%s:%d", jaegerConfig.Host, jaegerConfig.Port)

	// 配置项参考 https://github.com/jaegertracing/jaeger-client-go/blob/master/config/config.go
	configure := jaegerConfigure.Configuration{
		ServiceName: jaegerConfig.Name,
		Sampler: &jaegerConfigure.SamplerConfig{
			Type:  jaeger.SamplerTypeProbabilistic,
			Param: jaegerConfig.Rate,
		},
		Reporter: &jaegerConfigure.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  address,
		},
	}

	closer, err := configure.InitGlobalTracer(jaegerConfig.Name, opts...)
	if err != nil {
		panic(fmt.Errorf("fatal error: connect jaeger: %s\n", err))
	}

	j = &Jaeger{
		close: closer,
	}
}

func CloseJaeger(ctx context.Context) {
	if err := j.close.Close(); err != nil {
		logger.Error(ctx, err.Error())
	}
}
