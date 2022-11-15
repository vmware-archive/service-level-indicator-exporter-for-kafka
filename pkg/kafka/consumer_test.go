package kafka

import (
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	msConfig "github.com/vmware/service-level-indicator-exporter-for-kafka/config"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/pkg/metrics"
)

func testConsumer(t *testing.T) {
	assert := assert.New(t)
	configs := []struct {
		kafkaConfig              msConfig.KafkaConfig
		expectErr                bool
		expectedTotalMessageRead float64
		expectedClusterUp        float64
	}{
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: "localhost2:9093",
				Topic:           "monitoring-topic",
			},
			expectErr:                true,
			expectedTotalMessageRead: 0,
			expectedClusterUp:        0,
		},
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: "localhost:9092",
				Topic:           "monitoring-topic",
			},
			expectErr:                false,
			expectedTotalMessageRead: 1,
			expectedClusterUp:        1,
		},
	}
	for _, config := range configs {
		consumer, err := NewConsumer(config.kafkaConfig)
		if config.expectErr == true {
			assert.Error(err)
		} else {
			go consumer.Start()
			time.Sleep(10 * time.Second)

			producer, err := NewProducer(config.kafkaConfig)
			metrics.InitMetrics([]msConfig.KafkaConfig{config.kafkaConfig})
			message := &sarama.ProducerMessage{
				Topic:     config.kafkaConfig.Topic,
				Partition: -1,
				Value:     sarama.StringEncoder("example message"),
			}
			producer.sendMessage(message)
			if err != nil {
				fmt.Println("Unable to create producer for consumer test")
			}

			time.Sleep(3 * time.Second)
			consumer.Stop()
			assert.Equal(config.expectedTotalMessageRead, testutil.ToFloat64(metrics.TotalMessageRead.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
		}

	}
}
