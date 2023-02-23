package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/volatiletech/null/v8"
)

// updateStatus re-populates the autopi ingest registrar topic based on data we have in user_device_api_integrations
func updateState(ctx context.Context, pdb db.Store, logger *zerolog.Logger, autoPiSvc services.AutoPiAPIService) error {
	db := pdb.DBS().Reader

	apiInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.ExternalID.NEQ(null.StringFromPtr(nil)),
	).All(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve all API integrations with external IDs: %w", err)
	}

	for _, apiInt := range apiInts {
		err := autoPiSvc.UpdateState(apiInt.ExternalID.String, apiInt.Status)
		if err != nil {
			logger.Err(err)
		}
		time.Sleep(500)
	}

	return nil
}
