package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
)

// updateStatus re-populates the autopi ingest registrar topic based on data we have in user_device_api_integrations
func updateState(ctx context.Context, pdb db.Store, logger *zerolog.Logger, autoPiSvc services.AutoPiAPIService) error {
	reader := pdb.DBS().Reader

	const (
		autopi = "27qftVRWQYpVDcO5DltO5Ojbjxk"
	)

	apiInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(autopi),
		models.UserDeviceAPIIntegrationWhere.ExternalID.IsNotNull(),
	).All(ctx, reader)
	if err != nil {
		return fmt.Errorf("failed to retrieve all API integrations with external IDs: %w", err)
	}
	logger.Info().Msgf("found %d connected autopis to update status for", len(apiInts))

	for _, apiInt := range apiInts {
		err := autoPiSvc.UpdateState(apiInt.ExternalID.String, apiInt.Status)
		if err != nil {
			logger.Err(err).Msgf("failed to update status when calling autopi api for deviceId: %s", apiInt.ExternalID.String)
		} else {
			logger.Info().Msgf("successfully updated state for %s", apiInt.ExternalID.String)
		}
		time.Sleep(500)
	}

	return nil
}
