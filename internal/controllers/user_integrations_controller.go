package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	smartcar "github.com/smartcar/go-sdk"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/tmpcred"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/mod/semver"
)

// GetUserDeviceIntegration godoc
// @Description Receive status updates about a Smartcar integration
// @Tags        integrations
// @Success     200 {object} controllers.GetUserDeviceIntegrationResponse
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID} [get]
func (udc *UserDevicesController) GetUserDeviceIntegration(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")
	deviceExists, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
	).Exists(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}
	if !deviceExists {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("No user device with id %q.", userDeviceID))
	}

	apiIntegration, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice, models.UserDeviceRels.VehicleTokenSyntheticDevice)),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("User device %s does not have integration %s.", userDeviceID, integrationID))
		}
		return err
	}

	resp := GetUserDeviceIntegrationResponse{
		Status:     apiIntegration.Status,
		ExternalID: apiIntegration.ExternalID,
		CreatedAt:  apiIntegration.CreatedAt,
	}

	logger := udc.log.With().Str("userDeviceId", userDeviceID).Str("integrationId", integrationID).Logger()

	// Handle fetching virtual key status
	intd, err := udc.DeviceDefSvc.GetIntegrationByID(c.Context(), integrationID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "invalid integration id")
	}

	if intd.Vendor != constants.TeslaVendor {
		return c.JSON(resp)
	}
	var meta services.UserDeviceAPIIntegrationsMetadata
	err = apiIntegration.Metadata.Unmarshal(&meta)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Integration metadata is corrupted.")
	}

	apiVersion := 1

	if meta.TeslaAPIVersion != 0 {
		apiVersion = meta.TeslaAPIVersion
	}

	minted := apiIntegration.R.UserDevice.R.VehicleTokenSyntheticDevice != nil && !apiIntegration.R.UserDevice.R.VehicleTokenSyntheticDevice.TokenID.IsZero()

	resp.Tesla = &TeslaIntegrationInfo{
		APIVersion:            apiVersion,
		MissingRequiredScopes: []string{},
	}

	if apiVersion != constants.TeslaAPIV2 {
		return c.JSON(resp)
	}

	if !apiIntegration.ExternalID.Valid || !apiIntegration.R.UserDevice.VinIdentifier.Valid {
		return fiber.NewError(fiber.StatusInternalServerError, "Missing device or integration details.")
	}

	if apiIntegration.AccessExpiresAt.Valid && apiIntegration.AccessExpiresAt.Time.Before(time.Now()) {
		return c.JSON(resp)
	}

	accessToken, err := udc.cipher.Decrypt(apiIntegration.AccessToken.String)
	if err != nil {
		return fmt.Errorf("failed to decrypt access token: %w", err)
	}

	var claims partialTeslaClaims
	_, _, err = jwt.NewParser().ParseUnverified(accessToken, &claims)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Couldn't parse access token.")
	}

	if udc.Settings.TeslaRequiredScopes != "" {
		// Yes, wasteful Split.
		for _, scope := range strings.Split(udc.Settings.TeslaRequiredScopes, ",") {
			if !slices.Contains(claims.Scopes, scope) {
				resp.Tesla.MissingRequiredScopes = append(resp.Tesla.MissingRequiredScopes, scope)
			}
		}
	}

	telemStatus, err := udc.teslaFleetAPISvc.GetTelemetrySubscriptionStatus(c.Context(), accessToken, apiIntegration.R.UserDevice.VinIdentifier.String)
	if err != nil {
		logger.Err(err).Msg("Error checking Fleet Telemetry configuration.")
		if errors.Is(err, services.ErrUnauthorized) {
			// The task-worker should get this in the API soon.
			return c.JSON(resp)
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Error checking Fleet Telemetry configuration.")
	}

	resp.Tesla.TelemetrySubscribed = telemStatus.Configured && telemStatus.KeyPaired

	// This is a bit wasteful if you are, indeed, subscribed.
	fleetStatus, err := udc.teslaFleetAPISvc.VirtualKeyConnectionStatus(c.Context(), accessToken, apiIntegration.R.UserDevice.VinIdentifier.String)
	if err != nil {
		udc.log.Err(err).Str("userDeviceId", apiIntegration.UserDeviceID).Int64("integrationId", 2).Msg("Failed to check fleet status.")
		return fiber.NewError(fiber.StatusInternalServerError, "Error checking fleet status.")
	}

	fleetTelemetryCapable := IsFleetTelemetryCapable(fleetStatus)

	firmwareUpToDate, err := IsFirmwareFleetTelemetryCapable(fleetStatus.FirmwareVersion)
	if err != nil {
		logger.Warn().Err(err).Msgf("Couldn't parse firmware version %q.", fleetStatus.FirmwareVersion)
		firmwareUpToDate = false
	}

	resp.Tesla.VirtualKeyAdded = fleetStatus.KeyPaired
	if !fleetTelemetryCapable {
		resp.Tesla.VirtualKeyStatus = Incapable
	} else if fleetStatus.KeyPaired {
		resp.Tesla.VirtualKeyStatus = Paired
	} else {
		resp.Tesla.VirtualKeyStatus = Unpaired
	}

	if fleetTelemetryCapable && !telemStatus.Configured && fleetStatus.KeyPaired && minted && firmwareUpToDate {
		vid, _ := apiIntegration.R.UserDevice.TokenID.Int64()
		err := udc.teslaFleetAPISvc.SubscribeForTelemetryData(c.Context(), accessToken, apiIntegration.R.UserDevice.VinIdentifier.String)
		// TODO(elffjs): More SD information in the logs?
		if err != nil {
			udc.log.Err(err).Int64("vehicleId", vid).Int64("integrationId", 2).Msg("Failed to configure Fleet Telemetry.")
		} else {
			resp.Tesla.TelemetrySubscribed = true
			udc.log.Info().Int64("vehicleId", vid).Int64("integrationId", 2).Msg("Successfully configured Fleet Telemetry.")
		}
	}

	return c.JSON(resp)
}

