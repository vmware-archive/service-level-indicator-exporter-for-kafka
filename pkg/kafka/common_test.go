package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	msConfig "gitlab.eng.vmware.com/vdp/vdp-kafka-monitoring/config"
)

type testConfig struct {
	kafkaConfig                   msConfig.KafkaConfig
	expectErr                     bool
	expectedTotalMessageSend      float64
	expectedErrorTotalMessageSend float64
	expectedClusterUp             float64
}

func startEnviron() *testcontainers.LocalDockerCompose {
	//given
	kafka := testcontainers.NewLocalDockerCompose(
		[]string{"../../compose.yaml"},
		strings.ToLower(uuid.New().String()),
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
func destroyKafka(compose *testcontainers.LocalDockerCompose) {
	compose.Down()
	time.Sleep(3 * time.Second)
}
