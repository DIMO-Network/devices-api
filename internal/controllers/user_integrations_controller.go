package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	smartcar "github.com/smartcar/go-sdk"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/Shopify/sarama"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"golang.org/x/exp/slices"
	"golang.org/x/mod/semver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GetUserDeviceIntegration godoc
// @Description Receive status updates about a Smartcar integration
// @Tags        integrations
// @Success     200 {object} controllers.GetUserDeviceIntegrationResponse
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID} [get]
func (udc *UserDevicesController) GetUserDeviceIntegration(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
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

func (udc *UserDevicesController) deleteDeviceIntegration(ctx context.Context, userID, userDeviceID, integrationID string, dd *ddgrpc.GetDeviceDefinitionItemResponse) error {
	tx, err := udc.DBS().Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	apiInt, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
		qm.Load(models.UserDeviceAPIIntegrationRels.AutopiUnit),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	integ, err := udc.DeviceDefSvc.GetIntegrationByID(ctx, integrationID)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "deviceDefSvc error getting integration id: "+integrationID)
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
		if unit := apiInt.R.AutopiUnit; unit != nil && unit.PairRequestID.Valid {
			return fiber.NewError(fiber.StatusConflict, "Must un-pair on-chain before deleting integration.")
		}

		err = udc.autoPiIngestRegistrar.Deregister(apiInt.ExternalID.String, apiInt.UserDeviceID, apiInt.IntegrationID)
		if err != nil {
			return err
		}
	}

	_, err = apiInt.Delete(ctx, tx)
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
				Make:  dd.Make.Name,
				Model: dd.Type.Model,
				Year:  int(dd.Type.Year),
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

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}

	defer tx.Rollback() //nolint

	device, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID)),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id.")
		}
		return err
	}

	if len(device.R.UserDeviceAPIIntegrations) == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Device does presently have that integration.")
	}

	// Need this for activity log.
	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), device.DeviceDefinitionID)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+device.DeviceDefinitionID)
	}

	err = udc.deleteDeviceIntegration(c.Context(), userID, userDeviceID, integrationID, dd)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetIntegrations godoc
// @Description gets list of integrations we have defined
// @Tags        integrations
// @Produce     json
// @Success     200 {array} ddgrpc.Integration
// @Security    BearerAuth
// @Router      /integrations [get]
func (udc *UserDevicesController) GetIntegrations(c *fiber.Ctx) error {
	all, err := udc.DeviceDefSvc.GetIntegrations(c.Context())
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get integrations")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"integrations": all,
	})
}

// SendAutoPiCommand godoc
// @Description Closed off in prod. Submit a raw autopi command to unit. Device must be registered with autopi before this can be used
// @Tags        integrations
// @Accept      json
// @Param       AutoPiCommandRequest body controllers.AutoPiCommandRequest true "raw autopi command"
// @Success     200
// @Security    BearerAuth
// @Router      /user/devices/:userDeviceID/autopi/command [post]
func (udc *UserDevicesController) SendAutoPiCommand(c *fiber.Ctx) error {
	if udc.Settings.Environment == "prod" {
		return c.SendStatus(fiber.StatusGone)
	}
	userID := helpers.GetUserID(c)
	userDeviceID := c.Params("userDeviceID")
	req := new(AutoPiCommandRequest)
	err := c.BodyParser(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "unable to parse body json")
	}

	logger := udc.log.With().
		Str("userId", userID).
		Str("userDeviceId", userDeviceID).
		Str("handler", "SendAutoPiCommand").
		Str("autopiCmd", req.Command).
		Logger()
	logger.Info().Msg("Attempting to send autopi raw command")

	udai, _, err := udc.DeviceDefIntSvc.FindUserDeviceAutoPiIntegration(c.Context(), udc.DBS().Writer, userDeviceID, userID)
	if err != nil {
		logger.Err(err).Msg("error finding user device autopi integration")
		return err
	}
	apUnit, err := models.AutopiUnits(models.AutopiUnitWhere.AutopiDeviceID.EQ(udai.ExternalID), models.AutopiUnitWhere.UserID.EQ(null.StringFrom(userID))).
		One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}
	// call autopi
	commandResponse, err := udc.autoPiSvc.CommandRaw(c.Context(), apUnit.AutopiUnitID, apUnit.AutopiDeviceID.String, req.Command, userDeviceID)
	if err != nil {
		logger.Err(err).Msg("autopi returned error when calling raw command")
		return errors.Wrapf(err, "autopi returned error when calling raw command: %s", req.Command)
	}

	return c.Status(fiber.StatusOK).JSON(commandResponse)
}

// GetCommandRequestStatus godoc
// @Summary     Get the status of a submitted command.
// @Description Get the status of a submitted command by request id.
// @ID          get-command-request-status
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandRequestStatusResp
// @Produce     json
// @Param       userDeviceID  path string true "Device ID"
// @Param       integrationID path string true "Integration ID"
// @Param       requestID path string true "Command request ID"
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/{requestID} [get]
func (udc *UserDevicesController) GetCommandRequestStatus(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	requestID := c.Params("requestID")

	// Don't actually validate userDeviceID or integrationID, just following a URL pattern.
	// Is this beyond the pale?
	cr, err := models.DeviceCommandRequests(
		models.DeviceCommandRequestWhere.ID.EQ(requestID),
		qm.Load(models.DeviceCommandRequestRels.UserDevice),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No command request with that id found.")
		}
		udc.log.Err(err).Msg("Failed to search for command status.")
		return opaqueInternalError
	}

	if cr.R.UserDevice.UserID != userID {
		return fiber.NewError(fiber.StatusNotFound, "No command request with that id found.")
	}

	dcr := CommandRequestStatusResp{
		ID:        requestID,
		Command:   cr.Command,
		Status:    cr.Status,
		CreatedAt: cr.CreatedAt,
		UpdatedAt: cr.UpdatedAt,
	}

	return c.JSON(dcr)
}

type CommandRequestStatusResp struct {
	ID        string    `json:"id" example:"2D8LqUHQtaMHH6LYPqznmJMBeZm"`
	Command   string    `json:"command" example:"doors/unlock"`
	Status    string    `json:"status" enums:"Pending,Complete,Failed" example:"Complete"`
	CreatedAt time.Time `json:"createdAt" example:"2022-08-09T19:38:39Z"`
	UpdatedAt time.Time `json:"updatedAt" example:"2022-08-09T19:39:22Z"`
}

// handleEnqueueCommand enqueues the command specified by commandPath with the
// appropriate task service.
//
// Grabs user ID, device ID, and integration ID from Ctx.
func (udc *UserDevicesController) handleEnqueueCommand(c *fiber.Ctx, commandPath string) error {
	userID := helpers.GetUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	logger := udc.log.With().
		Str("feature", "commands").
		Str("userId", userID).
		Str("userDeviceId", userDeviceID).
		Str("integrationId", integrationID).
		Str("commandPath", commandPath).
		Logger()

	logger.Info().Msg("Received command request.")

	// Checking both that the device exists and that the user owns it.
	deviceOK, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
	).Exists(c.Context(), udc.DBS().Reader)
	if err != nil {
		logger.Err(err).Msg("Failed to search for device.")
		return opaqueInternalError
	}

	if !deviceOK {
		return fiber.NewError(fiber.StatusNotFound, "Device not found.")
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

	if udai.Status != models.UserDeviceAPIIntegrationStatusActive {
		return fiber.NewError(fiber.StatusConflict, "Integration is not active for this device.")
	}

	md := new(services.UserDeviceAPIIntegrationsMetadata)
	if err := udai.Metadata.Unmarshal(md); err != nil {
		logger.Err(err).Msg("Couldn't parse metadata JSON.")
		return opaqueInternalError
	}

	// TODO(elffjs): This map is ugly. Surely we interface our way out of this?
	commandMap := map[string]map[string]func(udai *models.UserDeviceAPIIntegration) (string, error){
		constants.SmartCarVendor: {
			"doors/unlock": udc.smartcarTaskSvc.UnlockDoors,
			"doors/lock":   udc.smartcarTaskSvc.LockDoors,
		},
		constants.TeslaVendor: {
			"doors/unlock": udc.teslaTaskService.UnlockDoors,
			"doors/lock":   udc.teslaTaskService.LockDoors,
			"trunk/open":   udc.teslaTaskService.OpenTrunk,
			"frunk/open":   udc.teslaTaskService.OpenFrunk,
		},
	}

	integration, err := udc.DeviceDefSvc.GetIntegrationByID(c.Context(), udai.IntegrationID)

	if err != nil {
		return helpers.GrpcErrorToFiber(err, "deviceDefSvc error getting integration id: "+udai.IntegrationID)
	}

	vendorCommandMap, ok := commandMap[integration.Vendor]
	if !ok {
		return fiber.NewError(fiber.StatusConflict, "Integration is not capable of this command.")
	}

	// This correctly handles md.Commands.Enabled being nil.
	if !slices.Contains(md.Commands.Enabled, commandPath) {
		return fiber.NewError(fiber.StatusConflict, "Integration is not capable of this command with this device.")
	}

	commandFunc, ok := vendorCommandMap[commandPath]
	if !ok {
		// Should never get here.
		logger.Error().Msg("Command was enabled for this device, but there is no function to execute it.")
		return fiber.NewError(fiber.StatusConflict, "Integration is not capable of this command.")
	}

	subTaskID, err := commandFunc(udai)
	if err != nil {
		logger.Err(err).Msg("Failed to start command task.")
		return opaqueInternalError
	}

	comRow := &models.DeviceCommandRequest{
		ID:            subTaskID,
		UserDeviceID:  userDeviceID,
		IntegrationID: integrationID,
		Command:       commandPath,
		Status:        models.DeviceCommandRequestStatusPending,
	}

	if err := comRow.Insert(c.Context(), udc.DBS().Writer, boil.Infer()); err != nil {
		logger.Err(err).Msg("Couldn't insert device command request record.")
		return opaqueInternalError
	}

	logger.Info().Msg("Successfully enqueued command.")

	return c.JSON(CommandResponse{RequestID: subTaskID})
}

