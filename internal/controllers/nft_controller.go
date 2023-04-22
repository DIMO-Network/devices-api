package controllers

import (
	"database/sql"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/go-mnemonic"
	"github.com/DIMO-Network/shared/db"
	pr "github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"golang.org/x/exp/slices"
)

type NFTController struct {
	Settings         *config.Settings
	DBS              func() *db.ReaderWriter
	s3               *s3.Client
	log              *zerolog.Logger
	deviceDefSvc     services.DeviceDefinitionService
	integSvc         services.DeviceDefinitionIntegrationService
	smartcarTaskSvc  services.SmartcarTaskService
	teslaTaskService services.TeslaTaskService
}

// NewNFTController constructor
func NewNFTController(settings *config.Settings, dbs func() *db.ReaderWriter, logger *zerolog.Logger, s3 *s3.Client,
	deviceDefSvc services.DeviceDefinitionService,
	smartcarTaskSvc services.SmartcarTaskService,
	teslaTaskService services.TeslaTaskService,
	integSvc services.DeviceDefinitionIntegrationService,
) NFTController {
	return NFTController{
		Settings:         settings,
		DBS:              dbs,
		log:              logger,
		s3:               s3,
		deviceDefSvc:     deviceDefSvc,
		smartcarTaskSvc:  smartcarTaskSvc,
		teslaTaskService: teslaTaskService,
		integSvc:         integSvc,
	}
}

const (
	NonLocationData int64 = 1
	Commands        int64 = 2
	CurrentLocation int64 = 3
	AllTimeLocation int64 = 4
)

// GetNFTMetadata godoc
// @Description retrieves NFT metadata for a given tokenID
// @Tags        nfts
// @Param       tokenId path int true "token id"
// @Produce     json
// @Success     200 {object} controllers.NFTMetadataResp
// @Failure     404
// @Router      /vehicle/{tokenId} [get]
func (nc *NFTController) GetNFTMetadata(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}

	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	var maybeName null.String
	var deviceDefinitionID string

	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(tid),
		qm.Load(models.VehicleNFTRels.UserDevice),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
		}
		nc.log.Err(err).Msg("Database error retrieving NFT metadata.")
		return opaqueInternalError
	}

	if nft.R.UserDevice == nil {
		return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
	}

	maybeName = nft.R.UserDevice.Name
	deviceDefinitionID = nft.R.UserDevice.DeviceDefinitionID

	def, err := nc.deviceDefSvc.GetDeviceDefinitionByID(c.Context(), deviceDefinitionID)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get device definition")
	}

	description := fmt.Sprintf("%s %s %d", def.Make.Name, def.Type.Model, def.Type.Year)

	var name string
	if maybeName.Valid {
		name = maybeName.String
	} else {
		name = description
	}

	return c.JSON(NFTMetadataResp{
		Name:        name,
		Description: description,
		Image:       fmt.Sprintf("%s/v1/vehicle/%s/image", nc.Settings.DeploymentBaseURL, ti),
		Attributes: []NFTAttribute{
			{TraitType: "Make", Value: def.Make.Name},
			{TraitType: "Model", Value: def.Type.Model},
			{TraitType: "Year", Value: strconv.Itoa(int(def.Type.Year))},
			{TraitType: "Creation Date", Value: strconv.FormatInt(nft.R.UserDevice.CreatedAt.Unix(), 10)},
		},
	})
}

type NFTMetadataResp struct {
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Image       string         `json:"image,omitempty"`
	Attributes  []NFTAttribute `json:"attributes"`
}

type NFTAttribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

// GetNFTImage godoc
// @Description Returns the image for the given vehicle NFT.
// @Tags        nfts
// @Param       tokenId     path  int  true  "token id"
// @Param       transparent query bool false "whether to remove the image background"
// @Produce     png
// @Success     200
// @Failure     404
// @Router      /vehicle/{tokenId}/image [get]
func (nc *NFTController) GetNFTImage(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}

	var transparent bool
	if c.Query("transparent") == "true" {
		transparent = true
	}

	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	var imageName string

	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(tid),
		qm.Load(models.VehicleNFTRels.UserDevice),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
		}
		nc.log.Err(err).Msg("Database error retrieving NFT metadata.")
		return opaqueInternalError
	}

	if nft.R.UserDevice == nil {
		return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
	}

	imageName = nft.MintRequestID
	suffix := ".png"

	if transparent {
		suffix = "_transparent.png"
	}

	s3o, err := nc.s3.GetObject(c.Context(), &s3.GetObjectInput{
		Bucket: aws.String(nc.Settings.NFTS3Bucket),
		Key:    aws.String(imageName + suffix),
	})
	if err != nil {
		if transparent {
			var nsk *s3types.NoSuchKey
			if errors.As(err, &nsk) {
				return fiber.NewError(fiber.StatusNotFound, "Transparent version not set.")
			}
		}
		nc.log.Err(err).Msg("Failure communicating with S3.")
		return opaqueInternalError
	}
	defer s3o.Body.Close()

	b, err := io.ReadAll(s3o.Body)
	if err != nil {
		return err
	}

	c.Set("Content-Type", "image/png")
	return c.Send(b)
}

