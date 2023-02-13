package main

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/Shopify/sarama"
	"github.com/volatiletech/null/v8"
)

func remakeSmartcarTopic(ctx context.Context, pdb db.Store, producer sarama.SyncProducer, ddSvc services.DeviceDefinitionService) error {
	reg := services.NewIngestRegistrar(services.Smartcar, producer)
	db := pdb.DBS().Reader

	// Grab the Smartcar integration ID, there should be exactly one.
	var scIntID string
	scInt, err := ddSvc.GetIntegrationByVendor(ctx, "SmartCar")
	if err != nil {
		return fmt.Errorf("failed to retrieve Smartcar integration: %w", err)
	}
	scIntID = scInt.Id

	// Find all integration instances that have acquired Smartcar ids.
	apiInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(scIntID),
		models.UserDeviceAPIIntegrationWhere.ExternalID.NEQ(null.StringFromPtr(nil)),
	).All(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve all API integrations with external IDs: %w", err)
	}

	// For each of these send a new registration message, keyed by Smartcar vehicle ID.
	for _, apiInt := range apiInts {
		if err := reg.Register(apiInt.ExternalID.String, apiInt.UserDeviceID, scIntID); err != nil {
			return fmt.Errorf("failed to register Smartcar-DIMO id link for device %s: %w", apiInt.UserDeviceID, err)
		}
	}

	return nil
}