type CommandResponse struct {
	RequestID string `json:"requestId"`
}

// UnlockDoors godoc
// @Summary     Unlock the device's doors
// @Description Unlock the device's doors.
// @ID          unlock-doors
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       userDeviceID  path string true "Device ID"
// @Param       integrationID path string true "Integration ID"
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/doors/unlock [post]
func (udc *UserDevicesController) UnlockDoors(c *fiber.Ctx) error {
	return udc.handleEnqueueCommand(c, "doors/unlock")
}

// LockDoors godoc
// @Summary     Lock the device's doors
// @Description Lock the device's doors.
// @ID          lock-doors
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       userDeviceID  path string true "Device ID"
// @Param       integrationID path string true "Integration ID"
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/doors/lock [post]
func (udc *UserDevicesController) LockDoors(c *fiber.Ctx) error {
	return udc.handleEnqueueCommand(c, "doors/lock")
}

// OpenTrunk godoc
// @Summary     Open the device's rear trunk
// @Description Open the device's front trunk. Currently, this only works for Teslas connected through Tesla.
// @ID          open-trunk
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       userDeviceID  path string true "Device ID"
// @Param       integrationID path string true "Integration ID"
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/trunk/open [post]
func (udc *UserDevicesController) OpenTrunk(c *fiber.Ctx) error {
	return udc.handleEnqueueCommand(c, "trunk/open")
}

// OpenFrunk godoc
// @Summary     Open the device's front trunk
// @Description Open the device's front trunk. Currently, this only works for Teslas connected through Tesla.
// @ID          open-frunk
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       userDeviceID  path string true "Device ID"
// @Param       integrationID path string true "Integration ID"
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/frunk/open [post]
func (udc *UserDevicesController) OpenFrunk(c *fiber.Ctx) error {
	return udc.handleEnqueueCommand(c, "frunk/open")
}

// GetAutoPiCommandStatus godoc
// @Description gets the status of an autopi raw command by jobID
// @Tags        integrations
// @Produce     json
// @Param       jobID path     string true "job id, from autopi"
// @Success     200   {object} services.AutoPiCommandJob
// @Security    BearerAuth
// @Router      /user/devices/:userDeviceID/autopi/command/:jobID [get]
func (udc *UserDevicesController) GetAutoPiCommandStatus(c *fiber.Ctx) error {
	_ = helpers.GetUserID(c)
	userDeviceID := c.Params("userDeviceID")
	jobID := c.Params("jobID")

	job, dbJob, err := udc.autoPiSvc.GetCommandStatus(c.Context(), jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusBadRequest).SendString("no job found with provided jobID")
		}
		return err
	}
	if dbJob.UserDeviceID.String != userDeviceID {
		return c.Status(fiber.StatusBadRequest).SendString("no job found")
	}
	return c.Status(fiber.StatusOK).JSON(job)
}

// GetAutoPiUnitInfo godoc
// @Description gets the information about the autopi by the unitId
// @Tags        integrations
// @Produce     json
// @Param       unitID path     string true "autopi unit id"
// @Success     200    {object} controllers.AutoPiDeviceInfo
// @Security    BearerAuth
// @Router      /autopi/unit/:unitID [get]
func (udc *UserDevicesController) GetAutoPiUnitInfo(c *fiber.Ctx) error {
	const minimumAutoPiRelease = "v1.22.8" // correct semver has leading v

	rawUnitID := c.Params("unitID")
	v, unitID := services.ValidateAndCleanUUID(rawUnitID)
	if !v {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Invalid serial number: %q", rawUnitID))
	}
	userID := helpers.GetUserID(c)
	// check if unitId has already been assigned to a different user - don't allow querying in this case
	udai, _ := udc.autoPiSvc.GetUserDeviceIntegrationByUnitID(c.Context(), unitID)
	if udai != nil {
		if udai.R.UserDevice.UserID != userID {
			udc.log.Warn().Str("userID", userID).Str("autopiUnitID", unitID).
				Msg("failed to validate autopi unit belongs to user for get info")
			return fiber.NewError(fiber.StatusForbidden, "AutoPi owned by another user.")
		}
	}

	unit, err := udc.autoPiSvc.GetDeviceByUnitID(unitID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return fiber.ErrNotFound
		}
		return err
	}

	svc := semver.Compare("v"+unit.Release.Version, minimumAutoPiRelease)

	//If you are not in prod, do not require an update.
	if udc.Settings.Environment != "prod" {
		svc = 0
	}

	var claim, pair, unpair *AutoPiTransactionStatus

	var tokenID *big.Int
	var ethereumAddress, ownerAddress *common.Address

	dbUnit, err := models.AutopiUnits(
		models.AutopiUnitWhere.AutopiUnitID.EQ(unitID),
		qm.Load(models.AutopiUnitRels.ClaimMetaTransactionRequest),
		qm.Load(models.AutopiUnitRels.PairRequest),
		qm.Load(models.AutopiUnitRels.UnpairRequest),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else {
		if !dbUnit.TokenID.IsZero() {
			tokenID = dbUnit.TokenID.Int(nil)
		}

		if dbUnit.EthereumAddress.Valid {
			addr := common.BytesToAddress(dbUnit.EthereumAddress.Bytes)
			ethereumAddress = &addr
		}

		if dbUnit.OwnerAddress.Valid {
			addr := common.BytesToAddress(dbUnit.OwnerAddress.Bytes)
			ownerAddress = &addr
			claim = &AutoPiTransactionStatus{
				Status: models.MetaTransactionRequestStatusConfirmed,
			}
		}

		if req := dbUnit.R.ClaimMetaTransactionRequest; req != nil {
			claim = &AutoPiTransactionStatus{
				Status:    req.Status,
				CreatedAt: req.CreatedAt,
				UpdatedAt: req.UpdatedAt,
			}
			if req.Status != models.MetaTransactionRequestStatusUnsubmitted {
				hash := hexutil.Encode(req.Hash.Bytes)
				claim.Hash = &hash
			}
		}

		// Check for pair.
		if req := dbUnit.R.PairRequest; req != nil {
			pair = &AutoPiTransactionStatus{
				Status:    req.Status,
				CreatedAt: req.CreatedAt,
				UpdatedAt: req.UpdatedAt,
			}
			if req.Status != models.MetaTransactionRequestStatusUnsubmitted {
				hash := hexutil.Encode(req.Hash.Bytes)
				pair.Hash = &hash
			}
		}

		// Check for unpair.
		if req := dbUnit.R.UnpairRequest; req != nil {
			unpair = &AutoPiTransactionStatus{
				Status:    req.Status,
				CreatedAt: req.CreatedAt,
				UpdatedAt: req.UpdatedAt,
			}
			if req.Status != models.MetaTransactionRequestStatusUnsubmitted {
				hash := hexutil.Encode(req.Hash.Bytes)
				unpair.Hash = &hash
			}
		}
	}

	adi := AutoPiDeviceInfo{
		IsUpdated:         unit.IsUpdated,
		DeviceID:          unit.ID,
		UnitID:            unit.UnitID,
		DockerReleases:    unit.DockerReleases,
		HwRevision:        unit.HwRevision,
		Template:          unit.Template,
		LastCommunication: unit.LastCommunication,
		ReleaseVersion:    unit.Release.Version,
		ShouldUpdate:      svc < 0,
		TokenID:           tokenID,
		EthereumAddress:   ethereumAddress,
		OwnerAddress:      ownerAddress,
		Claim:             claim,
		Pair:              pair,
		Unpair:            unpair,
	}
	return c.JSON(adi)
}

