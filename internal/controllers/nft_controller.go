package controllers

import (
	"database/sql"
	"fmt"
	"math/big"
	"slices"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/utils"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	grpcfiber "github.com/DIMO-Network/shared/pkg/grpcfiber"
	vinutil "github.com/DIMO-Network/shared/pkg/vin"
	pb_oracle "github.com/DIMO-Network/tesla-oracle/pkg/grpc"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type NFTController struct {
	Settings         *config.Settings
	DBS              func() *db.ReaderWriter
	s3               *s3.Client
	log              *zerolog.Logger
	deviceDefSvc     services.DeviceDefinitionService
	integSvc         services.DeviceDefinitionIntegrationService
	teslaTaskService services.TeslaTaskService
	oracleClient     pb_oracle.TeslaOracleClient
}

// NewNFTController constructor
func NewNFTController(settings *config.Settings, dbs func() *db.ReaderWriter, logger *zerolog.Logger, s3 *s3.Client,
	deviceDefSvc services.DeviceDefinitionService,
	teslaTaskService services.TeslaTaskService,
	integSvc services.DeviceDefinitionIntegrationService,
	oracleClient pb_oracle.TeslaOracleClient,
) NFTController {
	return NFTController{
		Settings:         settings,
		DBS:              dbs,
		log:              logger,
		s3:               s3,
		deviceDefSvc:     deviceDefSvc,
		teslaTaskService: teslaTaskService,
		integSvc:         integSvc,
		oracleClient:     oracleClient,
	}
}

func validVINChar(r rune) bool {
	return 'A' <= r && r <= 'Z' || '0' <= r && r <= '9'
}

// UpdateVINV2 godoc
// @Description updates the VIN on the user device record. Can optionally also update the protocol and the country code.
// VIN now comes from attestations, no need for this soon.
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       tokenId path int true "token id"
// @Param       vin body controllers.UpdateVINReq true "VIN"
// @Success     204
// @Security    BearerAuth
// @Router      /vehicle/{tokenId}/vin [patch]
func (udc *UserDevicesController) UpdateVINV2(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	tokenID, ok := new(big.Int).SetString(tis, 10)
	userEthAddr, _ := helpers.GetJWTEthAddr(c) // we use privilege auth - need to test in dev this would still work
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}
	logger := helpers.GetLogger(c, udc.log).With().Str("route", c.Route().Name).Logger()

	var req UpdateVINReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Could not parse request body.")
	}

	req.VIN = strings.TrimSpace(strings.ToUpper(req.VIN))

	vin := vinutil.VIN(req.VIN)
	if !vin.IsValidVIN() {
		if !vin.IsValidJapanChassis() {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid VIN.")
		}
	}

	// Don't want phantom reads.
	tx, err := udc.DBS().GetWriterConn().BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return opaqueInternalError
	}
	defer tx.Rollback() //nolint

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(utils.NullableBigToDecimal(tokenID)),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Vehicle NFT %d not found.", tokenID))
		}
		logger.Err(err).Msg("Failed to search for device.")
		return opaqueInternalError
	}

	// no update if the same
	if userDevice.VinIdentifier.String == req.VIN {
		return c.SendStatus(fiber.StatusNoContent)
	}

	// If signed, we should be able to set the VIN to validated.
	if req.Signature != "" {
		vinByte := []byte(req.VIN)
		sig := common.FromHex(req.Signature)
		if len(sig) != 65 {
			logger.Error().Str("rawSignature", req.Signature).Msg("Signature was not 65 bytes.")
			return fiber.NewError(fiber.StatusBadRequest, "Signature is not 65 bytes long.")
		}

		hash := crypto.Keccak256(vinByte)

		recAddr, err := helpers.Ecrecover(hash, sig)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Couldn't recover signer address.")
		}

		found, err := models.AftermarketDevices(
			models.AftermarketDeviceWhere.EthereumAddress.EQ(recAddr.Bytes()),
		).Exists(c.Context(), tx)
		if err != nil {
			return err
		}
		if !found {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("VIN signature author %s does not match any known aftermarket device.", recAddr))
		}
	}

	if req.Signature != "" && !userDevice.VinConfirmed { // if the user_device already exists and vin is confirmed, skip b/c likely somebody re-pairing the vehicle to different connection
		// check for existing vehicle from different user with same vin
		existing, err := models.UserDevices(
			models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(req.VIN)),
			models.UserDeviceWhere.VinConfirmed.EQ(true),
			models.UserDeviceWhere.TokenID.IsNotNull(),                                   // must be minted
			models.UserDeviceWhere.OwnerAddress.NEQ(null.BytesFrom(userEthAddr.Bytes())), // ok if owned by same user
		).Exists(c.Context(), tx)
		if err != nil {
			return err
		}
		if udc.Settings.IsProduction() && existing {
			logger.Warn().Msgf("VIN already in use by another vehicle: %s", req.VIN)
			return fiber.NewError(fiber.StatusConflict, "VIN already in use by another vehicle.")
		}
		userDevice.VinConfirmed = true
	}

	userDevice.VinIdentifier = null.StringFrom(req.VIN)
	if len(req.CountryCode) == 3 {
		// validate country_code
		if constants.FindCountry(req.CountryCode) == nil {
			return fiber.NewError(fiber.StatusBadRequest, "CountryCode does not exist: "+req.CountryCode)
		}
		userDevice.CountryCode = null.StringFrom(req.CountryCode)
	}
	if req.CANProtocol != "" {
		var udMD = &services.UserDeviceMetadata{}
		errMd := userDevice.Metadata.Unmarshal(udMD)
		if errMd != nil {
			udc.log.Err(errMd).Msgf("failed to unmarshal ud metadata. %s", string(userDevice.Metadata.JSON))
		} else {
			udMD.CANProtocol = &req.CANProtocol
			errMd = userDevice.Metadata.Marshal(udMD)
			if errMd != nil {
				udc.log.Err(errMd).Msgf("failed to marshal ud metadata. %+v", udMD)
			}
		}
	}

	if _, err := userDevice.Update(c.Context(), tx, boil.Infer()); err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if userDevice.CountryCode.Valid {
		if err := udc.updatePowerTrain(c.Context(), userDevice); err != nil {
			logger.Err(err).Msg("Failed to update powertrain type.")
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UnlockDoors godoc
// @Summary     Unlock the device's doors
// @Description Unlock the device's doors.
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       tokenID  path string true "Token ID"
// @Router      /vehicle/{tokenID}/commands/doors/unlock [post]
func (nc *NFTController) UnlockDoors(c *fiber.Ctx) error {
	return nc.handleEnqueueCommand(c, constants.DoorsUnlock)
}

// LockDoors godoc
// @Summary     Lock the device's doors
// @Description Lock the device's doors.
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       tokenID  path string true "Token ID"
// @Router      /vehicle/{tokenID}/commands/doors/lock [post]
func (nc *NFTController) LockDoors(c *fiber.Ctx) error {
	return nc.handleEnqueueCommand(c, constants.DoorsLock)
}

// StopCharging godoc
// @Summary     Stop the vehicle charging.
// @Description Stop the vehicle charging.
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       tokenID  path string true "Token ID"
// @Router      /vehicle/{tokenID}/commands/charge/stop [post]
func (nc *NFTController) ChargeStop(c *fiber.Ctx) error {
	return nc.handleEnqueueCommand(c, constants.ChargeStop)
}

// StartCharging godoc
// @Summary     Start the vehicle charging.
// @Description Start the vehicle charging.
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       tokenID  path string true "Token ID"
// @Router      /vehicle/{tokenID}/commands/charge/start [post]
func (nc *NFTController) ChargeStart(c *fiber.Ctx) error {
	return nc.handleEnqueueCommand(c, constants.ChargeStart)
}

// OpenTrunk godoc
// @Summary     Open the device's rear trunk
// @Description Open the device's front trunk. Currently, this only works for Teslas connected through Tesla.
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       tokenID  path string true "Token ID"
// @Router      /vehicle/{tokenID}/commands/trunk/open [post]
func (nc *NFTController) OpenTrunk(c *fiber.Ctx) error {
	return nc.handleEnqueueCommand(c, constants.TrunkOpen)
}

// OpenFrunk godoc
// @Summary     Open the device's front trunk
// @Description Open the device's front trunk. Currently, this only works for Teslas connected through Tesla.
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       tokenID  path string true "Token ID"
// @Router      /vehicle/{tokenID}/commands/frunk/open [post]
func (nc *NFTController) OpenFrunk(c *fiber.Ctx) error {
	return nc.handleEnqueueCommand(c, constants.FrunkOpen)
}

// handleEnqueueCommand enqueues the command specified by commandPath with the
// appropriate task service.
//
// Grabs token ID and privileges from Ctx.
func (nc *NFTController) handleEnqueueCommand(c *fiber.Ctx, commandPath string) error {
	tokenIDRaw := c.Params("tokenID")

	logger := nc.log.With().
		Str("feature", "commands").
		Str("tokenID", tokenIDRaw).
		Str("commandPath", commandPath).
		Logger()

	logger.Info().Msg("Received command request.")

	tokenID, ok := new(decimal.Big).SetString(tokenIDRaw)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tokenID))
	}

	// Checking both that the nft exists and is linked to a device.
	nft, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(types.NewNullDecimal(tokenID)),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "Vehicle NFT not found.")
		}
		logger.Err(err).Msg("Failed to search for device.")
		return opaqueInternalError
	}

	apInt, err := nc.integSvc.GetAutoPiIntegration(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Couldn't reach definitions server.")
	}

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(nft.ID),
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.NEQ(apInt.Id),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No command-capable integrations found for this vehicle.")
		}
		logger.Err(err).Msg("Failed to search for device integration record.")
		return opaqueInternalError
	}

	md := new(services.UserDeviceAPIIntegrationsMetadata)
	if err := udai.Metadata.Unmarshal(md); err != nil {
		logger.Err(err).Msg("Couldn't parse metadata JSON.")
		return opaqueInternalError
	}

	// TODO(elffjs): This map is ugly. Surely we interface our way out of this?
	commandMap := map[string]map[string]func(udai *models.UserDeviceAPIIntegration) (string, error){
		constants.TeslaVendor: {
			"doors/unlock": nc.teslaTaskService.UnlockDoors,
			"doors/lock":   nc.teslaTaskService.LockDoors,
			"trunk/open":   nc.teslaTaskService.OpenTrunk,
			"frunk/open":   nc.teslaTaskService.OpenFrunk,
			"charge/start": nc.teslaTaskService.ChargeStart,
			"charge/stop":  nc.teslaTaskService.ChargeStop,
		},
	}

	integration, err := nc.deviceDefSvc.GetIntegrationByID(c.Context(), udai.IntegrationID)
	if err != nil {
		return grpcfiber.GrpcErrorToFiber(err, "deviceDefSvc error getting integration id: "+udai.IntegrationID)
	}

	vendorCommandMap, ok := commandMap[integration.Vendor]
	if !ok {
		return fiber.NewError(fiber.StatusConflict, "Integration is not capable of this command.")
	}

	if (md.Commands == nil || !slices.Contains(md.Commands.Enabled, commandPath)) &&
		(integration.Vendor != constants.TeslaVendor || !slices.Contains([]string{"charge/start", "charge/stop"}, commandPath)) { // Ugly hack for Tesla charge being tacked on for a pilot.

		return fiber.NewError(fiber.StatusBadRequest, "Integration is not capable of this command.")
	}

	commandFunc, ok := vendorCommandMap[commandPath]
	if !ok {
		// Should never get here.
		logger.Error().Msg("Command was enabled for this device, but there is no function to execute it.")
		return fiber.NewError(fiber.StatusConflict, "Integration is not capable of this command.")
	}

	// we need to call tesla-oracle grpc endpoint to check if we should drop the command based on the subscription status : pending, active, inactive
	resp, err := nc.oracleClient.GetSyntheticDevicesByVIN(c.Context(), &pb_oracle.GetSyntheticDevicesByVINRequest{Vin: nft.VinIdentifier.String})
	if err != nil {
		logger.Err(err).Msg("Failed to reach tesla-oracle service.")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to reach tesla-oracle service.")
	}

	if len(resp.SyntheticDevices) == 0 {
		logger.Warn().Msgf("No synthetic devices found for VIN %s, dropping command request.", nft.VinIdentifier.String)
		return fiber.NewError(fiber.StatusNotFound, "No synthetic devices found for this VIN.")
	}

	sd := resp.SyntheticDevices[0] // is it ok to take first?
	// Check if the subscription status is inactive, and drop the command request if so.
	if sd.SubscriptionStatus == "inactive" {
		logger.Info().Str("subscriptionStatus", sd.SubscriptionStatus).Msgf("Dropping command request for vehicle %s due to subscription status.", nft.VinIdentifier.String)
		return fiber.NewError(fiber.StatusForbidden, "Vehicle subscription is not active.")
	}
	subTaskID, err := commandFunc(udai)
	if err != nil {
		logger.Err(err).Msg("Failed to start command task.")
		return opaqueInternalError
	}

	comRow := &models.DeviceCommandRequest{
		ID:            subTaskID,
		UserDeviceID:  nft.ID,
		IntegrationID: udai.IntegrationID,
		Command:       commandPath,
		Status:        models.DeviceCommandRequestStatusPending,
	}

	if err := comRow.Insert(c.Context(), nc.DBS().Writer, boil.Infer()); err != nil {
		logger.Err(err).Msg("Couldn't insert device command request record.")
		return opaqueInternalError
	}

	logger.Info().Msg("Successfully enqueued command.")

	return c.JSON(CommandResponse{RequestID: subTaskID})
}

// BurnRequest contains the user's signature for the burn request.
type BurnRequest struct {
	// Signature is the hex encoding of the EIP-712 signature result.
	Signature string `json:"signature" validate:"required"`
}
