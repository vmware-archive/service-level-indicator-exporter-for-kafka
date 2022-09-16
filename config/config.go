package config

import (
	"github.com/sirupsen/logrus"
)

type ConfigManager interface {
	LogConfig()
}

type PrometheusConfig struct {
	PrometheusServer string
	PrometheusPort   string
}

type KafkaConfig struct {
	BootstrapServer string
	Topic           string
	Tls             TlsConfig
}

type TlsConfig struct {
	Enabled     bool
	ClusterCert string
	ClientCert  string
	ClientKey   string
}

type Log struct {
	Level string
}
type Config struct {
	Prometheus PrometheusConfig
	Kafka      []KafkaConfig
	Log        Log
}

var Instance *Config

func (config *Config) LogConfig() {
	logrus.Infoln("Configuration loaded")
	logrus.Infoln(Instance)
}