// GetIsAutoPiOnline godoc
// @Description gets whether the autopi is online right now, if already paired with a user, makes sure user has access. returns json with {"online": true/false}
// @Tags        integrations
// @Produce     json
// @Param       unitID path string true "autopi unit id"
// @Success     200
// @Security    BearerAuth
// @Router      /autopi/unit/:unitID/is-online [get]
func (udc *UserDevicesController) GetIsAutoPiOnline(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	unitID := c.Params("unitID")

	valid, unitID := services.ValidateAndCleanUUID(unitID)
	if !valid {
		return fiber.NewError(fiber.StatusBadRequest, "Unit id is not a valid UUID.")
	}

	logger := udc.log.With().Str("userId", userID).Str("autoPiUnitId", unitID).Logger()

	var userDeviceID string

	// Create a record, using information from the AutoPi API, if necessary.
	autopiUnit, err := models.FindAutopiUnit(c.Context(), udc.DBS().Reader, unitID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		logger.Info().Msg("Creating AutoPi record.")

		apiUnit, err := udc.autoPiSvc.GetDeviceByUnitID(unitID)
		if err != nil {
			logger.Err(err).Msg("Failed to retrieve device from AutoPi API.")
			return fiber.NewError(fiber.StatusInternalServerError, "AutoPi API error.")
		}

		var maybeAddr null.Bytes

		if strAddr := apiUnit.EthereumAddress; common.IsHexAddress(strAddr) {
			maybeAddr = null.BytesFrom(common.FromHex(strAddr))
		} else {
			logger.Warn().Str("address", apiUnit.EthereumAddress).Msg("Invalid device Ethereum address from AutoPi.")
		}

		autopiUnit = &models.AutopiUnit{
			AutopiUnitID:    unitID,
			AutopiDeviceID:  null.StringFrom(apiUnit.ID),
			EthereumAddress: maybeAddr,
		}

		err = autopiUnit.Insert(c.Context(), udc.DBS().Writer, boil.Infer())
		if err != nil {
			logger.Err(err).Msg("Failed to insert new AutoPi record.")
			return opaqueInternalError
		}
	case err != nil:
		logger.Err(err).Msg("Failed searching for AutoPi in database.")
		return opaqueInternalError
	default:
		if uid := autopiUnit.UserID; uid.Valid && uid.String != userID {
			logger.Warn().Err(err).Str("ownerUserId", uid.String).Msg("Attempting to poll AutoPi owned by another user.")
			return fiber.NewError(fiber.StatusForbidden, "AutoPi belongs to another user.")
		}

		// This does not return an error if it doesn't find a row; instead, udai will be nil.
		udai, err := udc.autoPiSvc.GetUserDeviceIntegrationByUnitID(c.Context(), unitID)
		if err != nil {
			logger.Err(err).Msg("Failed to look up AutoPi pairing record.")
			return opaqueInternalError
		}

		if udai != nil {
			userDeviceID = udai.UserDeviceID
		}
	}

	// send command without webhook since we'll just query the jobid
	commandResponse, err := udc.autoPiSvc.CommandRaw(c.Context(), unitID, autopiUnit.AutopiDeviceID.String, "test.ping", userDeviceID)
	if err != nil {
		logger.Err(err).Msg("failed to send command to autopi api")
		return fiber.NewError(fiber.StatusInternalServerError, "Partner API returned an error")
	}
	// for loop with wait timer of 1 second at begining that calls autopi get job id
	backoffSchedule := []time.Duration{
		2 * time.Second,
		1 * time.Second,
		1 * time.Second,
		1 * time.Second,
		1 * time.Second,
		1 * time.Second,
	}
	online := false
	for _, backoff := range backoffSchedule {
		time.Sleep(backoff)
		job, _, err := udc.autoPiSvc.GetCommandStatus(c.Context(), commandResponse.Jid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "job id not found"})
			}
			continue // try again if error
		}
		if job.CommandState == "COMMAND_EXECUTED" {
			online = true
			break
		}
		if job.CommandState == "TIMEOUT" {
			break
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"online": online,
	})
}

// StartAutoPiUpdateTask godoc
// @Description checks to see if autopi unit needs to be updated, and starts update process if so.
// @Tags        integrations
// @Produce     json
// @Param       unitID path     string true "autopi unit id", ie. physical barcode
// @Success     200    {object} services.AutoPiTask
// @Security    BearerAuth
// @Router      /autopi/unit/:unitID/update [post]
func (udc *UserDevicesController) StartAutoPiUpdateTask(c *fiber.Ctx) error {
	unitID := c.Params("unitID") // save in task
	v, unitID := services.ValidateAndCleanUUID(unitID)
	if !v {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	userID := helpers.GetUserID(c)
	deviceID := ""

	// check if unitId has already been assigned to a different user - don't allow querying in this case
	autopiUnit, err := models.FindAutopiUnit(c.Context(), udc.DBS().Reader, unitID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if autopiUnit != nil {
		if autopiUnit.UserID != null.StringFrom(userID) {
			return c.SendStatus(fiber.StatusForbidden)
		}
		deviceID = autopiUnit.AutopiDeviceID.String
	}

	// check if device already updated
	unit, err := udc.autoPiSvc.GetDeviceByUnitID(unitID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Device not found.")
		}
		return err
	}
	if unit.IsUpdated {
		return c.JSON(services.AutoPiTask{
			TaskID:      "0",
			Status:      string(services.Success),
			Description: "autopi device is already up to date running version " + unit.Release.Version,
			Code:        200,
		})
	}
	if len(deviceID) == 0 {
		deviceID = unit.ID
	}
	// insert autopi unit if not claimed
	if autopiUnit == nil {
		autopiUnit = &models.AutopiUnit{
			AutopiUnitID:   unitID,
			AutopiDeviceID: null.StringFrom(deviceID),
			UserID:         null.StringFrom(userID),
		}
		err = autopiUnit.Insert(c.Context(), udc.DBS().Writer, boil.Infer())
		if err != nil {
			return err
		}
	}
	// fire off task
	taskID, err := udc.autoPiTaskService.StartAutoPiUpdate(deviceID, userID, unitID)
	if err != nil {
		return err
	}

	return c.JSON(services.AutoPiTask{
		TaskID:      taskID,
		Status:      "Pending",
		Description: "",
		Code:        100,
	})
}

// GetAutoPiTask godoc
// @Description gets the status of an autopi related task. In future could be other tasks too?
// @Tags        integrations
// @Produce     json
// @Param       taskID path     string true "task id", returned from endpoint that starts a task
// @Success     200    {object} services.AutoPiTask
// @Security    BearerAuth
// @Router      /autopi/task/:taskID [get]
func (udc *UserDevicesController) GetAutoPiTask(c *fiber.Ctx) error {
	taskID := c.Params("taskID") // save in task
	if len(taskID) == 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	//userID := api.GetUserID(c)
	task, err := udc.autoPiTaskService.GetTaskStatus(c.Context(), taskID)
	if err != nil {
		return err
	}

	// todo somewhere need to check this userID has access to that taskID
	return c.JSON(task)
}

// GetAutoPiClaimMessage godoc
// @Description Return the EIP-712 payload to be signed for AutoPi device claiming.
// @Produce json
// @Param unitID path string true "AutoPi unit id"
// @Success 200 {object} signer.TypedData
// @Security BearerAuth
// @Router /autopi/unit/:unitID/commands/claim [get]
func (udc *UserDevicesController) GetAutoPiClaimMessage(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	unitID := c.Params("unitID")

	logger := udc.log.With().Str("userId", userID).Str("unitId", unitID).Logger()
	logger.Info().Msg("Got AutoPi claim request.")

	unit, err := models.AutopiUnits(
		models.AutopiUnitWhere.AutopiUnitID.EQ(unitID),
		qm.Load(models.AutopiUnitRels.ClaimMetaTransactionRequest),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info().Msg("Unknown unit id.")
			return fiber.NewError(fiber.StatusNotFound, "AutoPi not minted, or unit ID invalid.")
		}
		logger.Err(err).Msg("Database failure searching for AutoPi.")
		return fiber.NewError(fiber.StatusInternalServerError, "Internal error.")
	}

	if unit.UserID.Valid && unit.UserID.String != userID {
		logger.Error().Str("existingUserId", unit.UserID.String).Msg("AutoPi already attached to another user.")
		return fiber.NewError(fiber.StatusForbidden, "AutoPi paired to another user.")
	}

	if unit.OwnerAddress.Valid {
		return fiber.NewError(fiber.StatusConflict, "Device already claimed.")
	}

	if unit.R.ClaimMetaTransactionRequest != nil && unit.R.ClaimMetaTransactionRequest.Status != "Failed" {
		return fiber.NewError(fiber.StatusConflict, "Claiming transaction in progress.")
	}

	if unit.TokenID.IsZero() {
		logger.Error().Msg("AutoPi not minted.")
		return fiber.NewError(fiber.StatusConflict, "AutoPi not minted.")
	}

	apToken := unit.TokenID.Int(nil)

	// TODO(elffjs): Really shouldn't be dialing so much.
	conn, err := grpc.Dial(udc.Settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		udc.log.Err(err).Msg("Failed to create users API client.")
		return opaqueInternalError
	}
	defer conn.Close()

	usersClient := pb.NewUserServiceClient(conn)

	user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Couldn't retrieve user record.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		udc.log.Error().Msg("No Ethereum address on file for user.")
		return fiber.NewError(fiber.StatusConflict, "User does not have an Ethereum address.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	cads := &registry.ClaimAftermarketDeviceSign{
		AftermarketDeviceNode: apToken,
		Owner:                 common.HexToAddress(*user.EthereumAddress),
	}

	var out *signer.TypedData = client.GetPayload(cads)

	return c.JSON(out)
}

