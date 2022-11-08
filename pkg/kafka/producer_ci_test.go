//go:build ci
// +build ci

package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	msConfig "gitlab.eng.vmware.com/vdp/vdp-kafka-monitoring/config"
	"gitlab.eng.vmware.com/vdp/vdp-kafka-monitoring/pkg/metrics"
	"testing"
)

func TestCIProducer(t *testing.T) {
	assert := assert.New(t)
	configs := []testConfig{
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: "localhost:9093",
				Topic:           "monitoring-topic",
			},
			expectErr:                     true,
			expectedTotalMessageSend:      0,
			expectedErrorTotalMessageSend: 0,
			expectedClusterUp:             0,
		},
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: "localhost:9092",
				Topic:           "monitoring-topic",
			},
			expectErr:                     false,
			expectedTotalMessageSend:      1,
			expectedErrorTotalMessageSend: 0,
			expectedClusterUp:             1,
		},
	}
	for _, config := range configs {
		producer, err := NewProducer(config.kafkaConfig)
		if config.expectErr == true {
			assert.Error(err)
		} else {
			metrics.InitMetrics([]msConfig.KafkaConfig{config.kafkaConfig})
			message := &sarama.ProducerMessage{
				Topic:     config.kafkaConfig.Topic,
				Partition: -1,
				Value:     sarama.StringEncoder("example message"),
			}
			producer.sendMessage(message)
			assert.Equal(config.expectedTotalMessageSend, testutil.ToFloat64(metrics.TotalMessageSend.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
			assert.Equal(config.expectedErrorTotalMessageSend, testutil.ToFloat64(metrics.ErrorTotalMessageSend.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
		}
	}
}
