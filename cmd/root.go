package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.eng.vmware.com/vdp/vdp-kafka-monitoring/config"
	"gitlab.eng.vmware.com/vdp/vdp-kafka-monitoring/pkg/metrics"
)

var version = "0.0.1"
var rootCmd = &cobra.Command{
	Use:     "vdp-kafka-monitoring",
	Version: version,
	Short:   "vdp-kafka-monitoring - a simple CLI to monitor kafka clusters",
	Long: `vdp-kafka-monitoring is a CLI to monitoring kafka cluster. Adding the config we will send
	a notification for every kafka cluster`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
var cfgFile string
var userLicense string

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.yaml", "config file (default is config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "name of license for the project")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	viper.SetDefault("license", "vmware")

	cobra.OnInitialize(initCommand)
}

func initCommand() {

	createConfig()
	// Only log the warning severity or above.
	logrus.SetFormatter(&logrus.JSONFormatter{})
	// LOG_LEVEL not set, let's default to info
	lvl := config.Instance.Log.Level
	if len(lvl) == 0 {
		lvl = "info"
	}
	// parse string, this is built-in feature of logrus
	ll, err := logrus.ParseLevel(lvl)
	if err != nil {
		ll = logrus.DebugLevel
	}
	// set global log level
	logrus.SetLevel(ll)
	metrics.InitMetrics(config.Instance)
}
func createConfig() {

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logrus.Info("Using config file: " + viper.ConfigFileUsed())
	}

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatal("fatal error config file: " + err.Error())
	}
	config.Instance = new(config.Config)
	err = viper.Unmarshal(&config.Instance)
	if err != nil {
		logrus.Fatal("fatal error unmashal config file: " + err.Error())
	}
}