// GetAutoPiPairMessage godoc
// @Description Return the EIP-712 payload to be signed for AutoPi device pairing. The device must
// @Description either already be integrated with the vehicle, or you must provide its unit id
// @Description as a query parameter. In the latter case, the integration process will start
// @Description once the transaction confirms.
// @Produce json
// @Param userDeviceID path string true "Device id"
// @Param external_id query string false "External id, for now AutoPi unit id"
// @Success 200 {object} signer.TypedData "EIP-712 message for pairing."
// @Security BearerAuth
// @Router /user/devices/:userDeviceID/autopi/commands/pair [get]
func (udc *UserDevicesController) GetAutoPiPairMessage(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	userDeviceID := c.Params("userDeviceID")

	logger := udc.log.With().Str("userId", userID).Str("userDeviceId", userDeviceID).Logger()

	logger.Info().Msg("Got AutoPi pair request.")

	autoPiInt, err := udc.DeviceDefIntSvc.GetAutoPiIntegration(c.Context())
	if err != nil {
		logger.Err(err).Msg("Failed to retrieve AutoPi integration.")
		return helpers.GrpcErrorToFiber(err, "failed to retrieve AutoPi integration.")
	}

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.VehicleNFT),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id found.")
		}
		logger.Err(err).Msg("Database failure searching for device.")
		return opaqueInternalError
	}

	var autoPiUnit *models.AutopiUnit

	if extID := c.Query("external_id"); extID != "" {
		unitID, err := uuid.Parse(extID)
		if err != nil {
			return err
		}

		autoPiUnit, err = models.AutopiUnits(
			models.AutopiUnitWhere.AutopiUnitID.EQ(unitID.String()),
			qm.Load(models.AutopiUnitRels.PairRequest),
			qm.Load(models.AutopiUnitRels.UnpairRequest),
		).One(c.Context(), udc.DBS().Reader)
		if err != nil {
			return err
		}
	}

	udai, err := ud.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(autoPiInt.Id),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.AutopiUnit, models.AutopiUnitRels.PairRequest)),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.AutopiUnit, models.AutopiUnitRels.UnpairRequest)),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Err(err).Msg("Database failure searching for device's AutoPi integration.")
			return opaqueInternalError
		}
	} else {
		if !udai.AutopiUnitID.Valid {
			return opaqueInternalError
		}

		// Conflict with web2 pairing?
		if autoPiUnit != nil && (!udai.AutopiUnitID.Valid || udai.AutopiUnitID.String != autoPiUnit.AutopiUnitID) {
			return fiber.NewError(fiber.StatusConflict, "Vehicle already paired with another AutoPi.")
		}

		autoPiUnit = udai.R.AutopiUnit
	}

	if autoPiUnit.R.PairRequest != nil && autoPiUnit.R.PairRequest.Status != "Failed" {
		if autoPiUnit.R.PairRequest.Status == models.MetaTransactionRequestStatusConfirmed {
			return fiber.NewError(fiber.StatusConflict, "AutoPi already paired.")
		}
		return fiber.NewError(fiber.StatusConflict, "AutoPi pairing in process.")
	}

	if autoPiUnit.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "AutoPi not yet minted.")
	}

	if ud.R.VehicleNFT == nil || ud.R.VehicleNFT.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not yet minted.")
	}

	if !autoPiUnit.OwnerAddress.Valid {
		return fiber.NewError(fiber.StatusConflict, "Device not yet claimed.")
	}

	if common.BytesToAddress(autoPiUnit.OwnerAddress.Bytes) != common.BytesToAddress(ud.R.VehicleNFT.OwnerAddress.Bytes) {
		return fiber.NewError(fiber.StatusConflict, "AutoPi and vehicle have different owners.")
	}

	apToken := autoPiUnit.TokenID.Int(nil)
	vehicleToken := ud.R.VehicleNFT.TokenID.Int(nil)

	// TODO(elffjs): Really shouldn't be dialing so much.
	conn, err := grpc.Dial(udc.Settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		udc.log.Err(err).Msg("Failed to create users API client.")
		return opaqueInternalError
	}
	defer conn.Close()

	usersClient := pb.NewUserServiceClient(conn)

	user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Failed to retrieve user information.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusConflict, "User does not have an Ethereum address.")
	}

	if common.HexToAddress(*user.EthereumAddress) != common.BytesToAddress(autoPiUnit.OwnerAddress.Bytes) {
		return fiber.NewError(fiber.StatusConflict, "AutoPi claimed by another user.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	pads := &registry.PairAftermarketDeviceSign{
		AftermarketDeviceNode: apToken,
		VehicleNode:           vehicleToken,
	}

	var out *signer.TypedData = client.GetPayload(pads)

	return c.JSON(out)
}

// PostPairAutoPi godoc
// @Description Submit the signature for pairing this device with its attached AutoPi.
// @Produce json
// @Param userDeviceID path string true "Device id"
// @Param userSignature body controllers.AutoPiPairRequest true "User signature."
// @Security BearerAuth
// @Router /user/devices/:userDeviceID/autopi/commands/pair [post]
func (udc *UserDevicesController) PostPairAutoPi(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	userDeviceID := c.Params("userDeviceID")

	logger := udc.log.With().Str("userId", userID).Str("userDeviceId", userDeviceID).Str("route", c.Route().Name).Logger()
	logger.Info().Msg("Got AutoPi pair request.")

	autoPiInt, err := udc.DeviceDefIntSvc.GetAutoPiIntegration(c.Context())
	if err != nil {
		logger.Err(err).Msg("Failed to retrieve AutoPi integration.")
		return helpers.GrpcErrorToFiber(err, "failed to retrieve AutoPi integration.")
	}

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.VehicleNFT),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id found.")
		}
		logger.Err(err).Msg("Database failure searching for device.")
		return opaqueInternalError
	}

	if ud.UserID != userID {
		// Err on the side of privacy.
		return fiber.NewError(fiber.StatusNotFound, "No device with that id found.")
	}

	var pairReq AutoPiPairRequest
	err = c.BodyParser(&pairReq)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	var autoPiUnit *models.AutopiUnit

	if extIDStr := pairReq.ExternalID; extIDStr != "" {
		unitID, err := uuid.Parse(extIDStr)
		if err != nil {
			return err
		}

		logger = logger.With().Str("autoPiUnitId", unitID.String()).Logger()

		autoPiUnit, err = models.AutopiUnits(
			models.AutopiUnitWhere.AutopiUnitID.EQ(unitID.String()),
			qm.Load(models.AutopiUnitRels.PairRequest),
			qm.Load(models.AutopiUnitRels.UnpairRequest),
			qm.Load(models.AutopiUnitRels.UserDeviceAPIIntegrations),
		).One(c.Context(), udc.DBS().Reader)
		if err != nil {
			return err
		}

		for _, udai := range autoPiUnit.R.UserDeviceAPIIntegrations {
			if udai.UserDeviceID != userDeviceID {
				logger.Error().Str("existingUserDeviceId", udai.UserDeviceID).Msg("AutoPi already web2-paired with another vehicle.")
				return fiber.NewError(fiber.StatusConflict, "AutoPi connected to another vehicle.")
			}
		}
	}

	udai, err := ud.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(autoPiInt.Id),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.AutopiUnit, models.AutopiUnitRels.PairRequest)),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.AutopiUnit, models.AutopiUnitRels.UnpairRequest)),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Err(err).Msg("Database failure searching for device's AutoPi integration.")
			return opaqueInternalError
		}
	} else {
		// Conflict with web2 pairing?
		if autoPiUnit != nil && (!udai.AutopiUnitID.Valid || udai.AutopiUnitID.String != autoPiUnit.AutopiUnitID) {
			return fiber.NewError(fiber.StatusConflict, "Vehicle already paired with another AutoPi.")
		}

		if !udai.AutopiUnitID.Valid {
			return opaqueInternalError
		}

		autoPiUnit = udai.R.AutopiUnit
	}

	if autoPiUnit.R.PairRequest != nil && autoPiUnit.R.PairRequest.Status != "Failed" {
		if autoPiUnit.R.PairRequest.Status == models.MetaTransactionRequestStatusConfirmed {
			return fiber.NewError(fiber.StatusConflict, "AutoPi already paired.")
		}
		return fiber.NewError(fiber.StatusConflict, "AutoPi pairing in process.")
	}

	if autoPiUnit.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "AutoPi not yet minted.")
	}

	if ud.R.VehicleNFT == nil || ud.R.VehicleNFT.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not yet minted.")
	}

	if !autoPiUnit.OwnerAddress.Valid {
		return fiber.NewError(fiber.StatusConflict, "Device not yet claimed.")
	}

	if common.BytesToAddress(autoPiUnit.OwnerAddress.Bytes) != common.BytesToAddress(ud.R.VehicleNFT.OwnerAddress.Bytes) {
		return fiber.NewError(fiber.StatusConflict, "AutoPi and vehicle have different owners.")
	}

	apToken := autoPiUnit.TokenID.Int(nil)
	vehicleToken := ud.R.VehicleNFT.TokenID.Int(nil)

	// TODO(elffjs): Really shouldn't be dialing so much.
	conn, err := grpc.Dial(udc.Settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		udc.log.Err(err).Msg("Failed to create users API client.")
		return opaqueInternalError
	}
	defer conn.Close()

	usersClient := pb.NewUserServiceClient(conn)

	user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Failed to retrieve user information.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusConflict, "User does not have an Ethereum address.")
	}

	if common.HexToAddress(*user.EthereumAddress) != common.BytesToAddress(autoPiUnit.OwnerAddress.Bytes) {
		return fiber.NewError(fiber.StatusConflict, "AutoPi claimed by another user.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	pads := registry.PairAftermarketDeviceSign{
		AftermarketDeviceNode: apToken,
		VehicleNode:           vehicleToken,
	}

	realAddr := common.HexToAddress(*user.EthereumAddress)

	hash, err := client.Hash(&pads)
	if err != nil {
		return err
	}

	sigBytes := common.FromHex(pairReq.Signature)

	if len(sigBytes) != 65 {
		logger.Error().Str("rawSignature", pairReq.Signature).Msg("Signature was not 65 bytes.")
		return fiber.NewError(fiber.StatusBadRequest, "Signature was not 65 bytes long.")
	}

	recAddr, err := recoverAddress2(hash[:], sigBytes)
	if err != nil {
		return err
	}

	if recAddr != realAddr {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signature.")
	}

	requestID := ksuid.New().String()

	mtr := models.MetaTransactionRequest{
		ID:     requestID,
		Status: models.MetaTransactionRequestStatusUnsubmitted,
	}
	err = mtr.Insert(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return err
	}

	autoPiUnit.UnpairRequestID = null.String{}
	autoPiUnit.PairRequestID = null.StringFrom(requestID)
	_, err = autoPiUnit.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return err
	}

	err = client.PairAftermarketDeviceSign(requestID, apToken, vehicleToken, sigBytes)
	if err != nil {
		return err
	}

	return nil
}

