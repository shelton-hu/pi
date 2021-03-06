package micro

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/micro/go-micro/v2/server"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/shelton-hu/logger"
	"github.com/shelton-hu/util/protoutil"
)

// recoverHandlerWrapper ...
func recoverHandlerWrapper() server.HandlerWrapper {
	return func(hf server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, resp interface{}) error {
			defer func() {
				if p := recover(); p != nil {
					s := make([]byte, 2048)
					n := runtime.Stack(s, false)
					logger.Error(ctx, "rpc server exception, %s, %s", p, s[:n])
				}
			}()
			return hf(ctx, req, resp)
		}
	}
}

// traceHandlerWrapper ...
func traceHandlerWrapper() server.HandlerWrapper {
	return func(hf server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, resp interface{}) error {
			span := opentracing.SpanFromContext(ctx)
			request, err := json.Marshal(req.Body())
			if err != nil {
				logger.Error(ctx, err.Error())
			}
			if span != nil {
				ext.SpanKindRPCServer.Set(span)
				span.LogKV("request", string(request))
			}
			begin := time.Now()
			err = hf(ctx, req, resp)
			end := time.Now()
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			afterWrapper(span, ctx, name, request, resp, float64(end.Sub(begin))/1e6, err)
			return err
		}
	}
}

// afterWrapper ...
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
	if strings.Contains(strings.ToLower(name), "collection") {
		params = "*"
	}

	logger.Info(ctx, "%s, %s, %v, %s, %f", name, params, response, errStr, spend)
}
