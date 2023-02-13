package main

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func startSmartcarFromRefresh(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb db.Store, cipher shared.Cipher, userDeviceID string, scClient services.SmartcarClient, scTask services.SmartcarTaskService, ddSvc services.DeviceDefinitionService) error {
	db := pdb.DBS().Writer
	scInt, err := ddSvc.GetIntegrationByVendor(ctx, "SmartCar")
	if err != nil {
		return fmt.Errorf("couldn't find SmartCar integration: %w", err)
	}

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(scInt.Id),
	).One(ctx, db)
	if err != nil {
		return fmt.Errorf("couldn't find a Smartcar integration for %s: %w", userDeviceID, err)
	}

	officialClient := smartcar.NewClient()
	officialClient.SetAPIVersion("2.0")

	authClient := officialClient.NewAuth(&smartcar.AuthParams{
		ClientID:     settings.SmartcarClientID,
		ClientSecret: settings.SmartcarClientSecret,
		// Don't need anything else.
	})

	newToken, err := authClient.ExchangeRefreshToken(ctx, &smartcar.ExchangeRefreshTokenParams{Token: udai.RefreshToken.String})
	if err != nil {
		return fmt.Errorf("couldn't exchange refresh token with Smartcar: %w", err)
	}

	logger.Info().Str("userDeviceId", userDeviceID).Str("refreshToken", newToken.Refresh).Msg("Got new refresh token.")

	encAccess, err := cipher.Encrypt(newToken.Access)
	if err != nil {
		return fmt.Errorf("couldn't encrypt access token: %w", err)
	}

	encRefresh, err := cipher.Encrypt(newToken.Refresh)
	if err != nil {
		return fmt.Errorf("couldn't encrypt refresh token: %w", err)
	}

	realExternalID, err := scClient.GetExternalID(ctx, newToken.Access)
	if err != nil {
		return fmt.Errorf("couldn't retrieve external ID from Smartcar: %w", err)
	}

	if realExternalID != udai.ExternalID.String {
		return fmt.Errorf("token should have been for external ID %s but was for %s", udai.ExternalID.String, realExternalID)
	}

	perms, err := scClient.GetEndpoints(ctx, newToken.Access, realExternalID)
	if err != nil {
		return fmt.Errorf("couldn't get permissions from Smartcar: %w", err)
	}

	meta := new(services.UserDeviceAPIIntegrationsMetadata)
	if err := udai.Metadata.Unmarshal(meta); err != nil {
		return fmt.Errorf("couldn't parse integration metadata: %w", err)
	}

	meta.SmartcarEndpoints = perms

	udai.AccessToken = null.StringFrom(encAccess)
	udai.RefreshToken = null.StringFrom(encRefresh)
	udai.AccessExpiresAt = null.TimeFrom(newToken.AccessExpiry)
	udai.TaskID = null.StringFrom(ksuid.New().String())

	if err := udai.Metadata.Marshal(meta); err != nil {
		return fmt.Errorf("couldn't serialize updated integration metadata: %w", err)
	}

	if _, err := udai.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("couldn't update integration: %w", err)
	}

	if err := scTask.StartPoll(udai); err != nil {
		return fmt.Errorf("couldn't start Kafka task: %w", err)
	}

	return nil
}
