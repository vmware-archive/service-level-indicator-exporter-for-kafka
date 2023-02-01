package kafka

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"

	"github.com/testcontainers/testcontainers-go/wait"
)

func TestE2E(t *testing.T) {
	kafka := startEnviron()
	defer destroyKafka(kafka)
	testProducer(t)
	testConsumer(t)
}

func startEnviron() *tc.LocalDockerCompose {
	//given
	kafka := tc.NewLocalDockerCompose(
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
func destroyKafka(compose *tc.LocalDockerCompose) {
	compose.Down()
	time.Sleep(3 * time.Second)
}
