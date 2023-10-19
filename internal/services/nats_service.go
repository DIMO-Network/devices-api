package services

import (
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

type NATSService struct {
	log              *zerolog.Logger
	JetStream        nats.JetStreamContext
	JetStreamName    string
	JetStreamSubject string
	AckTimeout       time.Duration
	DurableConsumer  string
}

func NewNATSService(settings *config.Settings, log *zerolog.Logger) (*NATSService, error) {
	n, err := nats.Connect(settings.NATSURL)
	if err != nil {
		return nil, err
	}

	js, err := n.JetStream()
	if err != nil {
		return nil, err
	}

	to, err := time.ParseDuration(settings.NATSAckTimeout)
	if err != nil {
		return nil, err
	}

	natsSvc := &NATSService{
		log:              log,
		JetStream:        js,
		JetStreamName:    settings.NATSStreamName,
		JetStreamSubject: settings.NATSValuationSubject,
		AckTimeout:       to,
		DurableConsumer:  settings.NATSDurableConsumer}

	return natsSvc, nil
}
