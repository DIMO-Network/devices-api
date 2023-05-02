package mock_services

import (
	"time"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/nats-io/nats.go"
)

func NewMockNATSService() *services.NATSService {
	n, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		return nil
	}

	js, err := n.JetStream()
	if err != nil {
		return nil
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:      "test-stream",
		Retention: nats.WorkQueuePolicy,
		Subjects:  []string{"test-subject"},
	})

	if err != nil {
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

	return natsSvc
}