var teslaFirmwareStart = regexp.MustCompile(`^(\d{4})\.(\d+)`)

func IsFirmwareFleetTelemetryCapable(v string) (bool, error) {
	m := teslaFirmwareStart.FindStringSubmatch(v)
	if len(m) == 0 {
		return false, fmt.Errorf("unexpected firmware version format %q", v)
	}

	year, err := strconv.Atoi(m[1])
	if err != nil {
		return false, fmt.Errorf("couldn't parse year %q", m[1])
	}

	week, err := strconv.Atoi(m[2])
	if err != nil {
		return false, fmt.Errorf("couldn't parse week %q", m[2])
	}

	return year > 2024 || year == 2024 && week >= 26, nil
}

func (udc *UserDevicesController) deleteDeviceIntegration(ctx context.Context, userID, userDeviceID, integrationID string, dd *ddgrpc.GetDeviceDefinitionItemResponse, tx *sql.Tx) error {
	apiInt, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
		qm.Load(models.UserDeviceAPIIntegrationRels.SerialAftermarketDevice),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	if unit := apiInt.R.SerialAftermarketDevice; unit != nil && (unit.PairRequestID.Valid || !unit.VehicleTokenID.IsZero()) {
		return fiber.NewError(fiber.StatusBadRequest, "Must un-pair device on-chain instead.")
	}

	integ, err := udc.DeviceDefSvc.GetIntegrationByID(ctx, integrationID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "deviceDefSvc error getting integration id: "+integrationID)
	}

	switch integ.Vendor {
	case constants.SmartCarVendor:
		if apiInt.TaskID.Valid {
			err = udc.smartcarTaskSvc.StopPoll(apiInt)
			if err != nil {
				return err
			}
		}
	case constants.TeslaVendor:
		if apiInt.TaskID.Valid {
			err = udc.teslaTaskService.StopPoll(apiInt)
			if err != nil {
				return err
			}
		}
	case constants.AutoPiVendor:
		// Should never hit this.
		err = udc.autoPiIngestRegistrar.Deregister(apiInt.ExternalID.String, apiInt.UserDeviceID, apiInt.IntegrationID)
		if err != nil {
			return err
		}
	}

	_, err = apiInt.Delete(ctx, tx)
	if err != nil {
		return err
	}

	var vin string
	if apiInt.R.UserDevice.VinConfirmed {
		vin = apiInt.R.UserDevice.VinIdentifier.String
	}

	err = udc.eventService.Emit(&shared.CloudEvent[any]{
		Type:    "com.dimo.zone.device.integration.delete",
		Source:  "devices-api",
		Subject: userDeviceID,
		Data: services.UserDeviceIntegrationEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: services.UserDeviceEventDevice{
				ID:           userDeviceID,
				Make:         dd.Make.Name,
				Model:        dd.Model,
				Year:         int(dd.Year),
				VIN:          vin,
				DefinitionID: dd.Id,
			},
			Integration: services.UserDeviceEventIntegration{
				ID:     integ.Id,
				Type:   integ.Type,
				Style:  integ.Style,
				Vendor: integ.Vendor,
			},
		},
	})
	if err != nil {
		udc.log.Err(err).Msg("Failed to emit integration deletion")
	}

	return err
}

