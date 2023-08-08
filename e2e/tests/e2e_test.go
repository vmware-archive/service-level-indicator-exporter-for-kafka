package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
	msConfig "github.com/vmware/service-level-indicator-exporter-for-kafka/config"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/pkg/kafka"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/pkg/metrics"

	"github.com/testcontainers/testcontainers-go/wait"
)

type testConfig struct {
	kafkaConfig                   msConfig.KafkaConfig
	expectErr                     bool
	expectedTotalMessageSend      float64
	expectedErrorTotalMessageSend float64
	expectedTotalMessageRead      float64
	expectedClusterUp             float64
}

func TestE2E(t *testing.T) {
	configs := []testConfig{
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: "localhost:9093",
				Topic:           "monitoring-topic",
			},
			expectErr:                     true,
			expectedTotalMessageSend:      0,
			expectedTotalMessageRead:      0,
			expectedErrorTotalMessageSend: 0,
			expectedClusterUp:             0,
		},
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: "localhost:9092",
				Topic:           "monitoring-topic",
				ConsumerConfig: msConfig.ConsumerConfig{
					FromBeginning: true,
				},
			},
			expectErr:                     false,
			expectedTotalMessageSend:      1,
			expectedTotalMessageRead:      1,
			expectedErrorTotalMessageSend: 0,
			expectedClusterUp:             1,
		},
	}
	kafka := startEnviron()
	defer destroyKafka(kafka)
	for _, config := range configs {
		testProducer(t, config)
		testConsumer(t, config)
	}
}

func startEnviron() *tc.LocalDockerCompose {
	//given
	kafka := tc.NewLocalDockerCompose(
		[]string{"../../compose.yaml"},
		"vdp-kafka-monitoring",
	)
	err := kafka.WithCommand([]string{"up", "-d"}).
		WaitForService("broker", wait.NewLogStrategy("started")).
		Invoke()
	if err.Error != nil {
		fmt.Print(err.Error)
		return nil
	} else {
		time.Sleep(10 * time.Second)
		return kafka
	}

}
func destroyKafka(compose *tc.LocalDockerCompose) {
	compose.Down()
	time.Sleep(3 * time.Second)
}

func testProducer(t *testing.T, config testConfig) {
	assert := assert.New(t)
	producer, err := kafka.NewProducer(config.kafkaConfig)
	if config.expectErr == true {
		assert.Error(err)
	} else {
		metrics.InitMetrics([]msConfig.KafkaConfig{config.kafkaConfig})
		message := &sarama.ProducerMessage{
			Topic:     config.kafkaConfig.Topic,
			Partition: -1,
			Value:     sarama.StringEncoder("example message"),
		}
		producer.SendMessage(message)
		assert.Equal(config.expectedTotalMessageSend, testutil.ToFloat64(metrics.TotalMessageSend.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
		assert.Equal(config.expectedErrorTotalMessageSend, testutil.ToFloat64(metrics.ErrorTotalMessageSend.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
	}
}

func testConsumer(t *testing.T, config testConfig) {
	assert := assert.New(t)

	consumer, err := kafka.NewConsumer(config.kafkaConfig)
	if config.expectErr == true {
		assert.Error(err)
	} else {
		go consumer.Start()
		time.Sleep(10 * time.Second)
		consumer.Stop()
		assert.Equal(config.expectedTotalMessageRead, testutil.ToFloat64(metrics.TotalMessageRead.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
	}

}
