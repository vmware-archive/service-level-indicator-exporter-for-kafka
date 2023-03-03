package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/config"
)

//TotalMessageSend Producer instance will increase counter with total messages send per kafka cluster
var TotalMessageSend = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_monitoring_total_messages_send",
		Help: "Number of messages send to kafka",
	},
	[]string{"cluster", "topic"},
)

//ErrorTotalMessageSend Producer instance will increase counter if we are not able of send a message per kafka cluster
var ErrorTotalMessageSend = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_monitoring_error_total_messages_send",
		Help: "Number of messages send with failure to kafka",
	},
	[]string{"cluster", "topic"},
)

//ClusterUp Producer will set up Gauge values to 0 if cluster is unreacheable or 1 if we are able to connect to kafka cluster
var ClusterUp = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "kafka_monitoring_cluster_up",
		Help: "Kafka clusters with errors",
	},
	[]string{"cluster"},
)

//MessageSendDuration Producer summary with rate duration/reqs send
var MessageSendDuration = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name: "kafka_monitoring_message_send_duration",
		Help: "Duration of kafka monitoring connection",
	},
	[]string{"cluster", "topic"},
)

//TotalMessageRead Consumer instance will increase counter with total messages read per kafka cluster
var TotalMessageRead = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_monitoring_total_messages_read",
		Help: "Number of messages read for kafka consumer",
	},
	[]string{"cluster", "topic"},
)

//TotalMessageRead Consumer instance will increase counter if it is unable of read from kafka cluster
var ErrorInRead = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_monitoring_error_in_read",
		Help: "Errors in kafka consumer",
	},
	[]string{"cluster", "topic"},
)

//InitMetrics function call when app start for register and init the metrics
func InitMetrics(cfg []config.KafkaConfig) {
	for _, cluster := range cfg {
		prometheus.Register(TotalMessageSend.WithLabelValues(cluster.BootstrapServer, cluster.Topic))
		prometheus.Register(TotalMessageRead.WithLabelValues(cluster.BootstrapServer, cluster.Topic))
		prometheus.Register(ClusterUp.WithLabelValues(cluster.BootstrapServer))
		prometheus.Register(ErrorTotalMessageSend.WithLabelValues(cluster.BootstrapServer, cluster.Topic))
		prometheus.Register(ErrorInRead.WithLabelValues(cluster.BootstrapServer, cluster.Topic))
		prometheus.Register(MessageSendDuration)
	}
}
