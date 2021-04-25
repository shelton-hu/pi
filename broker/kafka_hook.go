package broker

import (
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

type KafkaHook struct {
	k         kBroker
	topic     string
	formatter logrus.Formatter
}

// Create a new KafkaHook.
func NewKafkaHook(topic string, b kBroker) *KafkaHook {
	hook := &KafkaHook{
		k:         b,
		formatter: &logrus.JSONFormatter{},
		topic:     topic,
	}
	return hook
}

func (hook *KafkaHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.Level(0),
		logrus.Level(1),
		logrus.Level(2),
		logrus.Level(3),
		logrus.Level(4),
		logrus.Level(5),
		logrus.Level(6),
	}
}

func (hook *KafkaHook) Fire(entry *logrus.Entry) (err error) {
	msgB, err := hook.formatter.Format(entry)
	//组装message
	msg := &sarama.ProducerMessage{}
	msg.Topic = hook.topic
	msg.Value = sarama.ByteEncoder(msgB)
	go func() {
		hook.k.p.Input() <- msg
		select {
		case <-hook.k.p.Successes():
			return
		case <-hook.k.p.Errors():
			return
		}
	}()
	return
}