// DeleteUserDeviceIntegration godoc
// @Description Remove an integration from a device.
// @Tags        integrations
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID} [delete]
func (udc *UserDevicesController) DeleteUserDeviceIntegration(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	device, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.MintRequest),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID)),
		qm.Load(models.UserDeviceRels.VehicleTokenSyntheticDevice),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id.")
		}
		return err
	}

	if device.R.MintRequest != nil && device.R.MintRequest.Status != models.MetaTransactionRequestStatusFailed && device.R.MintRequest.Status != models.MetaTransactionRequestStatusConfirmed {
		return fiber.NewError(fiber.StatusBadRequest, "Wait for vehicle minting to complete before deleting integration.")
	}

	if len(device.R.UserDeviceAPIIntegrations) == 0 {
		// The synthetic burn event handler might have already deleted it.
		// Return success so the app doesn't freak out.
		return c.SendStatus(fiber.StatusNoContent)
	}
	integr, err := udc.DeviceDefSvc.GetIntegrationByID(c.Context(), integrationID)
	if err != nil {
		return err
	}

	autopiDeviceID := ""
	for _, udai := range device.R.UserDeviceAPIIntegrations {
		if udai.IntegrationID == integrationID && integr.Vendor == constants.AutoPiVendor {
			unit, _ := udc.autoPiSvc.GetDeviceByUnitID(udai.Serial.String)
			if unit != nil {
				autopiDeviceID = unit.ID
			} else {
				udc.log.Warn().Msgf("failed to find autopi device with serial: %s and user device id: %s", udai.Serial.String, device.ID)
			}
		}
	}

	if device.R.VehicleTokenSyntheticDevice != nil {
		sd := device.R.VehicleTokenSyntheticDevice

		integrTokenID, _ := device.R.VehicleTokenSyntheticDevice.IntegrationTokenID.Uint64()
		if integr.TokenId == integrTokenID {
			if sd.BurnRequestID.Valid {
				return fiber.NewError(fiber.StatusConflict, "Synthetic device burn in progress.")
			}
			return fiber.NewError(fiber.StatusConflict, "Burn synthetic device before deleting udai.")
		}
	}

	// Need this for activity log.
	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionBySlug(c.Context(), device.DefinitionID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+device.DefinitionID)
	}

	err = udc.deleteDeviceIntegration(c.Context(), userID, userDeviceID, integrationID, dd, tx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	udc.markAutoPiUnpaired(autopiDeviceID)

	if err := tx.Commit(); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type CommandResponse struct {
	RequestID string `json:"requestId"`
}

// TelemetrySubscribe godoc
// @Summary     Subscribe vehicle for Telemetry Data
// @Description Subscribe vehicle for Telemetry Data. Currently, this only works for Teslas connected through Tesla.
// @ID          telemetry-subscribe
// @Tags        device,integration,command
// @Produce     json
// @Param       userDeviceID  path string true "Device ID"
// @Param       integrationID path string true "Integration ID"
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/telemetry/subscribe [post]
func (udc *UserDevicesController) TelemetrySubscribe(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	logger := helpers.GetLogger(c, udc.log).With().
		Str("IntegrationID", integrationID).
		Str("Name", "Telemetry/Subscribe").
		Logger()

	logger.Info().Msg("Received command request.")

	device, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Device not found.")
		}
		logger.Err(err).Msg("Failed to search for device.")
		return opaqueInternalError
	}

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Integration not found for this device.")
		}
		logger.Err(err).Msg("Failed to search for device integration record.")
		return opaqueInternalError
	}

	if udai.Status == models.UserDeviceAPIIntegrationStatusAuthenticationFailure {
		return fiber.NewError(fiber.StatusBadRequest, "Integration credentials have expired. Reauthenticate before attempting to subscribe.")
	}

	md := new(services.UserDeviceAPIIntegrationsMetadata)
	if err := udai.Metadata.Unmarshal(md); err != nil {
		logger.Err(err).Msg("Couldn't parse metadata JSON.")
		return opaqueInternalError
	}

	if md.Commands == nil {
		return fiber.NewError(fiber.StatusBadRequest, "No commands config for integration and device")
	}

	if !slices.Contains(md.Commands.Enabled, constants.TelemetrySubscribe) {
		return fiber.NewError(fiber.StatusBadRequest, "Telemetry command not available for device and integration combination.")
	}

	integration, err := udc.DeviceDefSvc.GetIntegrationByID(c.Context(), udai.IntegrationID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "deviceDefSvc error getting integration id: "+udai.IntegrationID)
	}

	switch integration.Vendor {
	case constants.TeslaVendor:
		accessToken, err := udc.cipher.Decrypt(udai.AccessToken.String)
		if err != nil {
			return fmt.Errorf("failed to decrypt access token: %w", err)
		}
		if err := udc.teslaFleetAPISvc.SubscribeForTelemetryData(c.Context(),
			accessToken,
			device.VinIdentifier.String,
		); err != nil {
			logger.Error().Err(err).Msg("error registering for telemetry")
			var subErr *services.TeslaSubscriptionError
			if errors.As(err, &subErr) {
				switch subErr.Type {
				case services.KeyUnpaired:
					return fiber.NewError(fiber.StatusBadRequest, "Virtual key not paired with vehicle.")
				case services.UnsupportedVehicle:
					return fiber.NewError(fiber.StatusBadRequest, "Pre-2021 Model S and X do not support telemetry.")
				case services.UnsupportedFirmware:
					return fiber.NewError(fiber.StatusBadRequest, "Vehicle firmware version is earlier than 2024.26.")
				}
			}
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to update telemetry configuration.")
		}
	default:
		return fiber.NewError(fiber.StatusBadRequest, "Integration not supported for this command")
	}

	logger.Info().Msg("Successfully subscribed to telemetry")

	return c.JSON(fiber.Map{"message": "Successfully subscribed to vehicle telemetry."})
}