// CloudRepairAutoPi godoc
// @Description Re-apply AutoPi cloud actions in an attempt to get the device transmitting data again.
// @Produce json
// @Param userDeviceID path string true "Device id"
// @Success 204
// @Security BearerAuth
// @Router /user/devices/:userDeviceID/autopi/commands/cloud-repair [post]
func (udc *UserDevicesController) CloudRepairAutoPi(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	userDeviceID := c.Params("userDeviceID")

	logger := udc.log.With().Str("userId", userID).Str("userDeviceId", userDeviceID).Logger()
	logger.Info().Msg("Got AutoPi pair request.")

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenAutopiUnit)),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id found.")
		}
		logger.Err(err).Msg("Database failure searching for device.")
		return opaqueInternalError
	}

	if ud.R.VehicleNFT == nil || ud.R.VehicleNFT.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not yet minted.")
	}

	if ud.R.VehicleNFT.R.VehicleTokenAutopiUnit == nil {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not paired on-chain with any AutoPi.")
	}

	vehicleID := ud.R.VehicleNFT.TokenID.Int(nil)
	autoPiID := ud.R.VehicleNFT.R.VehicleTokenAutopiUnit.TokenID.Int(nil)

	err = udc.autoPiIntegration.Pair(c.Context(), autoPiID, vehicleID)
	if err != nil {
		return err
	}

	return c.SendStatus(204)
}

// UnairAutoPi godoc
// @Description Submit the signature for unpairing this device from its attached AutoPi.
// @Produce json
// @Param userDeviceID path string true "Device id"
// @Param userSignature body controllers.AutoPiPairRequest true "User signature."
// @Security BearerAuth
// @Router /user/devices/:userDeviceID/autopi/commands/unpair [post]
func (udc *UserDevicesController) UnpairAutoPi(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	// Make sure we have an Ethereum address.
	// TODO(elffjs): Really shouldn't be dialing so much. Do we even need to do this? We have
	// the owner's address.
	conn, err := grpc.Dial(udc.Settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		udc.log.Err(err).Msg("Failed to create users API client.")
		return opaqueInternalError
	}
	defer conn.Close()

	usersClient := pb.NewUserServiceClient(conn)

	user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Failed to retrieve user information.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusConflict, "User does not have an Ethereum address.")
	}

	realAddr := common.HexToAddress(*user.EthereumAddress)

	userDeviceID := c.Params("userDeviceID")

	logger := udc.log.With().Str("userId", userID).Str("userDeviceId", userDeviceID).Logger()
	logger.Info().Msg("Got AutoPi unpair request.")

	// TODO(elffjs): Is SELECT ... FOR UPDATE better here?
	tx, err := udc.DBS().Writer.BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenAutopiUnit)),
	).One(c.Context(), tx)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id found.")
		}
		return err
	}

	vnft := ud.R.VehicleNFT

	if vnft == nil || vnft.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not yet minted.")
	}

	if !vnft.OwnerAddress.Valid {
		logger.Error().Msg("Vehicle minted but has no owner.")
		return opaqueInternalError
	}

	if owner := common.BytesToAddress(vnft.OwnerAddress.Bytes); owner != realAddr {
		logger.Error().Str("ownerAddress", owner.Hex()).Str("userAddress", realAddr.Hex()).Msg("Vehicle owner and user Ethereum address no longer match.")
		return opaqueInternalError
	}

	apnft := vnft.R.VehicleTokenAutopiUnit

	if apnft == nil {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not paired to an AutoPi on-chain.")
	}

	if apnft.UnpairRequestID.Valid {
		// If unpairing had finished, we wouldn't have a link from the vehicle NFT.
		return fiber.NewError(fiber.StatusConflict, "Unpairing already in progress.")
	}

	vehicleToken := vnft.TokenID.Int(nil)
	apToken := apnft.TokenID.Int(nil)

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	uads := registry.UnPairAftermarketDeviceSign{
		AftermarketDeviceNode: apToken,
		VehicleNode:           vehicleToken,
	}

	var pairReq AutoPiPairRequest
	err = c.BodyParser(&pairReq)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	hash, err := client.Hash(&uads)
	if err != nil {
		return err
	}

	sigBytes := common.FromHex(pairReq.Signature)

	recAddr, err := recoverAddress2(hash[:], sigBytes)
	if err != nil {
		return err
	}

	if recAddr != realAddr {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signature.")
	}

	requestID := ksuid.New().String()

	mtr := models.MetaTransactionRequest{
		ID:     requestID,
		Status: models.MetaTransactionRequestStatusUnsubmitted,
	}
	err = mtr.Insert(c.Context(), tx, boil.Infer())
	if err != nil {
		return err
	}

	apnft.UnpairRequestID = null.StringFrom(requestID)
	_, err = apnft.Update(c.Context(), tx, boil.Whitelist(models.AutopiUnitColumns.UnpairRequestID))
	if err != nil {
		return err
	}

	// This is a bit iffy, since we don't want to save this record and then fail to send to Kafka.
	// But the opposite is worse, I think.
	err = tx.Commit()
	if err != nil {
		return err
	}

	return client.UnPairAftermarketDeviceSign(requestID, apToken, vehicleToken, sigBytes)
}

