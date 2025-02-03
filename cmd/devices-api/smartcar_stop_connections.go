package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
	const smartcarIntegrationID = "22N2xaPOq2WW2gAHBHd0Ikn4Zob"
	// open csv file
	// read VIN's, get the user device api integration filtered by smartcar and get the relation to userdevice
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err, filename)
		return subcommands.ExitFailure
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV file:", err)
		return subcommands.ExitFailure
	}

	if len(records) == 0 {
		fmt.Println("CSV file is empty")
		return subcommands.ExitFailure
	}

	// Find the index of the 'vin' column
	var vinIndex = -1
	for i, columnName := range records[0] {
		if columnName == "vin" {
			vinIndex = i
			break
		}
	}

	if vinIndex == -1 {
		fmt.Println("Column 'vin' not found in CSV file")
		return subcommands.ExitFailure
	}

	// Print the VIN values
	fmt.Println("VIN values found in CSV file:", len(records))
	for _, row := range records[1:] { // Skip header row
		if vinIndex < len(row) {
			vin := row[vinIndex]
			fmt.Println("VIN:", vin)

			// get the user device api integration filtered by smartcar and get the relation to userdevice
			ud, err := models.UserDevices(models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(vin)),
				models.UserDeviceWhere.VinConfirmed.EQ(true)).One(ctx, p.pdb.DBS().Reader)
			if err != nil {
				fmt.Println("Error getting user device, continuing:", err)
				continue
			}

			scInt, err := models.UserDeviceAPIIntegrations(
				models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(ud.ID),
				models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(smartcarIntegrationID),
				qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice)).One(ctx, p.pdb.DBS().Reader)
			if err != nil {
				fmt.Println("Error getting user device api integration, continuing:", err)
				continue
			}

			err = p.stopConnections(ctx, scInt)
			if err != nil {
				fmt.Println("Error stopping connections, continuing:", err)
				continue
			}
			fmt.Println("Stopped connections for VIN:", vin)
		}
	}
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
	return nil
}
