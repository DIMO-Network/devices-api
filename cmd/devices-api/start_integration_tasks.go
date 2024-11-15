package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/goccy/go-json"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
)

type startIntegrationTask struct {
	logger    zerolog.Logger
	container dependencyContainer
	settings  config.Settings
	pdb       db.Store

	producer sarama.SyncProducer
}

func (*startIntegrationTask) Name() string { return "start-integration-tasks" }
func (*startIntegrationTask) Synopsis() string {
	return "start-integration-tasks args to stdout."
}
func (*startIntegrationTask) Usage() string {
	return `start-integration-tasks:
	start-integration-tasks args.
  `
}

var intMap = map[string]string{
	"smartcar": "22N2xaPOq2WW2gAHBHd0Ikn4Zob",
	"tesla":    "26A5Dk3vvvQutjSyF0Jka2DP5lg",
}

// nolint
func (p *startIntegrationTask) SetFlags(f *flag.FlagSet) {

}

func (p *startIntegrationTask) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.producer = p.container.getKafkaProducer()
	err := p.startIntegrationTaskGo()
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running SD task start.")
	}
	return subcommands.ExitSuccess
}

func (p *startIntegrationTask) startIntegrationTaskGo() error {
	ctx := context.Background()
	if len(os.Args[2:]) != 1 {
		p.logger.Fatal().Msg("Expected an argument, the integration name.")
	}

	intName := os.Args[2]

	intID, ok := intMap[intName]
	if !ok {
		p.logger.Fatal().Msg("Not a recognized integration name.")
	}

	udais, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(intID),
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
		models.UserDeviceAPIIntegrationWhere.TaskID.IsNotNull(),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice, models.UserDeviceRels.VehicleTokenSyntheticDevice)),
	).All(ctx, p.container.dbs().Reader)
	if err != nil {
		return err
	}

	// Unfortunately we do not want to overwrite the credentials KTable, so we avoid using
	// stuff from smartcar_task_service.go and so on.

	for _, udai := range udais {
		ud := udai.R.UserDevice
		if ud.TokenID.IsZero() || ud.R.VehicleTokenSyntheticDevice == nil {
			continue
		}

		if intName == "smartcar" {
			var md services.UserDeviceAPIIntegrationsMetadata
			if err := udai.Metadata.Unmarshal(&md); err != nil {
				return err
			}

			e := shared.CloudEvent[services.SmartcarTask]{
				ID:          ksuid.New().String(),
				Source:      "dimo/integration/" + udai.IntegrationID,
				SpecVersion: "1.0",
				Subject:     udai.UserDeviceID,
				Time:        time.Now(),
				Type:        "zone.dimo.task.smartcar.poll.scheduled",
				Data: services.SmartcarTask{
					TaskID:        udai.TaskID.String,
					UserDeviceID:  udai.UserDeviceID,
					IntegrationID: udai.IntegrationID,
					Identifiers: services.SmartcarIdentifiers{
						ID: udai.ExternalID.String,
					},
					Paths: md.SmartcarEndpoints,
				},
			}

			b, err := json.Marshal(e)
			if err != nil {
				return err
			}

			_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
				Topic: p.settings.TaskRunNowTopic,
				Key:   sarama.StringEncoder(udai.TaskID.String),
				Value: sarama.ByteEncoder(b),
			})
			if err != nil {
				return err
			}

			p.logger.Info().Str("userDeviceId", udai.UserDeviceID).Str("taskId", udai.TaskID.String).Msg("Started task.")
		}
	}

	return nil
}
