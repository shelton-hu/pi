package broker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

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

type kBroker struct {
	p       sarama.AsyncProducer
	c       sarama.Client
	sc      []sarama.Client
	scMutex sync.Mutex
	clg     *cluster.Config
	addrs   []string
	v       sarama.KafkaVersion
}

type SubscribeOption func(*SubscribeOptions)

type SubscribeOptions struct {
	GroupId string

	Context context.Context
}

var Kafka kBroker //业务

var CKafka kBroker //日志

type Handler func(ctx context.Context, msg []byte) error

//初始化配置
func initConfig(version sarama.KafkaVersion, broker *kBroker) *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Version = version
	//设置一下版本
	broker.v = config.Version
	return config
}

func Connect(addresses []string, version sarama.KafkaVersion, broker *kBroker) (err error) {

	config := initConfig(version, broker)

	//client
	c, err := sarama.NewClient(addresses, config)
	if err != nil {
		return
	}
	broker.c = c

	//create producer from client
	p, err := sarama.NewAsyncProducerFromClient(c)
	if err != nil {
		return
	}
	broker.p = p

	broker.scMutex.Lock()
	defer broker.scMutex.Unlock()
	broker.sc = make([]sarama.Client, 0)

	//consumer
	broker.addrs = addresses

	broker.clg = cluster.NewConfig()
	broker.clg.Consumer.Offsets.Initial = sarama.OffsetNewest
	broker.clg.Group.Return.Notifications = true
	broker.clg.Version = config.Version
	return
}

func Disconnect(broker *kBroker) error {
	broker.scMutex.Lock()
	defer broker.scMutex.Unlock()
	for _, client := range broker.sc {
		client.Close()
	}
	broker.sc = nil
	broker.p.Close()
	return broker.c.Close()
}

func (kb *kBroker) Publish(ctx context.Context, topic string, m []byte) (err error) {
	headers := make(map[string]string)
	//build msg head by jaeger start
	name := fmt.Sprintf("%s.%s", _KafkaProducer, topic)
	_, span, err := microTracing.StartSpanFromContext(ctx, opentracing.GlobalTracer(), name)
	if err == nil {
		//设置kafka特有标签
		ext.SpanKindProducer.Set(span)
		ext.Component.Set(span, _KafkaComponent)
		ext.PeerService.Set(span, _KafkaPeerService)
		ext.MessageBusDestination.Set(span, topic)
		//span注入header
		_ = opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(headers))
		span.LogKV("data", string(m))
		defer span.Finish()
	}
	//组装message
	msg := &sarama.ProducerMessage{}
	//kafka从0.11版本开始支持header
	if kb.v.IsAtLeast(sarama.V0_11_0_0) {
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

	kb.p.Input() <- msg
	select {
	case <-kb.p.Successes():
		return
	case fail := <-kb.p.Errors():
		return fail.Err
	}
}

func (kb *kBroker) Subscribe(topic string, handler Handler, opts ...SubscribeOption) (err error) {
	//trans opts
	opt := SubscribeOptions{
		GroupId: _KafkaDefautGroupId,
	}
	for _, o := range opts {
		o(&opt)
	}
	if opt.Context == nil {
		opt.Context = context.TODO()
	}
	//create consumer peer subscribe
	topics := []string{topic}
	consumer, err := cluster.NewConsumer(kb.addrs, opt.GroupId, topics, kb.clg)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	// trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// consume errors
	go func() {
		for err = range consumer.Errors() {
			return
		}
	}()

	// consume notifications
	go func(ctx context.Context) {
		for ntf := range consumer.Notifications() {
			logger.Info(opt.Context, "rebalanced: %+v", ntf)
		}
	}(opt.Context)

	// consume messages, watch signals
	for {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				ctx := opt.Context
				span := opentracing.SpanFromContext(ctx)
				name := fmt.Sprintf("%s.%s", _KafkaConsumer, topic)
				if span == nil {
					//解析header到jaeger中
					headers := make(map[string]string)
					//kafka从0.11版本开始支持header
					if kb.v.IsAtLeast(sarama.V0_11_0_0) {
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

				//handler msg
				err = handler(ctx, msg.Value)
				if err != nil {
					ext.Error.Set(span, true)
					span.LogKV("error_msg", err.Error())
				}

				//不能通过defer的方式声明
				span.Finish()

				// mark message as processed
				consumer.MarkOffset(msg, "")
			}
		case <-signals:
			return
		}
	}
}

func GetConn() *kBroker {
	return &Kafka
}

func GetCConn() *kBroker {
	return &CKafka
}