// GetAftermarketDeviceInfo godoc
// @Description Gets the information about the aftermarket device by serial number.
// @Tags        integrations
// @Produce     json
// @Param       serial path     string true "AutoPi unit id or Macaron serial number"
// @Success     200    {object} controllers.AutoPiDeviceInfo
// @Security    BearerAuth
// @Router      /aftermarket/device/by-serial/{serial} [get]
func (udc *UserDevicesController) GetAftermarketDeviceInfo(c *fiber.Ctx) error {
	const minimumAutoPiRelease = "v1.22.8" // correct semver has leading v

	serial := c.Locals("serial").(string)

	var claim, pair, unpair *TransactionStatus

	var tokenID *big.Int
	var ethereumAddress, beneficiaryAddress *common.Address
	var ownerAddress *string // Frontend is doing a case-sensitive match.

	var mfr *ManufacturerInfo

	dbUnit, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.Serial.EQ(serial),
		qm.Load(models.AftermarketDeviceRels.ClaimMetaTransactionRequest),
		qm.Load(models.AftermarketDeviceRels.PairRequest),
		qm.Load(models.AftermarketDeviceRels.UnpairRequest),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else {
		tokenID = dbUnit.TokenID.Int(nil)

		addr := common.BytesToAddress(dbUnit.EthereumAddress)
		ethereumAddress = &addr

		if dbUnit.OwnerAddress.Valid {
			addr := common.BytesToAddress(dbUnit.OwnerAddress.Bytes)
			addrStr := addr.Hex()
			ownerAddress = &addrStr
			beneficiaryAddress = &addr
			// We do this because we're worried the claim originated in the chain and not our
			// backend.
			claim = &TransactionStatus{
				Status: models.MetaTransactionRequestStatusConfirmed,
			}
		}

		if dbUnit.Beneficiary.Valid {
			addr := common.BytesToAddress(dbUnit.Beneficiary.Bytes)
			beneficiaryAddress = &addr
		}

		if req := dbUnit.R.ClaimMetaTransactionRequest; req != nil {
			claim = &TransactionStatus{
				Status:        req.Status,
				CreatedAt:     req.CreatedAt,
				UpdatedAt:     req.UpdatedAt,
				FailureReason: req.FailureReason.Ptr(),
			}
			if req.Hash.Valid {
				hash := hexutil.Encode(req.Hash.Bytes)
				claim.Hash = &hash
			}
		}

		// Check for pair.
		if req := dbUnit.R.PairRequest; req != nil {
			pair = &TransactionStatus{
				Status:        req.Status,
				CreatedAt:     req.CreatedAt,
				UpdatedAt:     req.UpdatedAt,
				FailureReason: req.FailureReason.Ptr(),
			}
			if req.Status != models.MetaTransactionRequestStatusUnsubmitted {
				hash := hexutil.Encode(req.Hash.Bytes)
				pair.Hash = &hash
			}
		}

		// Check for unpair.
		if req := dbUnit.R.UnpairRequest; req != nil {
			unpair = &TransactionStatus{
				Status:        req.Status,
				CreatedAt:     req.CreatedAt,
				UpdatedAt:     req.UpdatedAt,
				FailureReason: req.FailureReason.Ptr(),
			}
			if req.Status != models.MetaTransactionRequestStatusUnsubmitted {
				hash := hexutil.Encode(req.Hash.Bytes)
				unpair.Hash = &hash
			}
		}

		tib := dbUnit.DeviceManufacturerTokenID.Int(nil)

		dm, err := udc.DeviceDefSvc.GetMakeByTokenID(c.Context(), tib)
		if err != nil {
			return err
		}

		mfr = &ManufacturerInfo{
			TokenID: tib,
			Name:    dm.Name,
		}
	}

	if mfr != nil && mfr.Name != constants.AutoPiVendor {
		// Might be a Macaron
		adi := AutoPiDeviceInfo{
			IsUpdated:          true,
			UnitID:             serial,
			ShouldUpdate:       false,
			TokenID:            tokenID,
			EthereumAddress:    ethereumAddress,
			OwnerAddress:       ownerAddress,
			BeneficiaryAddress: beneficiaryAddress,
			Claim:              claim,
			Pair:               pair,
			Unpair:             unpair,
			Manufacturer:       mfr,
		}

		return c.JSON(adi)
	}

	// This is hitting AutoPi.
	unit, err := udc.autoPiSvc.GetDeviceByUnitID(serial)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Serial %s unknown to AutoPi.", serial))
		}
		return err
	}

	// Must be an AutoPi.
	shouldUpdate := false
	if udc.Settings.IsProduction() {
		version := unit.Release.Version
		if string(unit.Release.Version[0]) != "v" {
			version = "v" + version
		}
		shouldUpdate = semver.Compare(version, minimumAutoPiRelease) < 0
	}

	adi := AutoPiDeviceInfo{
		IsUpdated:          unit.IsUpdated,
		DeviceID:           unit.ID,
		UnitID:             unit.UnitID,
		DockerReleases:     unit.DockerReleases,
		HwRevision:         unit.HwRevision,
		Template:           unit.Template,
		LastCommunication:  unit.LastCommunication,
		ReleaseVersion:     unit.Release.Version,
		ShouldUpdate:       shouldUpdate,
		TokenID:            tokenID,
		EthereumAddress:    ethereumAddress,
		OwnerAddress:       ownerAddress,
		BeneficiaryAddress: beneficiaryAddress,
		Claim:              claim,
		Pair:               pair,
		Unpair:             unpair,
		Manufacturer:       mfr,
	}
	return c.JSON(adi)
}

func (udc *UserDevicesController) registerDeviceIntegrationInner(c *fiber.Ctx, userID, userDeviceID, integrationID string) error {
	logger := udc.log.With().
		Str("userId", userID).
		Str("userDeviceId", userDeviceID).
		Str("integrationId", integrationID).
		Str("route", c.Route().Path).
		Logger()
	logger.Info().Msg("Attempting to register device integration")

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to create transaction: %s", err))
	}
	defer tx.Rollback() //nolint
	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
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

	countryRecord := constants.FindCountry(ud.CountryCode.String)
	if countryRecord == nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("can't find compatibility region for country %s", ud.CountryCode.String))
	}
	logger = logger.With().Str("region", countryRecord.Region).Logger()

	integration, err := udc.DeviceDefSvc.GetIntegrationByID(c.Context(), integrationID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "failed to get integration with id: "+integrationID)
	}

	// if exists, likely means already handled from previous /fromsmartcar endpoint, just return nil but log warn in case
	if exists, err := models.UserDeviceAPIIntegrationExists(c.Context(), tx, userDeviceID, integrationID); err != nil {
		logger.Err(err).Msg("Unexpected database error looking for existing instance of integration")
		return err
	} else if exists {
		// if the user has already registered it from previous step, we can just log and return success
		logger.Warn().Msgf("userDeviceId %s already has a user_device_api_integration with integrationId %s, continuing - but consider deleting if support issues", userDeviceID, integrationID)
		return nil
	}

	var regErr error
	// The per-integration handler is responsible for handling the fiber context and committing the
	// transaction.
	switch vendor := integration.Vendor; vendor {
	case constants.SmartCarVendor:
		regErr = udc.registerSmartcarIntegration(c, &logger, tx, integration, ud)
	case constants.TeslaVendor:
		regErr = udc.registerDeviceTesla(c, &logger, tx, userDeviceID, integration, ud)
	case constants.CompassIotVendor:
		regErr = udc.registerDeviceCompass(c, &logger, tx, userDeviceID, integration, ud)
	case constants.AutoPiVendor:
		logger.Error().Msg("autopi register request via invalid route: /user/devices/:userDeviceID/integrations/:integrationID")
		return errors.New("this route cannot be used to register an autopi anymore - update your client")
	default:
		logger.Error().Str("vendor", vendor).Msg("Attempted to register an unsupported integration")
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("unsupported integration %s", integrationID))
	}

	if regErr != nil {
		return regErr
	}

	udc.runPostRegistration(c.Context(), &logger, userDeviceID, integrationID, integration)

	return nil
}

// RegisterDeviceIntegration godoc
// @Description Submit credentials for registering a device with a given integration. This must be called for any new pairing as well as eg. /fromsmartcar
// @Tags        integrations
// @Accept      json
// @Param       userDeviceIntegrationRegistration body controllers.RegisterDeviceIntegrationRequest true "Integration credentials"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID} [post]
func (udc *UserDevicesController) RegisterDeviceIntegration(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	return udc.registerDeviceIntegrationInner(c, userID, userDeviceID, integrationID)
}

