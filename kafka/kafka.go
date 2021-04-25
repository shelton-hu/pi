package kafka

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	microTracing "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/shelton-hu/logger"
)

const (
	_KafkaDefautGroupId = "default"
	_KafkaPartition     = "partition"
	_KafkaOffset        = "offset"
	_KafkaConsumer      = "kafka.consumer"
	_KafkaProducer      = "kafka.producer"
	_KafkaComponent     = "go-kafka"
	_KafkaPeerService   = "kafka"
)

type Handler func(ctx context.Context, msg []byte) error

type SubscribeOptions func(s *SubscribeOption)

type SubscribeOption struct {
	groupId string
}

func (k *Kafka) Publish(ctx context.Context, topic string, m []byte) (err error) {
	headers := make(map[string]string)
	// build msg head by jaeger start
	name := fmt.Sprintf("%s.%s", _KafkaProducer, topic)
	_, span, err := microTracing.StartSpanFromContext(ctx, opentracing.GlobalTracer(), name)
	if err == nil {
		// 设置kafka特有标签
		ext.SpanKindProducer.Set(span)
		ext.Component.Set(span, _KafkaComponent)
		ext.PeerService.Set(span, _KafkaPeerService)
		ext.MessageBusDestination.Set(span, topic)
		// span注入header
		_ = opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(headers))
		span.LogKV("data", string(m))
		defer span.Finish()
	}
	// 组装message
	msg := &sarama.ProducerMessage{}
	// kafka从0.11版本开始支持header
	if k.cfg.Version.IsAtLeast(sarama.V0_11_0_0) {
		for key, value := range headers {
			header := sarama.RecordHeader{
				Key:   []byte(key),
				Value: []byte(value),
			}
			msg.Headers = append(msg.Headers, header)
		}
	}
	msg.Topic = topic
	msg.Value = sarama.ByteEncoder(m)

	k.p.Input() <- msg
	select {
	case <-k.p.Successes():
		return
	case fail := <-k.p.Errors():
		return fail.Err
	}
}

func (k *Kafka) Subscribe(ctx context.Context, topic string, handler Handler, opts ...SubscribeOptions) error {
	// apply opts
	s := newSubscribeOption()
	s.applyOpts(opts...)

	// create consumer peer subscribe
	topics := []string{topic}
	consumer, err := cluster.NewConsumer(k.addrs, s.groupId, topics, k.clcfg)
	if err != nil {
		logger.Error(ctx, err.Error())
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Error(ctx, err.Error())
		}
	}()

	// trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// consume errors
	go func(ctx context.Context) {
		for err := range consumer.Errors() {
			logger.Error(ctx, err.Error())
			return
		}
	}(ctx)

	// consume notifications
	go func(ctx context.Context) {
		for ntf := range consumer.Notifications() {
			logger.Info(ctx, "rebalanced: %+v", ntf)
		}
	}(ctx)

	// consume messages, watch signals
	for {
		select {
		case msg, ok := <-consumer.Messages():
			if !ok {
				continue
			}
			span := opentracing.SpanFromContext(ctx)
			name := fmt.Sprintf("%s.%s", _KafkaConsumer, topic)
			if span == nil {
				// 解析header到jaeger中
				headers := make(map[string]string)
				// kafka从0.11版本开始支持header
				if k.cfg.Version.IsAtLeast(sarama.V0_11_0_0) {
					for _, header := range msg.Headers {
						headers[string(header.Key)] = string(header.Value)
					}
				}
				spanContext, _ := opentracing.GlobalTracer().Extract(
					opentracing.TextMap,
					opentracing.TextMapCarrier(headers),
				)
				if spanContext != nil {
					ctx, span, _ = microTracing.StartSpanFromContext(ctx, opentracing.GlobalTracer(), name, opentracing.ChildOf(spanContext))
				} else {
					ctx, span, _ = microTracing.StartSpanFromContext(ctx, opentracing.GlobalTracer(), name)
				}
			}
			ext.Component.Set(span, _KafkaComponent)
			ext.PeerService.Set(span, _KafkaPeerService)

			span.SetTag(_KafkaPartition, msg.Partition)
			span.SetTag(_KafkaOffset, msg.Offset)

			span.LogKV("data", string(msg.Value))

			// handler msg
			err = handler(ctx, msg.Value)
			if err != nil {
				ext.Error.Set(span, true)
				span.LogKV("error_msg", err.Error())
			}

			// 不能通过defer的方式声明
			span.Finish()

			// mark message as processed
			consumer.MarkOffset(msg, "")
		case <-signals:
			return nil
		}
	}
}

func newSubscribeOption() *SubscribeOption {
	return &SubscribeOption{
		groupId: _KafkaDefautGroupId,
	}
}

func (s *SubscribeOption) applyOpts(opts ...SubscribeOptions) {
	for _, opt := range opts {
		opt(s)
	}
}

func SetSubscribeGroupId(groupId string) SubscribeOptions {
	return func(s *SubscribeOption) {
		s.groupId = groupId
	}
}
