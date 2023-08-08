//go:build !ci
// +build !ci

package kafka

import (
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	msConfig "github.com/vmware/service-level-indicator-exporter-for-kafka/config"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/pkg/metrics"
)

type testConfig struct {
	kafkaConfig                   msConfig.KafkaConfig
	expectConnectedErr            bool
	expectSendErr                 bool
	expectedTotalMessageSend      float64
	expectedErrorTotalMessageSend float64
	expectedClusterUp             float64
	expectedKafkaResponse         sarama.KError
	numberMessagesSend            int
}

const TopicName = "monitoring-topic"

func TestProducerSuccess(t *testing.T) {
	assert := assert.New(t)
	seedBroker := sarama.NewMockBroker(t, 1)
	leader := sarama.NewMockBroker(t, 2)

	metadataResponse := new(sarama.MetadataResponse)
	metadataResponse.AddBroker(leader.Addr(), leader.BrokerID())
	metadataResponse.AddTopicPartition(TopicName, 0, leader.BrokerID(), nil, nil, nil, sarama.ErrNoError)

	configs := []testConfig{
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: "localhost:9093",
				Topic:           TopicName,
			},
			expectConnectedErr:            true,
			expectedTotalMessageSend:      0,
			expectedErrorTotalMessageSend: 0,
			expectedClusterUp:             0,
		},
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: seedBroker.Addr(),
				Topic:           TopicName,
				ProducerConfig: msConfig.ProducerConfig{
					MinVersion: true,
				},
			},
			expectConnectedErr:            false,
			expectSendErr:                 false,
			expectedTotalMessageSend:      1,
			expectedErrorTotalMessageSend: 0,
			expectedClusterUp:             1,
			expectedKafkaResponse:         sarama.ErrNoError,
		},
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: seedBroker.Addr(),
				Topic:           TopicName,
				ProducerConfig: msConfig.ProducerConfig{
					MinVersion: true,
				},
			},
			expectConnectedErr:            false,
			expectSendErr:                 true,
			expectedTotalMessageSend:      0,
			expectedErrorTotalMessageSend: 1,
			expectedClusterUp:             1,
			expectedKafkaResponse:         sarama.ErrPreferredLeaderNotAvailable,
		},
	}

	for _, config := range configs {
		seedBroker.Returns(metadataResponse)
		producer, err := NewProducer(config.kafkaConfig)
		if config.expectConnectedErr == true {
			assert.Error(err)
		} else {
			assert.NoError(err)
			prodSuccess := new(sarama.ProduceResponse)
			prodSuccess.AddTopicPartition(config.kafkaConfig.Topic, 0, config.expectedKafkaResponse)
			leader.Returns(prodSuccess)

			metrics.InitMetrics([]msConfig.KafkaConfig{config.kafkaConfig})
			message := &sarama.ProducerMessage{
				Topic:     config.kafkaConfig.Topic,
				Partition: -1,
				Value:     sarama.StringEncoder("example message"),
			}
			err = producer.SendMessage(message)
			if config.expectSendErr == true {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(config.expectedTotalMessageSend, testutil.ToFloat64(metrics.TotalMessageSend.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
			assert.Equal(config.expectedErrorTotalMessageSend, testutil.ToFloat64(metrics.ErrorTotalMessageSend.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
			metrics.ResetMetrics([]msConfig.KafkaConfig{config.kafkaConfig})
		}
	}
}

func TestProducerLoopSuccess(t *testing.T) {
	assert := assert.New(t)
	seedBroker := sarama.NewMockBroker(t, 1)
	leader := sarama.NewMockBroker(t, 2)

	metadataResponse := new(sarama.MetadataResponse)
	metadataResponse.AddBroker(leader.Addr(), leader.BrokerID())
	metadataResponse.AddTopicPartition(TopicName, 0, leader.BrokerID(), nil, nil, nil, sarama.ErrNoError)

	configs := []testConfig{
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: seedBroker.Addr(),
				Topic:           TopicName,
				ProducerConfig: msConfig.ProducerConfig{
					MinVersion: true,
				},
			},
			numberMessagesSend:            10,
			expectedTotalMessageSend:      10,
			expectedErrorTotalMessageSend: 0,
			expectedClusterUp:             1,
			expectedKafkaResponse:         sarama.ErrNoError,
		},
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: seedBroker.Addr(),
				Topic:           TopicName,
				ProducerConfig: msConfig.ProducerConfig{
					MinVersion: true,
				},
			},
			numberMessagesSend:            10,
			expectedTotalMessageSend:      0,
			expectedErrorTotalMessageSend: 10,
			expectedClusterUp:             0,
			expectedKafkaResponse:         sarama.ErrPreferredLeaderNotAvailable,
		},
	}

	for _, config := range configs {
		seedBroker.Returns(metadataResponse)
		producer, err := NewProducer(config.kafkaConfig)
		if config.expectConnectedErr == true {
			assert.Error(err)
		} else {
			assert.NoError(err)
			prodSuccess := new(sarama.ProduceResponse)

			go func() {
				for i := 0; i < config.numberMessagesSend; i++ {
					prodSuccess.AddTopicPartition(config.kafkaConfig.Topic, 0, config.expectedKafkaResponse)
					leader.Returns(prodSuccess)
				}
			}()

			metrics.InitMetrics([]msConfig.KafkaConfig{config.kafkaConfig})
			go producer.Start()
			time.Sleep(5 * time.Second)
			go producer.Stop()
			assert.Equal(config.expectedClusterUp, testutil.ToFloat64(metrics.ClusterUp.WithLabelValues(config.kafkaConfig.BootstrapServer)))
			assert.Equal(config.expectedTotalMessageSend, testutil.ToFloat64(metrics.TotalMessageSend.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
			assert.Equal(config.expectedErrorTotalMessageSend, testutil.ToFloat64(metrics.ErrorTotalMessageSend.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
			metrics.ResetMetrics([]msConfig.KafkaConfig{config.kafkaConfig})
		}
	}
}