// GetAutoPiUnpairMessage godoc
// @Description Return the EIP-712 payload to be signed for AutoPi device unpairing.
// @Produce json
// @Param userDeviceID path string true "Device id"
// @Success 200 {object} signer.TypedData
// @Security BearerAuth
// @Router /user/devices/:userDeviceID/autopi/commands/unpair [get]
func (udc *UserDevicesController) GetAutoPiUnpairMessage(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	userDeviceID := c.Params("userDeviceID")

	logger := udc.log.With().Str("userId", userID).Str("userDeviceId", userDeviceID).Logger()
	logger.Info().Msg("Got AutoPi pair request.")

	autoPiInt, err := udc.DeviceDefIntSvc.GetAutoPiIntegration(c.Context())
	if err != nil {
		logger.Err(err).Msg("Failed to retrieve AutoPi integration.")
		return helpers.GrpcErrorToFiber(err, "failed to retrieve AutoPi integration.")
	}

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.VehicleNFT),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id found.")
		}
		logger.Err(err).Msg("Database failure searching for device.")
		return opaqueInternalError
	}

	if ud.UserID != userID {
		// Err on the side of privacy.
		return fiber.NewError(fiber.StatusNotFound, "No device with that id found.")
	}

	udai, err := ud.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(autoPiInt.Id)).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusConflict, "Device does not have an AutoPi associated.")
		}
		logger.Err(err).Msg("Database failure searching for device's AutoPi integration.")
		return opaqueInternalError
	}

	if !udai.AutopiUnitID.Valid {
		// This shouldn't happen.
		logger.Error().Msg("Active AutoPi integration with no associated unit id.")
		return opaqueInternalError
	}

	autoPiUnit, err := udai.AutopiUnit().One(c.Context(), udc.DBS().Reader)
	if err != nil {
		logger.Error().Msg("Failed to retrieve AutoPi record.")
		return opaqueInternalError
	}

	if autoPiUnit.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "AutoPi not yet minted.")
	}

	if ud.R.VehicleNFT == nil || ud.R.VehicleNFT.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not yet minted.")
	}

	apToken := autoPiUnit.TokenID.Int(nil)
	vehicleToken := ud.R.VehicleNFT.TokenID.Int(nil)

	// TODO(elffjs): Really shouldn't be dialing so much.
	conn, err := grpc.Dial(udc.Settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		udc.log.Err(err).Msg("Failed to create users API client.")
		return opaqueInternalError
	}
	defer conn.Close()

	usersClient := pb.NewUserServiceClient(conn)

	user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Failed to retrieve user information.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusConflict, "User does not have an Ethereum address.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	uads := &registry.UnPairAftermarketDeviceSign{
		AftermarketDeviceNode: apToken,
		VehicleNode:           vehicleToken,
	}

	var out *signer.TypedData = client.GetPayload(uads)

	return c.JSON(out)
}

type AutoPiClaimRequest struct {
	// UserSignature is the signature from the user, using their private key.
	UserSignature string `json:"userSignature"`
	// AftermarketDeviceSignature is the signature from the aftermarket device.
	AftermarketDeviceSignature string `json:"aftermarketDeviceSignature"`
}

type AutoPiPairRequest struct {
	ExternalID string `json:"externalId"`
	Signature  string `json:"signature"`
}

// PostClaimAutoPi godoc
// @Description Return the EIP-712 payload to be signed for AutoPi device claiming.
// @Produce json
// @Param unitID path string true "AutoPi unit id"
// @Param claimRequest body controllers.AutoPiClaimRequest true "Signatures from the user and AutoPi"
// @Success 204
// @Security BearerAuth
// @Router /autopi/unit/:unitID/commands/claim [post]
func (udc *UserDevicesController) PostClaimAutoPi(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	unitID := c.Params("unitID")

	logger := udc.log.With().Str("userId", userID).Str("autoPiUnitId", unitID).Str("route", c.Route().Name).Logger()

	reqBody := AutoPiClaimRequest{}
	err := c.BodyParser(&reqBody)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	udc.log.Info().Interface("payload", reqBody).Msg("Got claim request.")

	unit, err := models.AutopiUnits(
		models.AutopiUnitWhere.AutopiUnitID.EQ(unitID),
		qm.Load(models.AutopiUnitRels.ClaimMetaTransactionRequest),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "AutoPi not minted, or unit ID invalid.")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Internal error.")
	}

	if unit.UserID.Valid && unit.UserID.String != userID {
		return fiber.NewError(fiber.StatusForbidden, "AutoPi paired to another user.")
	}

	if unit.TokenID.IsZero() || !unit.EthereumAddress.Valid {
		return fiber.NewError(fiber.StatusNotFound, "AutoPi not minted.")
	}

	if unit.OwnerAddress.Valid {
		return fiber.NewError(fiber.StatusConflict, "Device already claimed.")
	}

	if unit.R.ClaimMetaTransactionRequest != nil && unit.R.ClaimMetaTransactionRequest.Status != "Failed" {
		return fiber.NewError(fiber.StatusConflict, "Claiming transaction in progress.")
	}

	apToken := unit.TokenID.Int(nil)

	conn, err := grpc.Dial(udc.Settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		udc.log.Err(err).Msg("Failed to create users API client.")
		return opaqueInternalError
	}
	defer conn.Close()

	usersClient := pb.NewUserServiceClient(conn)

	user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Couldn't retrieve user record.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusBadRequest, "User does not have an Ethereum address.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	realUserAddr := common.HexToAddress(*user.EthereumAddress)

	cads := &registry.ClaimAftermarketDeviceSign{
		AftermarketDeviceNode: apToken,
		Owner:                 realUserAddr,
	}

	hash, err := client.Hash(cads)
	if err != nil {
		return err
	}

	userSig := common.FromHex(reqBody.UserSignature)

	if len(userSig) != 65 {
		logger.Error().Str("rawSignature", reqBody.UserSignature).Msg("User signature was not 65 bytes.")
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("User signature has invalid length %d.", len(userSig)))
	}

	recUserAddr, err := recoverAddress2(hash[:], userSig)
	if err != nil {
		return err
	}

	if recUserAddr != realUserAddr {
		return fiber.NewError(fiber.StatusBadRequest, "User signature invalid.")
	}

	amSig := common.FromHex(reqBody.AftermarketDeviceSignature)

	if len(amSig) != 65 {
		logger.Error().Str("rawSignature", reqBody.AftermarketDeviceSignature).Msg("Device signature was not 65 bytes.")
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Device signature has invalid length %d.", len(amSig)))
	}

	recAmAddr, err := recoverAddress2(hash[:], amSig)
	if err != nil {
		return err
	}

	realAmAddr := common.BytesToAddress(unit.EthereumAddress.Bytes)

	if recAmAddr != realAmAddr {
		return fiber.NewError(fiber.StatusBadRequest, "Aftermarket device signature invalid.")
	}

	requestID := ksuid.New().String()

	mtr := models.MetaTransactionRequest{
		ID:     requestID,
		Status: models.MetaTransactionRequestStatusUnsubmitted,
	}
	err = mtr.Insert(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return err
	}

	unit.UserID = null.StringFrom(userID)
	unit.ClaimMetaTransactionRequestID = null.StringFrom(requestID)
	_, err = unit.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return err
	}

	return client.ClaimAftermarketDeviceSign(requestID, apToken, realUserAddr, userSig, amSig)
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
		models.UserDeviceWhere.UserID.EQ(userID),
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

	dds, err := udc.DeviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), []string{ud.DeviceDefinitionID})
	if err != nil {
		logger.Err(err).Msg("grpc error searching for device definition")
		return helpers.GrpcErrorToFiber(err, "failed to get device definition with id: "+ud.DeviceDefinitionID)
	}
	dd := dds[0]
	logger.Info().Msgf("get device definition id result during registration %+v", dd)

	// filter out the desired integration from the compatible ones
	var deviceInteg *ddgrpc.Integration
	for _, integration := range dd.DeviceIntegrations {
		if integration.Integration.Id == integrationID {
			deviceInteg = &ddgrpc.Integration{
				Id:     integration.Integration.Id,
				Type:   integration.Integration.Type,
				Style:  integration.Integration.Style,
				Vendor: integration.Integration.Vendor,
			}
			break
		}
	}

	if deviceInteg == nil {
		// todo need a test for this
		return fiber.NewError(fiber.StatusBadRequest,
			fmt.Sprintf("deviceDefinitionId %s does not support integrationId %s for region %s", ud.DeviceDefinitionID, integrationID, countryRecord.Region))
	}

	if exists, err := models.UserDeviceAPIIntegrationExists(c.Context(), tx, userDeviceID, integrationID); err != nil {
		logger.Err(err).Msg("Unexpected database error looking for existing instance of integration")
		return err
	} else if exists {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("userDeviceId %s already has a user_device_api_integration with integrationId %s, please delete that first", userDeviceID, integrationID))
	}

	var regErr error
	// The per-integration handler is responsible for handling the fiber context and committing the
	// transaction.
	switch vendor := deviceInteg.Vendor; vendor {
	case constants.SmartCarVendor:
		regErr = udc.registerSmartcarIntegration(c, &logger, tx, deviceInteg, ud)
	case constants.TeslaVendor:
		regErr = udc.registerDeviceTesla(c, &logger, tx, userDeviceID, deviceInteg, ud)
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

	udc.runPostRegistration(c.Context(), &logger, userDeviceID, integrationID, deviceInteg)

	return nil
}

