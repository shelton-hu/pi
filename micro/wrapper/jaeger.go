package wrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/shelton-hu/logger"
	"github.com/shelton-hu/util/protoutil"
)

// rpc client call wrapper
func TracerCallWrapper() client.CallWrapper {
	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, resp interface{}, opts client.CallOptions) error {
			span := opentracing.SpanFromContext(ctx)
			request, _ := json.Marshal(req.Body())
			if span != nil {
				ext.SpanKindRPCClient.Set(span)
				ext.PeerAddress.Set(span, node.Address)
				ext.PeerHostname.Set(span, node.Id)
				span.LogKV("request", string(request))
			}

			begin := time.Now()

			err := cf(ctx, node, req, resp, opts)

			end := time.Now()

			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())

			afterWrapper(span, ctx, name, request, resp, float64(end.Sub(begin))/1e6, err)

			return err
		}
	}
}

// rpc server handler wrapper
func TracerHandlerWrapper() server.HandlerWrapper {
	return func(hf server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, resp interface{}) error {
			span := opentracing.SpanFromContext(ctx)
			request, _ := json.Marshal(req.Body())
			if span != nil {
				ext.SpanKindRPCServer.Set(span)
				span.LogKV("request", string(request))
			}

			begin := time.Now()

			err := hf(ctx, req, resp)

			end := time.Now()

			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())

			afterWrapper(span, ctx, name, request, resp, float64(end.Sub(begin))/1e6, err)

			return err
		}
	}
}

func afterWrapper(span opentracing.Span, ctx context.Context, name string, request []byte, resp interface{}, spend float64, err error) {
	response, _ := protoutil.ParseProtoToString(resp.(proto.Message))
	errStr := ""
	if span != nil {
		if err != nil {
			ext.SamplingPriority.Set(span, 1)
			errStr = err.Error()
			ext.Error.Set(span, true)
			span.LogKV("error_msg", errStr)
		}
		span.LogKV("response", response)
	}
	params := string(request)
	// 部分大的字段屏蔽
	if strings.Contains(strings.ToLower(name), "collection") {
		params = "*"
	}

	logger.Info(ctx, "%s, %s, %v, %s, %f", name, params, response, errStr, spend)
}

// rpc client recover wrapper
func RecoverCallWrapper() client.CallWrapper {
	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, resp interface{}, opts client.CallOptions) error {
			//异常处理，防止直接程序挂掉
			defer func() {
				if p := recover(); p != nil {
					//打印调用栈信息
					s := make([]byte, 2048)
					n := runtime.Stack(s, false)
					logger.Error(ctx, "rpc client exception, %s, %s", p, s[:n])
				}
			}()

			return cf(ctx, node, req, resp, opts)
		}
	}
}

// rpc server recover wrapper
func RecoverHandlerWrapper() server.HandlerWrapper {
	return func(hf server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, resp interface{}) error {
			// 异常处理，防止直接程序挂掉
			defer func() {
				if p := recover(); p != nil {
					// 打印调用栈信息
					s := make([]byte, 2048)
					n := runtime.Stack(s, false)
					logger.Error(ctx, "rpc server exception, %s, %s", p, s[:n])
				}
			}()

			return hf(ctx, req, resp)
		}
	}
}
