package redis

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	microTracing "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/shelton-hu/logger"
	"github.com/shelton-hu/util/scrutil"
)

const (
	// _RedisComponent is used for jeager record's tag, what
	// is `component=redigo`.
	_RedisComponent = "redigo"

	// _RedisPeerService is used for jeager record's tag, what
	// is `peer_service=redis`.
	_RedisPeerService = "redis"
)

// buildKey returns the key with prefix used for redis cmd.
func (r *Redis) buildKey(key string) string {
	if len(r.rdsKeyPrefix) > 0 {
		key = r.rdsKeyPrefix + ":" + key
	}
	return scrutil.MD5(key)
}

// do reimplement the `do` method of the redis.
func (r *Redis) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	// recover panic
	defer func() {
		if err := recover(); err != nil {
			logger.Error(r.ctx, "%v", err)
		}
	}()

	// close conn
	defer r.conn.Close()

	// check args
	if len(args) == 0 {
		return nil, errors.New("the numbers of redis command args < 1")
	}

	// start span
	_, span, err := microTracing.StartSpanFromContext(r.ctx, opentracing.GlobalTracer(), "Redis")
	if err != nil {
		return nil, err
	}
	defer span.Finish()

	// get key from args
	var logVal string
	var ok bool
	if logVal, ok = args[0].(string); !ok {
		return nil, errors.New("the first arg type of redis command is not string")
	}

	// buildKey and set logKey value
	var logKey string
	if commandName != "PUBLISH" {
		logKey = "key"
		args[0] = interface{}(r.buildKey(logVal))
	} else {
		logKey = "channel"
	}

	// redis do
	reply, err = r.conn.Conn.Do(commandName, args...)

	// set span
	ext.Component.Set(span, _RedisComponent)
	ext.PeerService.Set(span, _RedisPeerService)
	span.SetTag(logKey, logVal)
	span.LogKV("cmd", append([]string{commandName}, r.parseArgs(args)...))
	span.LogKV("res", r.parseArgs(reply))

	return reply, err
}

// parseArgs ...
func (r *Redis) parseArgs(args ...interface{}) []string {
	argStrs := make([]string, 0, len(args))
	for _, arg := range args {
		argStrs = append(argStrs, r.parseArg(arg, 0)...)
	}
	return argStrs
}

// parseArg ...
func (r *Redis) parseArg(arg interface{}, level int) []string {
	if arg == nil {
		return []string{}
	}

	typeOf := reflect.TypeOf(arg)
	valueOf := reflect.ValueOf(arg)
	switch typeOf.Kind() {
	case reflect.Bool:
		if valueOf.Bool() {
			return []string{"1"}
		} else {
			return []string{"0"}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []string{strconv.Itoa(int(valueOf.Int()))}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []string{strconv.Itoa(int(valueOf.Uint()))}
	case reflect.Float32, reflect.Float64:
		return []string{strconv.FormatFloat(float64(valueOf.Float()), 'f', -1, 64)}
	case reflect.Array:
		if level == 0 {
			lenOf := valueOf.Len()
			strs := make([]string, 0, lenOf)
			for i := 0; i < lenOf; i++ {
				strs = append(strs, r.parseArg(valueOf.Index(i).Interface(), level+1)...)
			}
			return strs
		} else {
			var buf bytes.Buffer
			fmt.Fprint(&buf, arg)
			return []string{buf.String()}
		}
	case reflect.Slice:
		if val, ok := arg.([]byte); ok {
			return []string{string(val)}
		}
		if level == 0 {
			lenOf := valueOf.Len()
			strs := make([]string, 0, lenOf)
			for i := 0; i < lenOf; i++ {
				strs = append(strs, r.parseArg(valueOf.Index(i).Interface(), level+1)...)
			}
			return strs
		} else {
			var buf bytes.Buffer
			fmt.Fprint(&buf, arg)
			return []string{buf.String()}
		}
	default:
		var buf bytes.Buffer
		fmt.Fprint(&buf, arg)
		return []string{buf.String()}
	}
}