// GetAftermarketDeviceNFTMetadata godoc
// @Description Retrieves NFT metadata for a given aftermarket device.
// @Tags        nfts
// @Param       tokenId path int true "token id"
// @Produce     json
// @Success     200 {object} controllers.NFTMetadataResp
// @Failure     404
// @Router      /aftermarket/device/{tokenId} [get]
func (nc *NFTController) GetAftermarketDeviceNFTMetadata(c *fiber.Ctx) error {
	tidStr := c.Params("tokenID")

	tid, ok := new(big.Int).SetString(tidStr, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse token id.")
	}

	unit, err := models.AutopiUnits(
		models.AutopiUnitWhere.TokenID.EQ(types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tid, 0))),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id.")
		}
		return err
	}
	var name string
	if three, err := mnemonic.EntropyToMnemonicThreeWords(unit.EthereumAddress.Bytes); err == nil {
		name = strings.Join(three, " ")
	}

	return c.JSON(NFTMetadataResp{
		Name:        name,
		Description: name + ", a hardware device",
		Image:       fmt.Sprintf("%s/v1/aftermarket/device/%s/image", nc.Settings.DeploymentBaseURL, tid),
		Attributes: []NFTAttribute{
			{TraitType: "Ethereum Address", Value: common.BytesToAddress(unit.EthereumAddress.Bytes).String()},
			{TraitType: "Serial Number", Value: unit.AutopiUnitID},
		},
	})
}

// GetAftermarketDeviceNFTImage godoc
// @Description Returns the image for the given aftermarket device NFT.
// @Tags        nfts
// @Param       tokenId     path  int  true  "token id"
// @Produce     png
// @Success     200
// @Failure     404
// @Router      /aftermarket/device/{tokenId}/image [get]
func (nc *NFTController) GetAftermarketDeviceNFTImage(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}

	exists, err := models.AutopiUnits(
		models.AutopiUnitWhere.TokenID.EQ(types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))),
	).Exists(c.Context(), nc.DBS().Reader)
	if err != nil {
		return err
	}

	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "No device with id.")
	}

	s3o, err := nc.s3.GetObject(c.Context(), &s3.GetObjectInput{
		Bucket: aws.String(nc.Settings.NFTS3Bucket),
		Key:    aws.String(nc.Settings.AutoPiNFTImage),
	})
	if err != nil {
		nc.log.Err(err).Msg("Failure communicating with S3.")
		return opaqueInternalError
	}
	defer s3o.Body.Close()

	b, err := io.ReadAll(s3o.Body)
	if err != nil {
		return err
	}

	c.Set("Content-Type", "image/png")
	return c.Send(b)
}

// GetManufacturerNFTMetadata godoc
// @Description Retrieves NFT metadata for a given manufacturer.
// @Tags        nfts
// @Param       tokenId path int true "token id"
// @Produce     json
// @Success     200 {object} controllers.NFTMetadataResp
// @Failure     404
// @Router      /manufacturer/{tokenId} [get]
func (nc *NFTController) GetManufacturerNFTMetadata(c *fiber.Ctx) error {
	tidStr := c.Params("tokenID")

	tid, ok := new(big.Int).SetString(tidStr, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse token id.")
	}

	dm, err := nc.deviceDefSvc.GetMakeByTokenID(c.Context(), tid)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "Couldn't retrieve manufacturer")
	}

	return c.JSON(NFTMetadataResp{
		Name:       dm.Name,
		Attributes: []NFTAttribute{},
	})
}

// GetVehicleStatus godoc
// @Description Returns the latest status update for the vehicle with a given token id.
// @Tags        permission
// @Param       tokenId path int true "token id"
// @Produce     json
// @Success     200 {object} controllers.DeviceSnapshot
// @Failure     404
// @Router      /vehicle/{tokenId}/status [get]
func (nc *NFTController) GetVehicleStatus(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	claims := c.Locals("tokenClaims").(pr.CustomClaims)

	privileges := claims.PrivilegeIDs

	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}

	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))
	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(tid),
		qm.Load(models.VehicleNFTRels.UserDevice),
	).One(c.Context(), nc.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
		}
		nc.log.Err(err).Msg("Database error retrieving NFT metadata.")
		return opaqueInternalError
	}

	if nft.R.UserDevice == nil {
		return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
	}

	deviceData, err := models.UserDeviceData(models.UserDeviceDatumWhere.UserDeviceID.EQ(nft.R.UserDevice.ID),
		qm.OrderBy("updated_at asc")).All(c.Context(), nc.DBS().Reader)
	if errors.Is(err, sql.ErrNoRows) || len(deviceData) == 0 || !deviceData[0].Data.Valid {
		return fiber.NewError(fiber.StatusNotFound, "no status updates yet")
	}
	if err != nil {
		return err
	}

	ds := PrepareDeviceStatusInformation(deviceData, privileges)

	return c.JSON(ds)
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
	return nc.handleEnqueueCommand(c, "doors/unlock")
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
	return nc.handleEnqueueCommand(c, "doors/lock")
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
	return nc.handleEnqueueCommand(c, "trunk/open")
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
	return nc.handleEnqueueCommand(c, "frunk/open")
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
	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(types.NewNullDecimal(tokenID)),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "Vehicle NFT not found.")
		}
		logger.Err(err).Msg("Failed to search for device.")
		return opaqueInternalError
	}

	if !nft.UserDeviceID.Valid {
		return fiber.NewError(fiber.StatusConflict, "NFT not attached to a user device.")
	}

	apInt, err := nc.integSvc.GetAutoPiIntegration(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Couldn't reach definitions server.")
	}

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(nft.UserDeviceID.String),
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
		constants.SmartCarVendor: {
			"doors/unlock": nc.smartcarTaskSvc.UnlockDoors,
			"doors/lock":   nc.smartcarTaskSvc.LockDoors,
		},
		constants.TeslaVendor: {
			"doors/unlock": nc.teslaTaskService.UnlockDoors,
			"doors/lock":   nc.teslaTaskService.LockDoors,
			"trunk/open":   nc.teslaTaskService.OpenTrunk,
			"frunk/open":   nc.teslaTaskService.OpenFrunk,
		},
	}

	integration, err := nc.deviceDefSvc.GetIntegrationByID(c.Context(), udai.IntegrationID)
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
		UserDeviceID:  nft.UserDeviceID.String,
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
