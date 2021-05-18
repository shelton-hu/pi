package kafka

import (
	"context"
	"sync"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"

	"github.com/shelton-hu/logger"

	"github.com/shelton-hu/pi/config"
)

var kafka *Kafka

// Kafka ...
type Kafka struct {
	addrs []string
	cfg   *sarama.Config
	c     sarama.Client
	p     sarama.AsyncProducer
	cs    []sarama.Client
	clcfg *cluster.Config

	mu sync.Mutex
}

// ConnectKafka ...
func ConnectKafka(ctx context.Context, kafkaConfig config.Kafka) {
	// kafka.addrs
	kafka.addrs = kafkaConfig.Addrs
	// kafka.cfg
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	var err error
	config.Version, err = sarama.ParseKafkaVersion(kafkaConfig.Version)
	if err != nil {
		panic(err)
	}
	kafka.cfg = config
	// kafka.c
	c, err := sarama.NewClient(kafka.addrs, kafka.cfg)
	if err != nil {
		panic(err)
	}
	kafka.c = c
	// kafka.p
	p, err := sarama.NewAsyncProducerFromClient(c)
	if err != nil {
		panic(err)
	}
	kafka.p = p
	// kafka.cs
	kafka.mu.Lock()
	defer kafka.mu.Unlock()
	kafka.cs = make([]sarama.Client, 0)
	// kafka.clcfg
	kafka.clcfg = cluster.NewConfig()
	kafka.clcfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	kafka.clcfg.Group.Return.Notifications = true
	kafka.clcfg.Version = kafka.cfg.Version
}

// CloseKafka ...
func CloseKafka(ctx context.Context) {
	for _, client := range kafka.cs {
		if err := client.Close(); err != nil {
			logger.Error(ctx, err.Error())
		}
	}
	kafka.cs = nil
	if err := kafka.p.Close(); err != nil {
		logger.Error(ctx, err.Error())
	}
	kafka = nil
}

// GetConnect returns the kafka instance.
func GetConnect(ctx context.Context) *Kafka {
	return kafka
}
