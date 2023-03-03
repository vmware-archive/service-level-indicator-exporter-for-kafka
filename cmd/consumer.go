package cmd

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/config"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/pkg/kafka"
)

var consumerCmd = &cobra.Command{
	Use:     "consumer",
	Aliases: []string{"cons"},
	Short:   "consumer",
	Run:     startConsumer,
}

func init() {
	rootCmd.AddCommand(consumerCmd)
}

// startConsumer create a new consumer an prometheus server with metrics
func startConsumer(cmd *cobra.Command, args []string) {

	cfg := config.Instance
	logrus.Info("Starting consumer.....")
	for _, cluster := range cfg.Kafka {

		kafkaClusterMonitoring, err := kafka.NewConsumer(cluster)
		if err != nil {
			logrus.Error("Error creating kafka consumer: " + cluster.BootstrapServer)
		} else {
			go kafkaClusterMonitoring.Start()
		}

	}
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(cfg.Prometheus.PrometheusServer+":"+cfg.Prometheus.PrometheusPort, nil)
	if err != nil {
		logrus.Error("Error creating prometheus server: " + cfg.Prometheus.PrometheusServer + ":" + cfg.Prometheus.PrometheusPort + err.Error())
	}
}
