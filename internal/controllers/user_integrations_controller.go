package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// GetUserDeviceIntegration godoc
// @Description  Receive status updates about a Smartcar integration
// @Tags         user-devices
// @Success      200  {object}  controllers.GetUserDeviceIntegrationResponse
// @Router       /user/devices/{userDeviceID}/integrations/{integrationID} [get]
func (udc *UserDevicesController) GetUserDeviceIntegration(c *fiber.Ctx) error {
	userID := getUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")
	deviceExists, err := models.UserDevices(
		models.UserDeviceWhere.UserID.EQ(userID),
		models.UserDeviceWhere.ID.EQ(userDeviceID),
	).Exists(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}
	if !deviceExists {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("no user device with ID %s", userDeviceID))
	}

	apiIntegration, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("user device %s does not have integration %s", userDeviceID, integrationID))
		}
		return err
	}
	return c.JSON(GetUserDeviceIntegrationResponse{Status: apiIntegration.Status, ExternalID: apiIntegration.ExternalID, CreatedAt: apiIntegration.CreatedAt})
}

// DeleteUserDeviceIntegration godoc
// @Description  Remove an user device's integration
// @Tags         user-devices
// @Success      204
// @Router       /user/devices/{userDeviceID}/integrations/{integrationID} [delete]
func (udc *UserDevicesController) DeleteUserDeviceIntegration(c *fiber.Ctx) error {
	userID := getUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	device, err := models.UserDevices(
		models.UserDeviceWhere.UserID.EQ(userID),
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.DeviceDefinition),
		qm.Load(qm.Rels(models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("no user device with ID %s", userDeviceID))
		}
		return err
	}

	// Probably don't need two queries if you're smart
	apiIntegration, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
		qm.Load(models.UserDeviceAPIIntegrationRels.Integration),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("user device %s does not have integration %s", userDeviceID, integrationID))
		}
		return err
	}

	if apiIntegration.R.Integration.Vendor == services.SmartCarVendor {
		if apiIntegration.ExternalID.Valid {
			err = udc.taskSvc.StartSmartcarDeregistrationTasks(userDeviceID, integrationID, apiIntegration.ExternalID.String, apiIntegration.AccessToken)
			if err != nil {
				return err
			}
		}
	} else if apiIntegration.R.Integration.Vendor == "Tesla" {
		if apiIntegration.ExternalID.Valid {
			if err := udc.teslaTaskService.StopPoll(apiIntegration); err != nil {
				return err
			}
		}
	} else {
		udc.log.Warn().Msgf("Don't know how to deregister integration %s for device %s", apiIntegration.IntegrationID, userDeviceID)
	}

	_, err = apiIntegration.Delete(c.Context(), tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	err = udc.eventService.Emit(&services.Event{
		Type:    "com.dimo.zone.device.integration.delete",
		Source:  "devices-api",
		Subject: userDeviceID,
		Data: services.UserDeviceIntegrationEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: services.UserDeviceEventDevice{
				ID:    userDeviceID,
				Make:  device.R.DeviceDefinition.R.DeviceMake.Name,
				Model: device.R.DeviceDefinition.Model,
				Year:  int(device.R.DeviceDefinition.Year),
			},
			Integration: services.UserDeviceEventIntegration{
				ID:     apiIntegration.R.Integration.ID,
				Type:   apiIntegration.R.Integration.Type,
				Style:  apiIntegration.R.Integration.Style,
				Vendor: apiIntegration.R.Integration.Vendor,
			},
		},
	})
	if err != nil {
		udc.log.Err(err).Msg("Failed to emit integration deletion")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RegisterDeviceIntegration godoc
// @Description Submit credentials for registering a device with a given integration.
// @Tags user-devices
// @Accept json
// @Param userDeviceIntegrationRegistration body controllers.RegisterDeviceIntegrationRequest true "Integration credentials"
// @Success 204
// @Router /user/devices/:userDeviceID/integrations/:integrationID [post]
func (udc *UserDevicesController) RegisterDeviceIntegration(c *fiber.Ctx) error {
	userID := getUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	logger := udc.log.With().
		Str("userId", userID).
		Str("userDeviceId", userDeviceID).
		Str("integrationId", integrationID).
		Str("handler", "RegisterIntegration").
		Logger()
	logger.Info().Msg("Attempting to register device integration")

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to create transaction: %s", err))
	}
	defer tx.Rollback() //nolint

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(qm.Rels(models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("could not find device with id %s for user %s", userDeviceID, userID))
		}
		logger.Err(err).Msg("Unexpected database error searching for user device")
		return err
	}

	if !ud.CountryCode.Valid {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("device %s does not have a country code, can't check compatibility", ud.ID))
	}

	countryRecord := services.FindCountry(ud.CountryCode.String)
	if countryRecord == nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("can't find compatibility region for country %s", ud.CountryCode.String))
	}

	integ, err := models.DeviceIntegrations(
		models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(ud.DeviceDefinitionID),
		models.DeviceIntegrationWhere.IntegrationID.EQ(integrationID),
		models.DeviceIntegrationWhere.Region.EQ(countryRecord.Region),
		qm.Load(models.DeviceIntegrationRels.Integration),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn().Msg("Attempted to register a device integration that didn't exist")
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("could not find device integration for device definition %s, integration %s and country %s", ud.DeviceDefinitionID, integrationID, ud.CountryCode.String))
		}
		logger.Err(err).Msg("Unexpected database error searching for device integration")
		return err
	}

	if exists, err := models.UserDeviceAPIIntegrationExists(c.Context(), tx, userDeviceID, integrationID); err != nil {
		logger.Err(err).Msg("Unexpected database error looking for existing instance of integration")
		return err
	} else if exists {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("device %s already has a registration with integration %s, please delete that first", userDeviceID, integrationID))
	}

	// In anticipation of a bunch more of these. Maybe move to a real internal integration registry.
	// The per-integration handler is responsible for handling the fiber context and committing the
	// transaction.
	switch integ.R.Integration.Vendor {
	case services.SmartCarVendor:
		return udc.registerSmartcarIntegration(c, &logger, tx, userDeviceID, integrationID)
	case "Tesla":
		return udc.RegisterDeviceTesla(c, &logger, tx, userDeviceID, integ.R.Integration, ud)
	default:
		logger.Error().Msg("Attempted to register an unsupported integration")
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("unsupported integration %s", integrationID))
	}
}

