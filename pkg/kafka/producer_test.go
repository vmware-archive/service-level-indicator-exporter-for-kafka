package kafka

import (
	"strings"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	testcontainers "github.com/testcontainers/testcontainers-go"
	msConfig "gitlab.eng.vmware.com/vdp/vdp-kafka-monitoring/config"
	"gitlab.eng.vmware.com/vdp/vdp-kafka-monitoring/pkg/metrics"
)

func startEnviron() *testcontainers.LocalDockerCompose {
	//given
	kafka := testcontainers.NewLocalDockerCompose(
		[]string{"../../compose.yaml"},
		strings.ToLower(uuid.New().String()),
	)
	kafka.WithCommand([]string{"up", "-d"}).Invoke()
	time.Sleep(15 * time.Second)
	return kafka
}
func destroyKafka(compose *testcontainers.LocalDockerCompose) {
	compose.Down()
	time.Sleep(1 * time.Second)
}
func TestProducer(t *testing.T) {
	kafka := startEnviron()
	assert := assert.New(t)
	configs := []struct {
		kafkaConfig                   msConfig.KafkaConfig
		expectErr                     bool
		expectedTotalMessageSend      float64
		expectedErrorTotalMessageSend float64
	}{
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: "localhost2:9093",
				Topic:           "monitoring-topic",
			},
			expectErr:                     true,
			expectedTotalMessageSend:      0,
			expectedErrorTotalMessageSend: 0,
		},
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: "localhost:9092",
				Topic:           "monitoring-topic",
			},
			expectErr:                     false,
			expectedTotalMessageSend:      1,
			expectedErrorTotalMessageSend: 0,
		},
	}
	for _, config := range configs {
		producer, err := NewProducer(config.kafkaConfig)
		if config.expectErr == true {
			assert.Error(err)
		} else {
			message := &sarama.ProducerMessage{
				Topic:     config.kafkaConfig.Topic,
				Partition: -1,
				Value:     sarama.StringEncoder("example message"),
			}
			err = producer.sendMessage(message)
			assert.ErrorIs(err, nil)
		}

		assert.Equal(config.expectedTotalMessageSend, testutil.ToFloat64(metrics.TotalMessageSend.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
		assert.Equal(config.expectedErrorTotalMessageSend, testutil.ToFloat64(metrics.ErrorTotalMessageSend.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))

	}
	destroyKafka(kafka)
}
