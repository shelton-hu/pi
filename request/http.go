package request

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	microTrace "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/shelton-hu/logger"
	"github.com/shelton-hu/util/netutil"
)

const (
	_ContentType = "Content-Type"

	_ContentTypeJson = "application/json"
	_ContentTypeUrl  = "application/x-www-form-urlencoded"
)

type HttpClient struct {
	host    string
	path    string
	header  map[string]string
	timeout time.Duration

	ctx context.Context
}

type HttpOptions func(*HttpClient)

func NewHttpClient(ctx context.Context, host string, opts ...HttpOptions) *HttpClient {
	c := &HttpClient{
		host:    host,
		header:  make(map[string]string),
		timeout: 5 * time.Second,
		ctx:     ctx,
	}
	c.applyOpts(opts...)

	return c
}

func (c *HttpClient) applyOpts(opts ...HttpOptions) {
	for _, opt := range opts {
		opt(c)
	}
}

func (c *HttpClient) GetUrl() string {
	return c.host + c.path
}

func (c *HttpClient) Get(params map[string]string, returnObj interface{}, opts ...HttpOptions) error {
	// applt opts
	c.applyOpts(opts...)

	// build url
	url, err := netutil.BuildUrl(c.GetUrl(), params)
	if err != nil {
		logger.Error(c.ctx, err.Error())
		return err
	}

	// do request
	resp, code, err := c.request(http.MethodGet, url, nil)
	if err != nil {
		logger.Error(c.ctx, err.Error())
		return err
	}

	// check http status code
	if !isOk(code) {
		return fmt.Errorf("request exception: [%d]%s", code, string(resp))
	}

	// unmarshal response
	if returnObj != nil {
		if err := json.Unmarshal(resp, &returnObj); err != nil {
			logger.Error(c.ctx, err.Error())
			return err
		}
	}

	return nil
}

func (c *HttpClient) Post(reqBody interface{}, returnObj interface{}, opts ...HttpOptions) error {
	// apply opts
	c.applyOpts(opts...)

	// check header
	contentType, ok := c.header["Content-Type"]
	if !ok {
		return errors.New("header not contains Content-Type")
	}

	// build request body
	var r []byte
	switch contentType {
	case _ContentTypeJson:
		r, ok = reqBody.([]byte)
		if !ok {
			err := errors.New("request body is illegal")
			logger.Error(c.ctx, err.Error())
			return err
		}
	case _ContentTypeUrl:
		params, ok := reqBody.(map[string]string)
		if !ok {
			err := errors.New("request body is illegal")
			logger.Error(c.ctx, err.Error())
			return err
		}
		r = []byte(netutil.BuildQuery(params))
	default:
		err := errors.New("request body is illegal")
		logger.Error(c.ctx, err.Error())
		return err
	}

	// do request
	resp, code, err := c.request(http.MethodPost, c.GetUrl(), r)
	if err != nil {
		logger.Error(c.ctx, err.Error())
		return err
	}

	// check http status code
	if !isOk(code) {
		return fmt.Errorf("request exception: [%d]%s", code, string(resp))
	}

	// unmarshal response
	if returnObj != nil {
		if err := json.Unmarshal(resp, &returnObj); err != nil {
			logger.Error(c.ctx, err.Error())
			return err
		}
	}

	return nil

}

func (c *HttpClient) PostJson(reqBody []byte, returnObj interface{}, opts ...HttpOptions) error {
	opts = append(opts, SetHeader(map[string]string{_ContentType: _ContentTypeJson}))

	if err := c.Post(reqBody, returnObj, opts...); err != nil {
		logger.Error(c.ctx, err.Error())
		return err
	}

	return nil
}

func (c *HttpClient) PostForm(reqBody map[string]string, returnObj interface{}, opts ...HttpOptions) error {
	opts = append(opts, SetHeader(map[string]string{_ContentType: _ContentTypeUrl}))

	if err := c.Post(reqBody, returnObj, opts...); err != nil {
		logger.Error(c.ctx, err.Error())
		return err
	}

	return nil
}

func (c *HttpClient) request(method, url string, reqBody []byte) (respBody []byte, code int, err error) {
	// recover panic
	defer func() {
		if e := recover(); e != nil {
			logger.Error(c.ctx, "http request exception, %s: %s", url, e)
		}
	}()

	// define request
	req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	if err != nil {
		panic(err)
	}

	// set header of request
	for key, val := range c.header {
		req.Header.Set(key, val)
	}

	// start span
	name := fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.EscapedPath())
	_, span, err := microTrace.StartSpanFromContext(c.ctx, opentracing.GlobalTracer(), name)
	if err != nil {
		panic(err)
	}
	defer span.Finish()

	// define client
	client := &http.Client{
		Timeout: c.timeout,
	}

	// do request
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// read resp.Body
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	// set span tag
	ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode))
	ext.HTTPMethod.Set(span, method)
	if resp.StatusCode >= http.StatusBadRequest {
		ext.SamplingPriority.Set(span, 1)
		ext.Error.Set(span, true)
	}

	// set span log
	span.LogKV("request", string(reqBody))
	span.LogKV("response", string(respBody))

	return respBody, resp.StatusCode, nil
}

func SetPath(path string) HttpOptions {
	return func(c *HttpClient) {
		c.path = path
	}
}

func SetHeader(header map[string]string) HttpOptions {
	return func(c *HttpClient) {
		for key, val := range header {
			c.header[key] = val
		}
	}
}

func SetTimeout(d time.Duration) HttpOptions {
	return func(c *HttpClient) {
		c.timeout = d
	}
}

func isOk(code int) bool {
	return code == http.StatusOK || code == http.StatusCreated
}