func (udc *UserDevicesController) registerSmartcarIntegration(c *fiber.Ctx, logger *zerolog.Logger, tx *sql.Tx, userDeviceID, integrationID string) error {
	reqBody := new(RegisterDeviceIntegrationRequest)
	if err := c.BodyParser(reqBody); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	token, err := udc.smartcarClient.ExchangeCode(c.Context(), reqBody.Code, reqBody.RedirectURI)
	if err != nil {
		logger.Err(err).Msg("Error exchanging authorization code with Smartcar")
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("failure exchanging code with Smartcar: %s", err))
	}

	// TODO: Encrypt the tokens. Note that you need the client id, client secret, and redirect
	// URL to make use of the tokens, but plain text is still a bad idea.
	integration := models.UserDeviceAPIIntegration{
		UserDeviceID:     userDeviceID,
		IntegrationID:    integrationID,
		Status:           models.UserDeviceAPIIntegrationStatusPending,
		AccessToken:      token.Access,
		AccessExpiresAt:  token.AccessExpiry,
		RefreshToken:     token.Refresh,
		RefreshExpiresAt: null.TimeFrom(token.RefreshExpiry),
	}

	if err := integration.Insert(c.Context(), tx, boil.Infer()); err != nil {
		logger.Err(err).Msg("Unexpected database error inserting new Smartcar integration registration")
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.Error().Msg("Failed to commit new integration")
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("failed to commit new integration: %s", err))
	}

	if err := udc.taskSvc.StartSmartcarRegistrationTasks(userDeviceID, integrationID); err != nil {
		logger.Err(err).Msg("Unexpected error starting Smartcar Machinery tasks")
		return err
	}

	logger.Info().Msg("Finished Smartcar device registration")

	return c.SendStatus(fiber.StatusNoContent)
}

var opaqueInternalError = fiber.NewError(fiber.StatusInternalServerError, "Internal error")

