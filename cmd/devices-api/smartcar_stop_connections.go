package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type smartcarStopConnectionsCmd struct {
	logger          zerolog.Logger
	settings        config.Settings
	pdb             db.Store
	smartcarTaskSvc services.SmartcarTaskService
}

func (*smartcarStopConnectionsCmd) Name() string { return "smartcar-stop-connections" }
func (*smartcarStopConnectionsCmd) Synopsis() string {
	return "stops smartcar connections from a csv file with SC VIN's. Stops Polling. Marks integration as auth failure."
}
func (*smartcarStopConnectionsCmd) Usage() string {
	return `smartcar-stop-connections`
}

func (p *smartcarStopConnectionsCmd) SetFlags(_ *flag.FlagSet) {

}

func (p *smartcarStopConnectionsCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	const filename = "smartcar_stop_connections.csv"
	// open csv file
	// read VIN's, get the user device api integration filtered by smartcar and get the relation to userdevice

	return subcommands.ExitSuccess
}

func (p *smartcarStopConnectionsCmd) stopConnections(ctx context.Context, scInt *models.UserDeviceAPIIntegration) error {
	if scInt.R.UserDevice == nil {
		return fmt.Errorf("failed to find user device %s for integration %s", scInt.UserDeviceID, scInt.IntegrationID)
	}
	if !scInt.TaskID.Valid {
		return fmt.Errorf("failed to stop device integration polling; invalid task id")
	}
	err := p.smartcarTaskSvc.StopPoll(scInt)
	if err != nil {
		return fmt.Errorf("failed to stop smartcar poll: %w", err)
	}

	scInt.Status = models.UserDeviceAPIIntegrationStatusAuthenticationFailure
	if _, err := scInt.Update(ctx, p.pdb.DBS().Writer, boil.Infer()); err != nil {
		return fmt.Errorf("failed to update integration table; task id: %s; %w", scInt.TaskID.String, err)
	}
}
