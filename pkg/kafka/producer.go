package kafka

import (
	"log"
	"strconv"
	"time"

	"github.com/vmware/service-level-indicator-exporter-for-kafka/pkg/metrics"

	"github.com/Shopify/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/config"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/pkg/common"
)

type KafkaProducer interface {
	Start()
	sendMessage(*sarama.ProducerMessage) error
}

type producer struct {
	Topic           string
	BootstrapServer string
	KafkaConfig     *sarama.Config
	KafkaClient     sarama.SyncProducer
}

var producerInstance *producer

func NewProducer(config config.KafkaConfig) (Producer KafkaProducer, err error) {

	producerInstance := new(producer)
	producerInstance.Topic = config.Topic
	producerInstance.BootstrapServer = config.BootstrapServer
	producerInstance.KafkaConfig = sarama.NewConfig()
	producerInstance.KafkaConfig.Producer.Partitioner = sarama.NewRandomPartitioner
	producerInstance.KafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	producerInstance.KafkaConfig.Producer.Return.Successes = true
	if config.ProducerConfig.MessagesSecond != 0 {
		flushFrequency := time.Duration(1000/config.ProducerConfig.MessagesSecond) * time.Millisecond
		producerInstance.KafkaConfig.Producer.Flush.Frequency = flushFrequency
		producerInstance.KafkaConfig.Producer.Flush.MaxMessages = 1
	}

	if config.Tls.Enabled == true {
		tlsConfig, err := common.NewTLSConfig(config.Tls.ClientCert,
			config.Tls.ClientKey,
			config.Tls.ClusterCert)
		if err != nil {
			log.Fatal(err)
		}
		// This can be used on test server if domain does not match cert:
		tlsConfig.InsecureSkipVerify = true
		producerInstance.KafkaConfig.Net.TLS.Enable = true
		producerInstance.KafkaConfig.Net.TLS.Config = tlsConfig
	}

	producerInstance.KafkaClient, err = sarama.NewSyncProducer([]string{config.BootstrapServer}, producerInstance.KafkaConfig)
	if err != nil {
		logrus.Error("Error creating the kafka sync producer " + err.Error())
	}
	return producerInstance, err
}

func (k *producer) Start() {
	for {
		message := &sarama.ProducerMessage{
			Topic:     k.Topic,
			Partition: -1,
			Value:     sarama.StringEncoder("example message"),
		}
		timer := prometheus.NewTimer(metrics.MessageSendDuration.WithLabelValues(k.BootstrapServer, k.Topic))
		err := k.sendMessage(message)
		if err != nil {
			metrics.ClusterUp.WithLabelValues(k.BootstrapServer).Set(0)
		} else {
			metrics.ClusterUp.WithLabelValues(k.BootstrapServer).Set(1)
		}
		timer.ObserveDuration()
	}
}

func (k *producer) sendMessage(msg *sarama.ProducerMessage) error {
	partition, offset, err := k.KafkaClient.SendMessage(msg)
	if err != nil {
		logrus.Error("Error sending message " + err.Error())
		metrics.ErrorTotalMessageSend.WithLabelValues(k.BootstrapServer, k.Topic).Inc()
		return err
	} else {
		metrics.TotalMessageSend.WithLabelValues(k.BootstrapServer, k.Topic).Inc()
		logrus.Info("Message was saved to partion: " + strconv.Itoa(int(partition)) + ". Message offset is: " + strconv.Itoa(int(offset)))
		return nil
	}
}
