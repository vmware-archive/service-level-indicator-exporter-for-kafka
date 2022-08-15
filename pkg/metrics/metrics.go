package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.eng.vmware.com/vdp/vdp-kafka-monitoring/config"
)

var TotalMessageSend = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_monitoring_total_messages_send",
		Help: "Number of messages send to kafka",
	},
	[]string{"cluster", "topic"},
)

var ErrorTotalMessageSend = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_monitoring_error_total_messages_send",
		Help: "Number of messages send with failure to kafka",
	},
	[]string{"cluster", "topic"},
)

var ClusterUp = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "kafka_monitoring_cluster_up",
		Help: "Kafka clusters with errors",
	},
	[]string{"cluster"},
)

var MessageSendDuration = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name: "kafka_monitoring_message_send_duration",
		Help: "Duration of kafka monitoring connection",
	},
	[]string{"cluster", "topic"},
)

var TotalMessageRead = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_monitoring_total_messages_read",
		Help: "Number of messages read for kafka consumer",
	},
	[]string{"cluster", "topic"},
)

func InitMetrics(cfg *config.Config) {
	for _, cluster := range cfg.Kafka {
		prometheus.Register(TotalMessageSend.WithLabelValues(cluster.BootstrapServer, cluster.Topic))
		prometheus.Register(ClusterUp.WithLabelValues(cluster.BootstrapServer))
		prometheus.Register(ErrorTotalMessageSend.WithLabelValues(cluster.BootstrapServer, cluster.Topic))
		prometheus.Register(MessageSendDuration)
	}
}
