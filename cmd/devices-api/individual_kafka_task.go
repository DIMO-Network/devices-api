package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/google/subcommands"
	"github.com/rs/zerolog"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/Shopify/sarama"
	"github.com/segmentio/ksuid"
)

type stopTaskByKeyCmd struct {
	logger    zerolog.Logger
	settings  config.Settings
	producer  sarama.SyncProducer
	container dependencyContainer
}

func (*stopTaskByKeyCmd) Name() string     { return "stop-task-by-key" }
func (*stopTaskByKeyCmd) Synopsis() string { return "stop-task-by-key args to stdout." }
func (*stopTaskByKeyCmd) Usage() string {
	return `stop-task-by-key:
	stop-task-by-key args.
  `
}

func (p *stopTaskByKeyCmd) SetFlags(f *flag.FlagSet) {

}

func (p *stopTaskByKeyCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	p.producer = p.container.getKafkaProducer()

	if len(os.Args[1:]) != 2 {
		p.logger.Fatal().Msgf("Expected an argument, the task key.")
	}
	taskKey := os.Args[2]
	p.logger.Info().Msgf("Stopping task %s", taskKey)
	err := stopTaskByKey(&p.settings, taskKey, p.producer)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error stopping task.")
	}
	return subcommands.ExitSuccess
}

func stopTaskByKey(settings *config.Settings, taskKey string, producer sarama.SyncProducer) error {
	tt := struct {
		services.CloudEventHeaders
		Data interface{} `json:"data"`
	}{
		CloudEventHeaders: services.CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/FAKE",
			SpecVersion: "1.0",
			Subject:     "FAKE",
			Time:        time.Now(),
			Type:        "zone.dimo.task.tesla.poll.stop",
		},
		Data: struct {
			TaskID        string `json:"taskId"`
			UserDeviceID  string `json:"userDeviceId"`
			IntegrationID string `json:"integrationId"`
		}{
			TaskID:        taskKey,
			UserDeviceID:  "FAKE",
			IntegrationID: "FAKE",
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	_, _, err = producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: settings.TaskStopTopic,
			Key:   sarama.StringEncoder(taskKey),
			Value: sarama.ByteEncoder(ttb),
		},
	)

	return err
}
