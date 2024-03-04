package mock_services //nolint:all

import (
	"time"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

const Timeout = 2 * time.Second

func RunServer() *server.Server {
	s := server.New(&server.Options{
		Host:           "127.0.0.1",
		Port:           server.RANDOM_PORT,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 2048,
	})

	go server.Run(s) // nolint:errcheck

	if !s.ReadyForConnections(10 * time.Second) {
		panic("nats server not ready for connections")
	}

	return s
}

func waitConnected(nc *nats.Conn) {
	timeout := time.Now().Add(Timeout)
	for time.Now().Before(timeout) {
		if nc.IsConnected() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	panic("nats server not connected")
}

func NewMockNATSService(streamName string) (*services.NATSService, *server.Server, error) {

	s := RunServer()

	err := s.EnableJetStream(&server.JetStreamConfig{})

	if err != nil {
		return nil, s, err
	}
	nc, err := nats.Connect("nats://" + s.Addr().String())
	if err != nil {
		return nil, s, err
	}

	waitConnected(nc)

	js, err := nc.JetStream()
	if err != nil {
		return nil, s, err
	}

	if _, err = js.AddStream(&nats.StreamConfig{
		Name:      streamName,
		Retention: nats.WorkQueuePolicy,
		Subjects:  []string{"test-subject"},
	}); err != nil {
		return nil, s, err
	}

	to, err := time.ParseDuration("2s")
	if err != nil {
		return nil, s, err
	}

	natsSvc := &services.NATSService{
		JetStream:        js,
		JetStreamName:    streamName,
		ValuationSubject: "test-subject",
		AckTimeout:       to,
		DurableConsumer:  "test-durable-consumer",
	}
	return natsSvc, s, nil
}