// RegisterDeviceIntegration godoc
// @Description Submit credentials for registering a device with a given integration.
// @Tags        integrations
// @Accept      json
// @Param       userDeviceIntegrationRegistration body controllers.RegisterDeviceIntegrationRequest true "Integration credentials"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/:userDeviceID/integrations/:integrationID [post]
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
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice)),
	).One(ctx, udc.DBS().Reader)
	if err != nil {
		logger.Err(err).Msg("Couldn't retrieve UDAI for post-registration tasks.")
		return
	}

	ud := udai.R.UserDevice
	// pull dd info again - don't pass it in, as it may have changed
	dds, err2 := udc.DeviceDefSvc.GetDeviceDefinitionsByIDs(ctx, []string{ud.DeviceDefinitionID})
	if err2 != nil {
		logger.Err(err2).Str("deviceDefinitionId", ud.DeviceDefinitionID).Msg("failed to retrieve device defintion")
	}
	dd := dds[0]

	err = udc.eventService.Emit(
		&services.Event{
			Type:    "com.dimo.zone.device.integration.create",
			Source:  "devices-api",
			Subject: userDeviceID,
			Data: services.UserDeviceIntegrationEvent{
				Timestamp: time.Now(),
				UserID:    ud.UserID,
				Device: services.UserDeviceEventDevice{
					ID:                 userDeviceID,
					DeviceDefinitionID: dd.DeviceDefinitionId,
					Make:               dd.Type.Make,
					Model:              dd.Type.Model,
					Year:               int(dd.Type.Year),
					VIN:                ud.VinIdentifier.String,
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

	region := ""
	if ud.CountryCode.Valid {
		countryRecord := constants.FindCountry(ud.CountryCode.String)
		if countryRecord != nil {
			region = countryRecord.Region
		}
	}
	err = udc.deviceDefinitionRegistrar.Register(services.DeviceDefinitionDTO{
		IntegrationID:      integ.Id,
		UserDeviceID:       ud.ID,
		DeviceDefinitionID: ud.DeviceDefinitionID,
		Make:               dd.Type.Make,
		Model:              dd.Type.Model,
		Year:               int(dd.Type.Year),
		Region:             region,
		MakeSlug:           dd.Type.MakeSlug,
		ModelSlug:          dd.Type.ModelSlug,
	})
	if err != nil {
		logger.Err(err).Msg("Failed to set values in device definition tables.")
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
			logger.Err(err).Msg("Failed to exchange authorization code with Smartcar.")
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

	// Prevent users from connecting a vehicle if it's already connected through another user
	// device object. Disabled outside of prod for ease of testing.
	if udc.Settings.IsProduction() {
		// Probably a race condition here. Need to either lock something or impose a greater
		// isolation level.
		conflict, err := models.UserDevices(
			models.UserDeviceWhere.ID.NEQ(ud.ID), // If you want to re-register, or register a different integration, that's okay.
			models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(vin)),
			models.UserDeviceWhere.VinConfirmed.EQ(true),
		).Exists(c.Context(), tx)
		if err != nil {
			logger.Err(err).Msg("Failed to search for VIN conflicts.")
			return opaqueInternalError
		}

		if conflict {
			logger.Error().Msg("VIN %s already in use.")
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("VIN %s in use by a previously connected device.", ud.VinIdentifier.String))
		}
	}

	year, err := udc.smartcarClient.GetYear(c.Context(), token.Access, externalID)
	if err != nil {
		return smartcarCallErr
	}

	if err := udc.fixSmartcarDeviceYear(c.Context(), logger, tx, integ, ud, year); err != nil {
		logger.Err(err).Msg("Failed to correct Smartcar device definition year.")
	}

	endpoints, err := udc.smartcarClient.GetEndpoints(c.Context(), token.Access, externalID)
	if err != nil {
		return smartcarCallErr
	}

	var commands *services.UserDeviceAPIIntegrationsMetadataCommands

	doorControl, err := udc.smartcarClient.HasDoorControl(c.Context(), token.Access, externalID)
	if err != nil {
		return smartcarCallErr
	}

	if doorControl {
		commands = &services.UserDeviceAPIIntegrationsMetadataCommands{
			Enabled: []string{"doors/unlock", "doors/lock"},
		}
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

	taskID := ksuid.New().String()

	integration := &models.UserDeviceAPIIntegration{
		TaskID:          null.StringFrom(taskID),
		ExternalID:      null.StringFrom(externalID),
		UserDeviceID:    ud.ID,
		IntegrationID:   integ.Id,
		Status:          models.UserDeviceAPIIntegrationStatusPendingFirstData,
		AccessToken:     null.StringFrom(encAccess),
		AccessExpiresAt: null.TimeFrom(token.AccessExpiry),
		RefreshToken:    null.StringFrom(encRefresh),
		Metadata:        null.JSONFrom(b),
	}

	if err := integration.Insert(c.Context(), tx, boil.Infer()); err != nil {
		logger.Err(err).Msg("Unexpected database error inserting new Smartcar integration registration.")
		return opaqueInternalError
	}

	if !ud.VinConfirmed {
		ud.VinIdentifier = null.StringFrom(strings.ToUpper(vin))
		ud.VinConfirmed = true
		_, err = ud.Update(c.Context(), tx, boil.Infer())
		if err != nil {
			return opaqueInternalError
		}
	}

	if err := udc.smartcarTaskSvc.StartPoll(integration); err != nil {
		logger.Err(err).Msg("Couldn't start Smartcar polling.")
		return opaqueInternalError
	}

	if err := tx.Commit(); err != nil {
		logger.Error().Msg("Failed to commit new user device integration.")
		return opaqueInternalError
	}

	logger.Info().Msg("Finished Smartcar device registration.")

	// fire off task to get drivly data
	taskID, err = udc.drivlyTaskService.StartDrivlyUpdate(ud.DeviceDefinitionID, ud.ID, vin)
	if err != nil {
		logger.Err(err).Msg("Failed to emit task drivly event task.")
	}

	logger.Info().Msgf("drivly update task ID = %s", taskID)

	return c.SendStatus(fiber.StatusNoContent)
}

// nolint
func (udc *UserDevicesController) AdminVehicleDeviceLink(c *fiber.Ctx) error {
	return nil
}

type web3UnclaimDevice struct {
	AftermarketDeviceNode *big.Int `json:"aftermarketDeviceNode"`
	AutoPiUnitID          string   `json:"autoPiUnitId"`
}

func (udc *UserDevicesController) AdminDeviceWeb3Unclaim(c *fiber.Ctx) error {
	wud := web3UnclaimDevice{}
	err := c.BodyParser(&wud)
	if err != nil {
		return err
	}

	type requestData struct {
		ID   string `json:"id"`
		To   string `json:"to"`
		Data string `json:"data"`
	}

	reqID := ksuid.New().String()

	node := wud.AftermarketDeviceNode

	var unit *models.AutopiUnit

	if wud.AutoPiUnitID != "" {
		unit, err = models.FindAutopiUnit(c.Context(), udc.DBS().Reader, wud.AutoPiUnitID)
		if err != nil {
			return err
		}

		node = unit.TokenID.Int(nil)
	} else {
		unit, err = models.AutopiUnits(
			models.AutopiUnitWhere.TokenID.EQ(types.NewNullDecimal(new(decimal.Big).SetBigMantScale(node, 0))),
		).One(c.Context(), udc.DBS().Reader)
		if err != nil {
			return err
		}
	}

	unit.ClaimMetaTransactionRequestID = null.String{}
	unit.OwnerAddress = null.Bytes{}
	_, err = unit.Update(c.Context(), udc.DBS().Reader, boil.Infer())
	if err != nil {
		return err
	}

	abi, err := contracts.RegistryMetaData.GetAbi()
	if err != nil {
		return err
	}

	data, err := abi.Pack("unclaimAftermarketDeviceNode", []*big.Int{node})
	if err != nil {
		return err
	}

	addr := common.HexToAddress(udc.Settings.DIMORegistryAddr)
	event := shared.CloudEvent[requestData]{
		ID:          ksuid.New().String(),
		Source:      "devices-api",
		SpecVersion: "1.0",
		Subject:     reqID,
		Time:        time.Now(),
		Type:        "zone.dimo.transaction.request",
		Data: requestData{
			ID:   reqID,
			To:   hexutil.Encode(addr[:]),
			Data: hexutil.Encode(data),
		},
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, _, err = udc.producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: "topic.transaction.request.send",
			Key:   sarama.StringEncoder(reqID),
			Value: sarama.ByteEncoder(eventBytes),
		},
	)

	return err
}

func (udc *UserDevicesController) AdminDeviceWeb3Unpair(c *fiber.Ctx) error {
	wud := web3UnclaimDevice{}
	err := c.BodyParser(&wud)
	if err != nil {
		return err
	}

	type requestData struct {
		ID   string `json:"id"`
		To   string `json:"to"`
		Data string `json:"data"`
	}

	reqID := ksuid.New().String()

	node := wud.AftermarketDeviceNode

	var unit *models.AutopiUnit

	if wud.AutoPiUnitID != "" {
		unit, err = models.FindAutopiUnit(c.Context(), udc.DBS().Reader, wud.AutoPiUnitID)
		if err != nil {
			return err
		}

		node = unit.TokenID.Int(nil)
	} else {
		unit, err = models.AutopiUnits(
			models.AutopiUnitWhere.TokenID.EQ(types.NewNullDecimal(new(decimal.Big).SetBigMantScale(node, 0))),
		).One(c.Context(), udc.DBS().Reader)
		if err != nil {
			return err
		}
	}

	unit.PairRequestID = null.String{}
	_, err = unit.Update(c.Context(), udc.DBS().Reader, boil.Infer())
	if err != nil {
		return err
	}

	abi, err := contracts.RegistryMetaData.GetAbi()
	if err != nil {
		return err
	}

	data, err := abi.Pack("unpairAftermarketDeviceByDeviceNode", []*big.Int{node})
	if err != nil {
		return err
	}

	addr := common.HexToAddress(udc.Settings.DIMORegistryAddr)
	event := shared.CloudEvent[requestData]{
		ID:          ksuid.New().String(),
		Source:      "devices-api",
		SpecVersion: "1.0",
		Subject:     reqID,
		Time:        time.Now(),
		Type:        "zone.dimo.transaction.request",
		Data: requestData{
			ID:   reqID,
			To:   hexutil.Encode(addr[:]),
			Data: hexutil.Encode(data),
		},
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, _, err = udc.producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: "topic.transaction.request.send",
			Key:   sarama.StringEncoder(reqID),
			Value: sarama.ByteEncoder(eventBytes),
		},
	)

	return err
}

func (udc *UserDevicesController) registerDeviceTesla(c *fiber.Ctx, logger *zerolog.Logger, tx *sql.Tx, userDeviceID string, integ *ddgrpc.Integration, ud *models.UserDevice) error {
	reqBody := new(RegisterDeviceIntegrationRequest)
	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	// We'll use this to kick off the job
	teslaID, err := strconv.Atoi(reqBody.ExternalID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse external id %q as an integer.", teslaID))
	}

	v, err := udc.teslaService.GetVehicle(reqBody.AccessToken, teslaID)
	if err != nil {
		logger.Err(err).Msg("Error on initial Tesla call.")
		// TODO(elffjs): 400 may not be entirely accurate.
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't retrieve vehicle from Tesla.")
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
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("VIN %s in use by another vehicle.", v.VIN))
		}
	}

	if err := fixTeslaDeviceDefinition(c.Context(), logger, udc.DeviceDefSvc, tx, integ, ud, v.VIN); err != nil {
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

	// TODO(elffjs): Stupid to marshal this again and again.
	meta := services.UserDeviceAPIIntegrationsMetadata{
		Commands: &services.UserDeviceAPIIntegrationsMetadataCommands{
			Enabled: []string{"doors/unlock", "doors/lock", "trunk/open", "frunk/open", "charge/limit"},
		},
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

	if err := udc.teslaService.WakeUpVehicle(reqBody.AccessToken, teslaID); err != nil {
		logger.Err(err).Msg("Couldn't wake up Tesla.")
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
func fixTeslaDeviceDefinition(ctx context.Context, logger *zerolog.Logger, ddSvc services.DeviceDefinitionService, exec boil.ContextExecutor, _ *ddgrpc.Integration, ud *models.UserDevice, vin string) error {
	vinMake := "Tesla"
	vinModel := shared.VIN(vin).TeslaModel()
	vinYear := shared.VIN(vin).Year()
	mmy, err := ddSvc.FindDeviceDefinitionByMMY(ctx, vinMake, vinModel, vinYear)

	if err != nil {
		return err
	}

	if mmy.DeviceDefinitionId != ud.DeviceDefinitionID {
		logger.Warn().Msgf(
			"Device moving to new device definition from %s to %s", ud.DeviceDefinitionID, mmy.DeviceDefinitionId,
		)
		ud.DeviceDefinitionID = mmy.DeviceDefinitionId
		_, err = ud.Update(ctx, exec, boil.Infer())
		if err != nil {
			return err
		}
	}

	return nil
}

// fixSmartcarDeviceYear tries to use the MMY provided by Smartcar to at least correct the year of
// the device definition used by the device.
//
// We do not attempt to create any new entries in integrations, device_definitions, or
// device_integrations. This seems too dangerous to me.
func (udc *UserDevicesController) fixSmartcarDeviceYear(ctx context.Context, logger *zerolog.Logger, _ boil.ContextExecutor, integ *ddgrpc.Integration, ud *models.UserDevice, year int) error {

	deviceDefinitionResponse, err := udc.DeviceDefSvc.GetDeviceDefinitionsByIDs(ctx, []string{ud.DeviceDefinitionID})

	if err != nil {
		return helpers.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+ud.DeviceDefinitionID)
	}

	dd := deviceDefinitionResponse[0]

	if int(dd.Type.Year) != year {
		logger.Warn().Msgf("Device was attached to year %d but should be %d.", dd.Type.Year, year)
		region := ""
		if countryRecord := constants.FindCountry(ud.CountryCode.String); countryRecord != nil {
			region = countryRecord.Region
		}
		// todo gprc pull by MMY from from device-defintions
		newDD, err := udc.DeviceDefSvc.FindDeviceDefinitionByMMY(ctx, dd.Make.Name, dd.Type.Model, year)

		if err != nil {
			return fmt.Errorf("grpc error: %w", err)
		}

		if newDD == nil {
			return fmt.Errorf("no device definition %s, %s, %d", dd.Make.Name, dd.Type.Model, year)
		}

		if len(newDD.DeviceIntegrations) == 0 {
			return fmt.Errorf("correct device definition %s has no integration %s for region %s", newDD.DeviceDefinitionId, integ.Id, region)
		}

		// todo: validate with james
		//if err := ud.SetDeviceDefinition(ctx, exec, false, newDD); err != nil {
		//	return fmt.Errorf("failed switching device definition to %s: %w", newDD.DeviceDefinitionID, err)
		//}
	}

	return nil
}

/** Structs for request / response **/

type UserDeviceIntegrationStatus struct {
	IntegrationID     string    `json:"integrationId"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"createdAt"`
	ExternalID        *string   `json:"externalId"`
	UpdatedAt         time.Time `json:"updatedAt"`
	Metadata          null.JSON `json:"metadata" swaggertype:"string"`
	IntegrationVendor string    `json:"integrationVendor"`
}

// RegisterDeviceIntegrationRequest carries credentials used to connect the device to a given
// integration.
type RegisterDeviceIntegrationRequest struct {
	// Code is an OAuth authorization code. Not used in all integrations.
	Code string `json:"code"`
	// RedirectURI is the OAuth redirect URI used by the frontend. Not used in all integrations.
	RedirectURI string `json:"redirectURI"`
	// ExternalID is the only field needed for AutoPi registrations. It is the UnitID.
	ExternalID   string `json:"externalId"`
	AccessToken  string `json:"accessToken"`
	ExpiresIn    int    `json:"expiresIn"`
	RefreshToken string `json:"refreshToken"`
}

type GetUserDeviceIntegrationResponse struct {
	// Status is one of "Pending", "PendingFirstData", "Active", "Failed", "DuplicateIntegration".
	Status string `json:"status"`
	// ExternalID is the identifier used by the third party for the device. It may be absent if we
	// haven't authorized yet.
	ExternalID null.String `json:"externalId" swaggertype:"string"`
	// CreatedAt is the creation time of this integration for this device.
	CreatedAt time.Time `json:"createdAt"`
}

type AutoPiCommandRequest struct {
	Command string `json:"command"`
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

	TokenID         *big.Int        `json:"tokenId,omitempty"`
	EthereumAddress *common.Address `json:"ethereumAddress,omitempty"`
	OwnerAddress    *common.Address `json:"ownerAddress,omitempty"`

	// Claim contains the status of the on-chain claiming meta-transaction.
	Claim *AutoPiTransactionStatus `json:"claim,omitempty"`
	// Pair contains the status of the on-chain pairing meta-transaction.
	Pair *AutoPiTransactionStatus `json:"pair,omitempty"`
	// Unpair contains the status of the on-chain unpairing meta-transaction.
	Unpair *AutoPiTransactionStatus `json:"unpair,omitempty"`
}

// AutoPiTransactionStatus summarizes the state of an on-chain AutoPi operation.
type AutoPiTransactionStatus struct {
	// Status is the state of the transaction performing this operation. There are only four options.
	Status string `json:"status" enums:"Unsubmitted,Submitted,Mined,Confirmed" example:"Mined"`
	// Hash is the hexidecimal transaction hash, available for any transaction at the Submitted stage or greater.
	Hash *string `json:"hash,omitempty" example:"0x28b4662f1e1b15083261a4a5077664f4003d58cb528826b7aab7fad466c28e70"`
	// CreatedAt is the timestamp of the creation of the meta-transaction.
	CreatedAt time.Time `json:"createdAt" example:"2022-10-01T09:22:21.002Z"`
	// UpdatedAt is the last time we updated the status of the transaction.
	UpdatedAt time.Time `json:"updatedAt" example:"2022-10-01T09:22:26.337Z"`
}
