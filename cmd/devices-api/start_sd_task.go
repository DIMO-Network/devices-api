package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/Shopify/sarama"
)

type startSDTask struct {
	logger    zerolog.Logger
	container dependencyContainer
	settings  config.Settings
	pdb       db.Store

	producer  sarama.SyncProducer
	scTask    services.SmartcarTaskService
	teslaTask services.TeslaTaskService
}

func (*startSDTask) Name() string { return "start-sd-task" }
func (*startSDTask) Synopsis() string {
	return "start-sd-task args to stdout."
}
func (*startSDTask) Usage() string {
	return `start-sd-task:
	start-sd-task args.
  `
}

// nolint
func (p *startSDTask) SetFlags(f *flag.FlagSet) {

}

func (p *startSDTask) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.producer = p.container.getKafkaProducer()
	p.scTask = services.NewSmartcarTaskService(&p.settings, p.producer)
	p.teslaTask = services.NewTeslaTaskService(&p.settings, p.producer)
	err := p.startSDTaskGo()
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running SD task start.")
	}
	return subcommands.ExitSuccess
}

func (p *startSDTask) startSDTaskGo() error {
	if len(os.Args[1:]) != 2 {
		p.logger.Fatal().Msgf("Expected an argument, the task key.")
	}

	taskID := os.Args[2]

	ctx := context.Background()

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.TaskID.EQ(null.StringFrom(taskID)),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice, models.UserDeviceRels.VehicleTokenSyntheticDevice)),
	).One(ctx, p.pdb.DBS().Reader)
	if err != nil {
		return err
	}

	sd := udai.R.UserDevice.R.VehicleTokenSyntheticDevice
	if sd == nil {
		return errors.New("no synthetic device")
	}

	switch udai.IntegrationID {
	case "22N2xaPOq2WW2gAHBHd0Ikn4Zob":
		p.logger.Err(p.scTask.StartPoll(udai, sd)).Msg("xd")
	case "26A5Dk3vvvQutjSyF0Jka2DP5lg":
		p.logger.Err(p.teslaTask.StartPoll(udai, sd)).Msg("xd")
	default:
		return fmt.Errorf("unexpected integration %s", udai.IntegrationID)
	}

	return nil
}