/** Refactored / helper methods **/

// runPostRegistration runs tasks that should be run after a successful integration. For now, this
// just means emitting an event to topic.event for the activity log.
func (udc *UserDevicesController) runPostRegistration(ctx context.Context, logger *zerolog.Logger, userDeviceID, integrationID string, integ *ddgrpc.Integration) {
	// Just reload the entire tree of attributes. Too many things modify this during the registration flow.
	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).One(ctx, udc.DBS().Reader)
	if err != nil {
		logger.Err(err).Msg("Couldn't retrieve UDAI for post-registration tasks.")
		return
	}

	ud := udai.R.UserDevice
	definitionID := ud.DefinitionID

	// pull dd info again - don't pass it in, as it may have changed
	dd, err2 := udc.DeviceDefSvc.GetDeviceDefinitionBySlug(ctx, definitionID)
	if err2 != nil {
		tid, _ := ud.TokenID.Uint64()
		logger.Err(err2).
			Str("definitionId", ud.DefinitionID).
			Str("userDeviceId", userDeviceID).
			Uint64("tokenID", tid).
			Msg("failed to retrieve device definition")
	}

	err = udc.eventService.Emit(
		&shared.CloudEvent[any]{
			Type:    "com.dimo.zone.device.integration.create",
			Source:  "devices-api",
			Subject: userDeviceID,
			Data: services.UserDeviceIntegrationEvent{
				Timestamp: time.Now(),
				UserID:    ud.UserID,
				Device: services.UserDeviceEventDevice{
					ID:           userDeviceID,
					Make:         dd.Make.Name,
					Model:        dd.Model,
					Year:         int(dd.Year),
					VIN:          ud.VinIdentifier.String,
					DefinitionID: dd.Id,
				},
				Integration: services.UserDeviceEventIntegration{
					ID:     integ.Id,
					Type:   integ.Type,
					Style:  integ.Style,
					Vendor: integ.Vendor,
				},
			},
		},
	)
	if err != nil {
		logger.Err(err).Msg("Failed to emit integration event.")
	}
}

var smartcarCallErr = fiber.NewError(fiber.StatusInternalServerError, "Error communicating with Smartcar.")

func (udc *UserDevicesController) registerSmartcarIntegration(c *fiber.Ctx, logger *zerolog.Logger, tx *sql.Tx, integ *ddgrpc.Integration, ud *models.UserDevice) error {
	reqBody := new(RegisterDeviceIntegrationRequest)
	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request JSON body.")
	}
	var token *smartcar.Token
	// check for token in redis, if exists do not call this.
	if ud.VinIdentifier.Valid {
		scTokenGet, err := udc.redisCache.Get(c.Context(), buildSmartcarTokenKey(ud.VinIdentifier.String, ud.UserID)).Result()
		if err == nil && len(scTokenGet) > 0 {
			decrypted, err := udc.cipher.Decrypt(scTokenGet)
			if err != nil {
				return errors.Wrap(err, "failed to decrypt sc token")
			}
			// found existing token
			token = &smartcar.Token{}
			err = json.Unmarshal([]byte(decrypted), token)
			if err != nil {
				udc.log.Err(err).Msgf("failed to unmarshal smartcar token found in redis cache for vin: %s", ud.VinIdentifier.String)
			}
			// clear cache
			udc.redisCache.Del(c.Context(), buildSmartcarTokenKey(ud.VinIdentifier.String, ud.UserID))
		}
	}
	if token == nil {
		// no token found or could be unmarshalled so try exchangecode, assumption is it has not been called before for this code
		var err error
		token, err = udc.smartcarClient.ExchangeCode(c.Context(), reqBody.Code, reqBody.RedirectURI)
		if err != nil {
			var scErr *services.SmartcarError
			if errors.As(err, &scErr) {
				logger.Error().Msgf("Failed exchanging Authorization code. Status code %d, request id %s, and body `%s`.", scErr.Code, scErr.RequestID, string(scErr.Body))
			} else {
				logger.Err(err).Msg("Failed to exchange authorization code with Smartcar.")
			}

			// This may not be the user's fault, but 400 for now.
			return fiber.NewError(fiber.StatusBadRequest, "Failed to exchange authorization code with Smartcar.")
		}
	}

	scUserID, err := udc.smartcarClient.GetUserID(c.Context(), token.Access)
	if err != nil {
		logger.Err(err).Msg("Failed to retrieve user ID from Smartcar.")
		return smartcarCallErr
	}

	externalID, err := udc.smartcarClient.GetExternalID(c.Context(), token.Access)
	if err != nil {
		logger.Err(err).Msg("Failed to retrieve vehicle ID from Smartcar.")
		return smartcarCallErr
	}
	// by default use vin from userdevice, unless if it is not confirmed, in that case pull from SC
	vin := ud.VinIdentifier.String
	if !ud.VinConfirmed {
		vin, err = udc.smartcarClient.GetVIN(c.Context(), token.Access, externalID)
		if err != nil {
			logger.Err(err).Msg("Failed to retrieve VIN from Smartcar.")
			return smartcarCallErr
		}

		if ud.VinConfirmed && ud.VinIdentifier.String != vin {
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("Vehicle's confirmed VIN does not match Smartcar's %s.", vin))
		}
	}
	localLog := logger.With().Str("vin", vin).Str("userId", ud.UserID).Logger()

	// Prevent users from connecting a vehicle if it's already connected through another user
	// device object. Disabled outside of prod for ease of testing.
	if udc.Settings.IsProduction() {
		if vin[0:3] == "0SC" {
			localLog.Error().Msgf("Smartcar test VIN %s is not allowed.", vin)
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("Smartcar test VIN %s is not allowed.", vin))
		}
		// Probably a race condition here. Need to either lock something or impose a greater
		// isolation level.
		conflict, err := models.UserDevices(
			models.UserDeviceWhere.ID.NEQ(ud.ID), // If you want to re-register, or register a different integration, that's okay.
			models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(vin)),
			models.UserDeviceWhere.VinConfirmed.EQ(true),
		).Exists(c.Context(), tx)
		if err != nil {
			localLog.Err(err).Msg("Failed to search for VIN conflicts.")
			return opaqueInternalError
		}

		if conflict {
			localLog.Error().Msg("VIN %s already in use.")
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("VIN %s in use by a previously connected device.", ud.VinIdentifier.String))
		}

		// TODO(elffjs): Super unfortunate to put another blocking call here. Maybe we do a batch?
		// Doing it only in production because sometimes it's nice to use Teslas to test Smartcar on dev.
		info, err := udc.smartcarClient.GetInfo(c.Context(), token.Access, externalID)
		if err != nil {
			logger.Err(err).Msg("Error listing vehicle details for Smartcar Tesla check.")
			return smartcarCallErr
		}

		if info.Make == "TESLA" { // They always have this in ALL CAPS for some reason.
			return fiber.NewError(fiber.StatusBadRequest, "Teslas should be connected using the official integration.")
		}
	}

	endpoints, err := udc.smartcarClient.GetEndpoints(c.Context(), token.Access, externalID)
	if err != nil {
		localLog.Err(err).Msg("Failed to retrieve permissions from Smartcar.")
		return smartcarCallErr
	}

	var commands *services.UserDeviceAPIIntegrationsMetadataCommands

	doorControl, err := udc.smartcarClient.HasDoorControl(c.Context(), token.Access, externalID)
	if err != nil {
		localLog.Err(err).Msg("Failed to retrieve door control permissions from Smartcar.")
		return smartcarCallErr
	}

	if doorControl {
		commands = udc.smartcarClient.GetAvailableCommands()
	}

	meta := services.UserDeviceAPIIntegrationsMetadata{
		SmartcarUserID:    &scUserID,
		SmartcarEndpoints: endpoints,
		Commands:          commands,
	}

	b, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	encAccess, err := udc.cipher.Encrypt(token.Access)
	if err != nil {
		return opaqueInternalError
	}

	encRefresh, err := udc.cipher.Encrypt(token.Refresh)
	if err != nil {
		return opaqueInternalError
	}

	err = udc.userDeviceSvc.CreateIntegration(c.Context(), tx, ud.ID, integ.Id, externalID, encAccess, token.AccessExpiry, encRefresh, b)
	if err != nil {
		localLog.Err(err).Msg("Unexpected database error inserting new Smartcar integration registration.")
		return opaqueInternalError
	}
	// todo this may cause issues
	if !ud.VinConfirmed {
		ud.VinIdentifier = null.StringFrom(strings.ToUpper(vin))
		ud.VinConfirmed = true
		_, err = ud.Update(c.Context(), tx, boil.Infer())
		if err != nil {
			return opaqueInternalError
		}
	}

	if err := tx.Commit(); err != nil {
		localLog.Error().Msg("Failed to commit new user device integration.")
		return opaqueInternalError
	}

	localLog.Info().Msg("Finished Smartcar device registration.")

	return c.SendStatus(fiber.StatusNoContent)
}

