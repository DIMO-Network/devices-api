package mock_services //nolint:all

import (
	"fmt"
	"time"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

const testPort = 8369

func RunServerOnPort(port int) *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.Port = port
	return RunServerWithOptions(&opts)
}

func RunServerWithOptions(opts *server.Options) *server.Server {
	return natsserver.RunServer(opts)
}

func NewMockNATSService() *services.NATSService {

	s := RunServerOnPort(testPort)
	defer s.Shutdown()

	_ = s.EnableJetStream(&server.JetStreamConfig{})

	time.Sleep(time.Second * 5)

	sURL := fmt.Sprintf("nats://127.0.0.1:%d", testPort)

	if nc, err := nats.Connect(sURL); err != nil {
		panic(err)
	} else {
		js, err := nc.JetStream()
		if err != nil {
			return nil
		}
		if _, err = js.AddStream(&nats.StreamConfig{
			Name:      "test-stream",
			Retention: nats.WorkQueuePolicy,
			Subjects:  []string{"test-subject"},
		}); err != nil {
			return nil
		}
		to, err := time.ParseDuration("5s")
		if err != nil {
			return nil
		}

		natsSvc := &services.NATSService{
			JetStream:        js,
			JetStreamName:    "test-stream",
			JetStreamSubject: "test-subject",
			AckTimeout:       to,
			DurableConsumer:  "test-durable-consumer",
		}

		defer s.Shutdown()
		return natsSvc
	}
}
