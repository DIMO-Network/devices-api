package main

import (
	"context"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func populateUSAPowertrain(ctx context.Context, logger *zerolog.Logger, pdb db.Store, nhtsaService services.INHTSAService) error {
	devices, err := models.UserDevices(
		models.UserDeviceWhere.CountryCode.EQ(null.StringFrom("USA")),
		models.UserDeviceWhere.VinIdentifier.IsNotNull(),
	).All(ctx, pdb.DBS().Writer)
	if err != nil {
		return err
	}

	process := func(device *models.UserDevice) error {
		resp, err := nhtsaService.DecodeVIN(device.VinIdentifier.String)
		if err != nil {
			return err
		}
		dt, err := resp.DriveType()
		if err != nil {
			return err
		}

		md := new(services.UserDeviceMetadata)
		if err := device.Metadata.Unmarshal(md); err != nil {
			return err
		}

		md.PowertrainType = &dt
		if err := device.Metadata.Marshal(md); err != nil {
			return err
		}
		if _, err := device.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			return err
		}
		return nil
	}

	for _, device := range devices {
		if err := process(device); err != nil {
			logger.Err(err).Str("userDeviceId", device.ID).Str("vin", device.VinIdentifier.String).Msg("Failed to update powertrain metadata.")
		}
	}

	return nil
}