func (udc *UserDevicesController) registerDeviceCompass(c *fiber.Ctx, logger *zerolog.Logger, tx *sql.Tx, userDeviceID string, integ *ddgrpc.Integration, ud *models.UserDevice) error {
	if existingIntegrations, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integ.Id),
	).Count(c.Context(), tx); err != nil {
		return err
	} else if existingIntegrations > 0 {
		return nil // integration already exists, continue
	}
	// err if no vin set
	if ud.VinIdentifier.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "VIN identifier is not set.")
	}

	// insert record
	integration := models.UserDeviceAPIIntegration{
		UserDeviceID:  userDeviceID,
		IntegrationID: integ.Id,
		ExternalID:    null.StringFrom(ud.VinIdentifier.String),
		Status:        models.UserDeviceAPIIntegrationStatusPendingFirstData,
	}
	if err := integration.Insert(c.Context(), tx, boil.Infer()); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Info().Msg("Finished CompassIot registration")

	return c.SendStatus(fiber.StatusNoContent)
}

func IsFleetTelemetryCapable(fs *services.VehicleFleetStatus) bool {
	// We used to check for the presence of a meaningful value (not ""
	// or "unknown") for fleet_telemetry_version, but this started
	// populating on old cars that are not capable of streaming.
	return fs.VehicleCommandProtocolRequired || !fs.DiscountedDeviceData
}