func (udc *UserDevicesController) RegisterDeviceTesla(c *fiber.Ctx, logger *zerolog.Logger, tx *sql.Tx, userDeviceID string, integ *models.Integration, ud *models.UserDevice) error {
	reqBody := new(RegisterDeviceIntegrationRequest)
	if err := c.BodyParser(reqBody); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	// We'll use this to kick off the job
	teslaID, err := strconv.Atoi(reqBody.ExternalID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "externalId for Tesla must be a positive integer")
	}
	v, err := udc.teslaService.GetVehicle(reqBody.AccessToken, teslaID)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	// Prevent users from connecting a vehicle if it's already connected through another user
	// device object. Disabled outside of prod for ease of testing.
	if udc.Settings.Environment == "prod" {
		// Probably a race condition here.
		var conflict bool
		conflict, err = models.UserDevices(
			models.UserDeviceWhere.ID.NEQ(userDeviceID), // If you want to re-register, that's okay.
			models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(v.VIN)),
			models.UserDeviceWhere.VinConfirmed.EQ(true),
		).Exists(c.Context(), tx)
		if err != nil {
			return err
		}

		if conflict {
			return fiber.NewError(fiber.StatusConflict, "VIN already used for another device's integration")
		}
	}

	if err := udc.fixTeslaDeviceDefinition(c.Context(), logger, tx, integ, ud, v.VIN); err != nil {
		logger.Err(err).Msg("Failed to fix up device definition")
		return opaqueInternalError
	}

	encAccessToken, err := udc.encrypter.Encrypt(reqBody.AccessToken)
	if err != nil {
		logger.Err(err).Msg("Failed encrypting access token")
		return opaqueInternalError
	}

	encRefreshToken, err := udc.encrypter.Encrypt(reqBody.RefreshToken)
	if err != nil {
		logger.Err(err).Msg("Failed encrypting refresh token")
		return opaqueInternalError
	}

	integration := models.UserDeviceAPIIntegration{
		UserDeviceID:    userDeviceID,
		IntegrationID:   integ.ID,
		ExternalID:      null.StringFrom(reqBody.ExternalID),
		Status:          models.UserDeviceAPIIntegrationStatusPendingFirstData,
		AccessToken:     encAccessToken,
		AccessExpiresAt: time.Now().Add(time.Duration(reqBody.ExpiresIn) * time.Second),
		RefreshToken:    encRefreshToken, // Don't know when this expires.
	}

	if err := integration.Insert(c.Context(), tx, boil.Infer()); err != nil {
		logger.Err(err).Msg("Unexpected database error inserting new Tesla integration registration")
		return err
	}

	ud.VinIdentifier = null.StringFrom(v.VIN)
	ud.VinConfirmed = true
	_, err = ud.Update(c.Context(), tx, boil.Infer())
	if err != nil {
		return err
	}

	err = udc.eventService.Emit(&services.Event{
		Type:    "com.dimo.zone.device.integration.create",
		Source:  "devices-api",
		Subject: userDeviceID,
		Data: services.UserDeviceIntegrationEvent{
			Timestamp: time.Now(),
			UserID:    ud.UserID,
			Device: services.UserDeviceEventDevice{
				ID:    userDeviceID,
				Make:  "Tesla", // this method is specific to Tesla so ok to hardcode
				Model: ud.R.DeviceDefinition.Model,
				Year:  int(ud.R.DeviceDefinition.Year),
				VIN:   v.VIN,
			},
			Integration: services.UserDeviceEventIntegration{
				ID:     integ.ID,
				Type:   integ.Type,
				Style:  integ.Style,
				Vendor: integ.Vendor,
			},
		},
	})
	if err != nil {
		logger.Err(err).Msg("Failed sending device integration creation event")
	}

	if err := udc.teslaService.WakeUpVehicle(reqBody.AccessToken, teslaID); err != nil {
		logger.Err(err).Msg("Failed waking up device")
	}

	if err := udc.teslaTaskService.StartPoll(v, &integration); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Info().Msg("Finished Tesla device registration")

	return c.SendStatus(fiber.StatusNoContent)
}

// fixTeslaDeviceDefinition tries to use the VIN provided by Tesla to correct the device definition
// used by a device.
//
// We do not attempt to create any new entries in integrations, device_definitions, or
// device_integrations. This should all be handled elsewhere for Tesla.
func (udc *UserDevicesController) fixTeslaDeviceDefinition(ctx context.Context, logger *zerolog.Logger, exec boil.ContextExecutor, integ *models.Integration, ud *models.UserDevice, vin string) error {
	vinMake := "Tesla"
	vinModel := shared.VIN(vin).TeslaModel()
	vinYear := shared.VIN(vin).Year()

	dd := ud.R.DeviceDefinition

	if dd.R.DeviceMake.Name != "Tesla" || dd.Model != vinModel || int(dd.Year) != vinYear {
		logger.Warn().Msgf(
			"Device was attached to %s, %s, %d but should be %s, %s, %d",
			dd.R.DeviceMake.Name, dd.Model, dd.Year,
			vinMake, vinModel, vinYear,
		)

		region := ""
		if countryRecord := services.FindCountry(ud.CountryCode.String); countryRecord != nil {
			region = countryRecord.Region
		}

		newDD, err := models.DeviceDefinitions(
			qm.InnerJoin(models.TableNames.DeviceMakes+" on "+models.DeviceMakeTableColumns.ID+" = "+models.DeviceDefinitionTableColumns.DeviceMakeID),
			models.DeviceMakeWhere.Name.EQ(vinMake),
			models.DeviceDefinitionWhere.Model.EQ(vinModel),
			models.DeviceDefinitionWhere.Year.EQ(int16(vinYear)),
			qm.Load(
				models.DeviceDefinitionRels.DeviceIntegrations,
				models.DeviceIntegrationWhere.IntegrationID.EQ(integ.ID),
				models.DeviceIntegrationWhere.Region.EQ(region),
			),
		).One(ctx, exec)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("no device definition %s, %s, %d", vinMake, vinModel, vinYear)
			}
			return fmt.Errorf("database error: %w", err)
		}

		if len(newDD.R.DeviceIntegrations) == 0 {
			return fmt.Errorf("correct device definition %s has no integration %s for country %s", newDD.ID, integ.ID, ud.CountryCode.String)
		}

		if err := ud.SetDeviceDefinition(ctx, exec, false, newDD); err != nil {
			return fmt.Errorf("failed switching device definition to %s: %w", newDD.ID, err)
		}
	}

	return nil
}
