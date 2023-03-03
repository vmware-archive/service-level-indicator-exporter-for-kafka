package cmd

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/config"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/pkg/kafka"
)

var producerCmd = &cobra.Command{
	Use:     "producer",
	Aliases: []string{"prod"},
	Short:   "producer",
	Run:     startProducer,
}

func init() {
	rootCmd.AddCommand(producerCmd)
}

// startProducer create a new kafka producer an prometheus server with metrics
func startProducer(cmd *cobra.Command, args []string) {

	cfg := config.Instance
	logrus.Info("Starting producers.....")
	for _, cluster := range cfg.Kafka {

		kafkaClusterMonitoring, err := kafka.NewProducer(cluster)
		if err != nil {
			logrus.Error("Error creating kafka producer: " + cluster.BootstrapServer)
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