func (udc *UserDevicesController) registerDeviceTesla(c *fiber.Ctx, logger *zerolog.Logger, tx *sql.Tx, userDeviceID string, integ *ddgrpc.Integration, ud *models.UserDevice) error {
	if existingIntegrations, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
	).Count(c.Context(), tx); err != nil {
		return err
	} else if existingIntegrations > 0 {
		return fiber.NewError(fiber.StatusConflict, "Delete existing integration before connecting through Tesla.")
	}

	reqBody := new(RegisterDeviceIntegrationRequest)
	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	if reqBody.ExternalID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing externalId field.")
	}

	if reqBody.AccessToken != "" {
		return fiber.NewError(fiber.StatusBadRequest, "We no longer support connecting Teslas by submitting access and refresh tokens. Please submit an authorization code for the Fleet API.")
	}

	// We'll use this to kick off the job
	teslaID, err := strconv.Atoi(reqBody.ExternalID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse externalId %q as an integer.", teslaID))
	}

	userAddr, err := helpers.GetJWTEthAddr(c)
	if err != nil {
		return err
	}

	// Yes, yes.
	store := &tmpcred.Store{
		Redis:  udc.redisCache,
		Cipher: udc.cipher,
	}

	cred, err := store.Retrieve(c.Context(), userAddr)
	if err != nil {
		if errors.Is(err, tmpcred.ErrNotFound) {
			return fiber.NewError(fiber.StatusBadRequest, "No credentials found for user.")
		}
		return err
	}

	reqBody.RefreshToken = cred.RefreshToken
	reqBody.AccessToken = cred.AccessToken
	reqBody.ExpiresIn = int(time.Until(cred.Expiry).Seconds())

	v, err := udc.getTeslaVehicle(c.Context(), reqBody.AccessToken, teslaID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't retrieve vehicle from Tesla.")
	}

	fs, err := udc.teslaFleetAPISvc.VirtualKeyConnectionStatus(c.Context(), reqBody.AccessToken, v.VIN)
	if err != nil {
		return fmt.Errorf("couldn't retrieve fleet status from Tesla: %w", err)
	}

	fleetTelemetryCapable := IsFleetTelemetryCapable(fs)

	// Prevent users from connecting a vehicle if it's already connected through another user
	// device object. Disabled outside of prod for ease of testing.
	if udc.Settings.IsProduction() {
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
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("VIN %s in use by another vehicle.", v.VIN))
		}
	}

	if err := fixTeslaDeviceDefinition(c.Context(), logger, tx, integ, ud, v.VIN); err != nil {
		return fmt.Errorf("correcting device definition: %w", err)
	}

	encAccessToken, err := udc.cipher.Encrypt(reqBody.AccessToken)
	if err != nil {
		return err
	}

	encRefreshToken, err := udc.cipher.Encrypt(reqBody.RefreshToken)
	if err != nil {
		return err
	}

	commands, err := udc.teslaFleetAPISvc.GetAvailableCommands(reqBody.AccessToken)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't determine available commands.")
	}

	meta := services.UserDeviceAPIIntegrationsMetadata{
		Commands:                   commands,
		TeslaAPIVersion:            constants.TeslaAPIV2,
		TeslaVehicleID:             v.VehicleID,
		TeslaVIN:                   v.VIN,
		TeslaDiscountedData:        &fs.DiscountedDeviceData,
		TeslaFleetTelemetryCapable: &fleetTelemetryCapable,
	}

	b, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	taskID := ksuid.New().String()

	integration := models.UserDeviceAPIIntegration{
		UserDeviceID:    userDeviceID,
		IntegrationID:   integ.Id,
		ExternalID:      null.StringFrom(reqBody.ExternalID),
		Status:          models.UserDeviceAPIIntegrationStatusPendingFirstData,
		AccessToken:     null.StringFrom(encAccessToken),
		AccessExpiresAt: null.TimeFrom(time.Now().Add(time.Duration(reqBody.ExpiresIn) * time.Second)),
		RefreshToken:    null.StringFrom(encRefreshToken), // Don't know when this expires.
		TaskID:          null.StringFrom(taskID),
		Metadata:        null.JSONFrom(b),
	}

	if err := integration.Insert(c.Context(), tx, boil.Infer()); err != nil {
		return err
	}

	ud.VinIdentifier = null.StringFrom(strings.ToUpper(v.VIN))
	ud.VinConfirmed = true
	_, err = ud.Update(c.Context(), tx, boil.Infer())
	if err != nil {
		return err
	}

	if err := udc.wakeupTeslaVehicle(c.Context(), reqBody.AccessToken, teslaID); err != nil {
		logger.Err(err).Msg("Couldn't wake up Tesla.")
	}

	if udc.Settings.IsProduction() && !ud.TokenID.IsZero() {
		tokenID, ok := ud.TokenID.Int64()
		if !ok {
			return errors.New("failed to parse vehicle token id")
		}
		udc.requestValuation(v.VIN, userDeviceID, tokenID)
		udc.requestInstantOffer(userDeviceID, tokenID)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Info().Msg("Finished Tesla device registration")

	return c.SendStatus(fiber.StatusNoContent)
}

func (udc *UserDevicesController) wakeupTeslaVehicle(ctx context.Context, token string, vehicleID int) error {
	return udc.teslaFleetAPISvc.WakeUpVehicle(ctx, token, vehicleID)
}

func (udc *UserDevicesController) getTeslaVehicle(ctx context.Context, token string, vehicleID int) (*services.TeslaVehicle, error) {
	return udc.teslaFleetAPISvc.GetVehicle(ctx, token, vehicleID)
}

// fixTeslaDeviceDefinition tries to use the VIN provided by Tesla to correct the device definition
// used by a device.
//
// We do not attempt to create any new entries in integrations, device_definitions, or
// device_integrations. This should all be handled elsewhere for Tesla.
func fixTeslaDeviceDefinition(ctx context.Context, logger *zerolog.Logger, exec boil.ContextExecutor, _ *ddgrpc.Integration, ud *models.UserDevice, vin string) error {
	vinMake := "Tesla"
	vinModel := shared.VIN(vin).TeslaModel()
	vinYear := shared.VIN(vin).Year()
	definitionID := fmt.Sprintf("%s_%s_%d", shared.SlugString(vinMake), shared.SlugString(vinModel), vinYear)

	if definitionID != ud.DefinitionID {
		logger.Warn().Msgf(
			"Device moving to new device definition from %s to %s", ud.DefinitionID, definitionID,
		)
		ud.DefinitionID = definitionID
		_, err := ud.Update(ctx, exec, boil.Infer())
		if err != nil {
			return err
		}
	}

	return nil
}

/** Structs for request / response **/

type UserDeviceIntegrationStatus struct {
	IntegrationID     string                 `json:"integrationId"`
	Status            string                 `json:"status"`
	CreatedAt         time.Time              `json:"createdAt"`
	ExternalID        *string                `json:"externalId"`
	UpdatedAt         time.Time              `json:"updatedAt"`
	Metadata          null.JSON              `json:"metadata" swaggertype:"string"`
	IntegrationVendor string                 `json:"integrationVendor"`
	Mint              *SyntheticDeviceStatus `json:"syntheticDevice,omitempty"`
	TokenID           *big.Int               `json:"tokenId,omitempty"`
}

