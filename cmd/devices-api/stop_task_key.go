package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/segmentio/ksuid"
)

type stopTaskByKeyCmd struct {
	logger    zerolog.Logger
	settings  config.Settings
	producer  sarama.SyncProducer
	container dependencyContainer
	pdb       db.Store
}

func (*stopTaskByKeyCmd) Name() string     { return "stop-task-by-key" }
func (*stopTaskByKeyCmd) Synopsis() string { return "stop-task-by-key args to stdout." }
func (*stopTaskByKeyCmd) Usage() string {
	return `stop-task-by-key:
	stop-task-by-key args.
  `
}

// nolint
func (p *stopTaskByKeyCmd) SetFlags(f *flag.FlagSet) {

}

func (p *stopTaskByKeyCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	p.producer = p.container.getKafkaProducer()

	if len(os.Args[1:]) != 2 {
		p.logger.Fatal().Msgf("Expected an argument, the task key.")
	}
	taskKey := os.Args[2]
	p.logger.Info().Msgf("Stopping task %s", taskKey)
	err := p.stopTaskByKey(&p.settings, taskKey, p.producer)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error stopping task.")
	}
	return subcommands.ExitSuccess
}

func (p *stopTaskByKeyCmd) stopTaskByKey(settings *config.Settings, taskKey string, producer sarama.SyncProducer) error {
	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.TaskID.EQ(null.StringFrom(taskKey)),
	).One(context.TODO(), p.container.dbs().Reader)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	} else {
		udai.Status = models.UserDeviceAPIIntegrationStatusAuthenticationFailure
		_, err := udai.Update(context.TODO(), p.container.dbs().Writer, boil.Whitelist(models.UserDeviceAPIIntegrationColumns.Status, models.UserDeviceAPIIntegrationColumns.UpdatedAt))
		if err != nil {
			return err
		}
	}

	tt := shared.CloudEvent[any]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/FAKE",
		SpecVersion: "1.0",
		Subject:     "FAKE",
		Time:        time.Now(),
		Type:        "zone.dimo.task.tesla.poll.stop",
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

	err = producer.SendMessages(
		[]*sarama.ProducerMessage{
			{
				Topic: settings.TaskStopTopic,
				Key:   sarama.StringEncoder(taskKey),
				Value: sarama.ByteEncoder(ttb),
			},
			{
				Topic: settings.TaskCredentialTopic,
				Key:   sarama.StringEncoder(taskKey),
				Value: nil,
			},
		},
	)

	return err
}
