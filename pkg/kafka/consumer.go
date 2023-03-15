package kafka

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/config"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/pkg/common"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/pkg/metrics"
)

type KafkaConsumer interface {
	Start()
	Stop()
	//sendMessage(*sarama.ProducerMessage) error
}

type consumer struct {
	ready           chan bool
	Topic           string
	BootstrapServer string
	KafkaConfig     *sarama.Config
	KafkaClient     sarama.ConsumerGroup
	Ctx             context.Context
	Cancel          context.CancelFunc
}

var consumerInstance *consumer

func NewConsumer(config config.KafkaConfig) (Consumer KafkaConsumer, err error) {

	consumerInstance := new(consumer)
	consumerInstance.ready = make(chan bool)
	consumerInstance.Topic = config.Topic
	consumerInstance.BootstrapServer = config.BootstrapServer
	consumerInstance.KafkaConfig = sarama.NewConfig()
	consumerInstance.KafkaConfig.Consumer.Return.Errors = true
	if config.ConsumerConfig.FromBeginning {
		consumerInstance.KafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
	consumerInstance.Ctx, consumerInstance.Cancel = context.WithCancel(context.Background())
	if config.Tls.Enabled {
		tlsConfig, err := common.NewTLSConfig(config.Tls.ClientCert,
			config.Tls.ClientKey,
			config.Tls.ClusterCert)
		if err != nil {
			log.Fatal(err)
		}
		// This can be used on test server if domain does not match cert:
		tlsConfig.InsecureSkipVerify = true
		consumerInstance.KafkaConfig.Net.TLS.Enable = true
		consumerInstance.KafkaConfig.Net.TLS.Config = tlsConfig
	}

	consumerInstance.KafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	consumerInstance.KafkaClient, err = sarama.NewConsumerGroup([]string{config.BootstrapServer}, "vdp-kafka-monitoring", consumerInstance.KafkaConfig)
	if err != nil {
		logrus.Error("Error creating the kafka consumer " + err.Error())
	}

	return consumerInstance, err
}

func (k *consumer) Start() {
	keepRunning := true
	consumptionIsPaused := false

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := k.KafkaClient.Consume(k.Ctx, strings.Split(k.Topic, ","), k); err != nil {
				logrus.Error("Error from consumer: " + err.Error())
				metrics.ErrorInRead.WithLabelValues(k.BootstrapServer, k.Topic).Inc()
				return
			}
			// check if context was cancelled, signaling that the consumer should stop
			if k.Ctx.Err() != nil {
				logrus.Error("Closing consumer because of context: " + k.Ctx.Err().Error())
				return
			}
			k.ready = make(chan bool)
		}
	}()

	<-k.ready // Await till the consumer has been set up
	logrus.Info("Starting sarama consumer group...")

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	for keepRunning {
		select {
		case <-k.Ctx.Done():
			logrus.Info("terminating: context cancelled")
			keepRunning = false
		case <-sigusr1:
			toggleConsumptionFlow(k.KafkaClient, &consumptionIsPaused)
		}
	}
	k.Cancel()
	wg.Wait()
	if err := k.KafkaClient.Close(); err != nil {
		logrus.Error("Error closing client: ", err)
	}
}

func toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		logrus.Info("Resuming consumption")
		client.ResumeAll()
	} else {
		logrus.Info("Pausing consumption")
		client.PauseAll()
	}

	*isPaused = !*isPaused
}

func (k *consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(k.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (k *consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (k *consumer) Stop() {
	k.Cancel()
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (k *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			logrus.Info("Message claimed: offset: " + strconv.Itoa(int(message.Offset)) + " for partition: " + strconv.Itoa(int(message.Partition)) + " in topic: " + message.Topic)
			session.MarkMessage(message, "")
			metrics.TotalMessageRead.WithLabelValues(k.BootstrapServer, k.Topic).Inc()
		case <-session.Context().Done():
			return nil
		}
	}
}