// RegisterDeviceIntegrationRequest carries credentials used to connect the device to a given
// integration.
type RegisterDeviceIntegrationRequest struct {
	// Code is an OAuth authorization code. Not used in all integrations.
	Code string `json:"code"`
	// RedirectURI is the OAuth redirect URI used by the frontend. Not used in all integrations.
	RedirectURI string `json:"redirectURI"`
	// ExternalID specifies which vehicle on the account to select. It is only used for
	// software integrations.
	ExternalID   string `json:"externalId"`
	AccessToken  string `json:"accessToken"`
	ExpiresIn    int    `json:"expiresIn"`
	RefreshToken string `json:"refreshToken"`
	Version      int    `json:"version"`
}

type TeslaIntegrationInfo struct {
	// APIVersion is the version of the Tesla API being used. There are currently two valid values:
	// 1 is the old "Owner API", 2 is the new "Fleet API".
	APIVersion int `json:"apiVersion"`
	// VirtualKeyAdded is true if the DIMO virtual key has been added to the vehicle. This is deprecated.
	// Use VirtualKeyStatus instead.
	VirtualKeyAdded bool `json:"virtualKeyAdded"`
	// TelemetrySubscribed is true if DIMO has subscribed to the vehicle's telemetry stream. Note that
	// virtual key pairing is required for this to work.
	TelemetrySubscribed bool `json:"telemetrySubscribed"`
	// VirtualKeyStatus indicates whether the Tesla can pair with DIMO's virtual key; and if it can,
	// whether the key has indeed been paired.
	VirtualKeyStatus VirtualKeyStatus `json:"virtualKeyStatus" swaggertype:"string" enums:"Paired,Unpaired,Incapable"`
	// MissingRequiredScopes lists scopes required by DIMO that we're missing.
	MissingRequiredScopes []string `json:"missingRequiredScopes"`
}

type VirtualKeyStatus int

const (
	Incapable VirtualKeyStatus = iota
	Paired
	Unpaired
)

func (s VirtualKeyStatus) String() string {
	switch s {
	case Incapable:
		return "Incapable"
	case Paired:
		return "Paired"
	case Unpaired:
		return "Unpaired"
	}
	return ""
}

func (s VirtualKeyStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *VirtualKeyStatus) UnmarshalText(text []byte) error {
	switch str := string(text); str {
	case "Incapable":
		*s = Incapable
	case "Paired":
		*s = Paired
	case "Unpaired":
		*s = Unpaired
	default:
		return fmt.Errorf("unrecognized status %q", str)
	}
	return nil
}

type GetUserDeviceIntegrationResponse struct {
	// Status is one of "Pending", "PendingFirstData", "Active", "Failed", "DuplicateIntegration".
	Status string `json:"status"`
	// ExternalID is the identifier used by the third party for the device. It may be absent if we
	// haven't authorized yet.
	ExternalID null.String `json:"externalId" swaggertype:"string"`

	// Contains further details about tesla integration status
	Tesla *TeslaIntegrationInfo `json:"tesla,omitempty"`

	// CreatedAt is the creation time of this integration for this device.
	CreatedAt time.Time `json:"createdAt"`
}

type ManufacturerInfo struct {
	TokenID *big.Int `json:"tokenId"`
	Name    string   `json:"name"`
}

// AutoPiDeviceInfo is used to get the info about a unit
type AutoPiDeviceInfo struct {
	IsUpdated         bool      `json:"isUpdated"`
	DeviceID          string    `json:"deviceId"`
	UnitID            string    `json:"unitId"`
	DockerReleases    []int     `json:"dockerReleases"`
	HwRevision        string    `json:"hwRevision"`
	Template          int       `json:"template"`
	LastCommunication time.Time `json:"lastCommunication"`
	ReleaseVersion    string    `json:"releaseVersion"`
	ShouldUpdate      bool      `json:"shouldUpdate"`

	TokenID            *big.Int        `json:"tokenId,omitempty"`
	EthereumAddress    *common.Address `json:"ethereumAddress,omitempty"`
	OwnerAddress       *string         `json:"ownerAddress,omitempty"`
	BeneficiaryAddress *common.Address `json:"beneficiaryAddress,omitempty"`

	// Claim contains the status of the on-chain claiming meta-transaction.
	Claim *TransactionStatus `json:"claim,omitempty"`
	// Pair contains the status of the on-chain pairing meta-transaction.
	Pair *TransactionStatus `json:"pair,omitempty"`
	// Unpair contains the status of the on-chain unpairing meta-transaction.
	Unpair *TransactionStatus `json:"unpair,omitempty"`

	Manufacturer *ManufacturerInfo `json:"manufacturer,omitempty"`
}

// TransactionStatus summarizes the state of an on-chain operation.
type TransactionStatus struct {
	// Status is the state of the transaction performing this operation.
	Status string `json:"status" enums:"Unsubmitted,Submitted,Mined,Confirmed,Failed" example:"Mined"`
	// Hash is the hexidecimal transaction hash, available for any transaction at the Submitted stage or greater.
	Hash *string `json:"hash,omitempty" example:"0x28b4662f1e1b15083261a4a5077664f4003d58cb528826b7aab7fad466c28e70"`
	// CreatedAt is the timestamp of the creation of the meta-transaction.
	CreatedAt time.Time `json:"createdAt" example:"2022-10-01T09:22:21.002Z"`
	// UpdatedAt is the last time we updated the status of the transaction.
	UpdatedAt time.Time `json:"updatedAt" example:"2022-10-01T09:22:26.337Z"`
	// FailureReason is populated with a human-readable error message if the status
	// is "Failed" because of an on-chain revert and we were able to decode the reason.
	FailureReason *string `json:"failureReason,omitempty"`
}
