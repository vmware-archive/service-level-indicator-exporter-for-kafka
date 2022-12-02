package cmd

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/config"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/pkg/kafka"
)

var appCmd = &cobra.Command{
	Use:     "app",
	Aliases: []string{"app"},
	Short:   "app",
	Run:     appComplete,
}

func init() {
	rootCmd.AddCommand(appCmd)
}

func appComplete(cmd *cobra.Command, args []string) {

	cfg := config.Instance
	logrus.Info("Starting producers/consumer.....")
	for _, cluster := range cfg.Kafka {

		kafkaClusterProducerMonitoring, err := kafka.NewProducer(cluster)
		if err != nil {
			logrus.Error("Error creating kafka producer: " + cluster.BootstrapServer)
		} else {
			go kafkaClusterProducerMonitoring.Start()
		}
		kafkaClusterConsumerMonitoring, err := kafka.NewConsumer(cluster)
		if err != nil {
			logrus.Error("Error creating kafka consumer: " + cluster.BootstrapServer)
		} else {
			go kafkaClusterConsumerMonitoring.Start()
		}

	}

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(cfg.Prometheus.PrometheusServer+":"+cfg.Prometheus.PrometheusPort, nil)
	if err != nil {
		logrus.Error("Error creating prometheus server: " + cfg.Prometheus.PrometheusServer + ":" + cfg.Prometheus.PrometheusPort + err.Error())
	}
}
