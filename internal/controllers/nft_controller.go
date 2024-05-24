package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/services/ipfs"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/internal/utils"
	"github.com/DIMO-Network/shared"
	"github.com/segmentio/ksuid"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/go-mnemonic"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
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
	deviceDataSvc    services.DeviceDataService
	ipfsSvc          *ipfs.IPFS
}

// NewNFTController constructor
func NewNFTController(settings *config.Settings, dbs func() *db.ReaderWriter, logger *zerolog.Logger, s3 *s3.Client,
	deviceDefSvc services.DeviceDefinitionService,
	smartcarTaskSvc services.SmartcarTaskService,
	teslaTaskService services.TeslaTaskService,
	integSvc services.DeviceDefinitionIntegrationService,
	deviceDataSvc services.DeviceDataService,
	ipfsSvc *ipfs.IPFS,
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
		deviceDataSvc:    deviceDataSvc,
		ipfsSvc:          ipfsSvc,
	}
}

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

	ud, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(tid),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Indexers start looking immediately.
			helpers.SkipErrorLog(c)
			return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
		}
		nc.log.Err(err).Msg("Database error retrieving NFT metadata.")
		return opaqueInternalError
	}

	deviceDefinitionID := ud.DeviceDefinitionID
	def, err := nc.deviceDefSvc.GetDeviceDefinitionByID(c.Context(), deviceDefinitionID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "failed to get device definition")
	}

	description := fmt.Sprintf("%s %s %d", def.Make.Name, def.Type.Model, def.Type.Year)

	var name string
	if ud.Name.Valid {
		name = ud.Name.String
	} else {
		name = description
	}

	imageURI := fmt.Sprintf("%s/v1/vehicle/%s/image", nc.Settings.DeploymentBaseURL, ti)
	if !nc.Settings.IsProduction() && ud.IpfsImageCid.Valid {
		imageURI = ipfs.URL(ud.IpfsImageCid.String)
	}

	return c.JSON(NFTMetadataResp{
		Name:        name,
		Description: description + ", a DIMO vehicle.",
		Image:       imageURI,
		Attributes: []NFTAttribute{
			{TraitType: "Make", Value: def.Make.Name},
			{TraitType: "Model", Value: def.Type.Model},
			{TraitType: "Year", Value: strconv.Itoa(int(def.Type.Year))},
		},
	})
}

// GetIntegrationNFTMetadata godoc
// @Description gets an integration using its tokenID
// @Tags        integrations
// @Produce     json
// @Success     200 {array} controllers.NFTMetadataResp
// @Router      /integration/{tokenID} [get]
func (nc *NFTController) GetIntegrationNFTMetadata(c *fiber.Ctx) error {
	tokenID := c.Params("tokenID")

	uTokenID, err := strconv.ParseUint(tokenID, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid tokenID provided")
	}

	integration, err := nc.deviceDefSvc.GetIntegrationByTokenID(c.Context(), uTokenID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "failed to get integration")
	}

	return c.JSON(NFTMetadataResp{
		Name:        integration.Vendor,
		Description: fmt.Sprintf("%s, a DIMO integration", integration.Vendor),
		Attributes:  []NFTAttribute{},
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

	// todo: NFT not found errors here were getting hit a lot in prod - should we have a prometheus metric or we don't care?
	nft, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(tid),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.SkipErrorLog(c)
			return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
		}
		nc.log.Err(err).Msg("Database error retrieving NFT metadata.")
		return opaqueInternalError
	}

	if nc.Settings.IsProduction() {
		if !nft.MintRequestID.Valid {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("No image available for vehicle %d.", ti))
		}

		imageName := nft.MintRequestID.String
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
					// todo: this error was getting hit a lot in production
					helpers.SkipErrorLog(c)
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

	if !nft.IpfsImageCid.Valid {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("nft image not set for vehicle %d.", ti))
	}

	imgB, err := nc.ipfsSvc.FetchImage(c.Context(), nft.IpfsImageCid.String)
	if err != nil {
		return fiber.NewError(fiber.StatusFailedDependency, "failed to fetch image from IPFS")
	}

	c.Set("Content-Type", "image/png")
	return c.Send(imgB)

}

// GetAftermarketDeviceNFTMetadataByAddress godoc
// @Description Retrieves NFT metadata for a given aftermarket device, using the device's
// @Description Ethereum address.
// @Tags        nfts
// @Param       address path string true "Ethereum address for the device."
// @Produce     json
// @Success     200 {object} controllers.NFTMetadataResp
// @Failure     404
// @Router      /aftermarket/device/by-address/{address} [get]
func (nc *NFTController) GetAftermarketDeviceNFTMetadataByAddress(c *fiber.Ctx) error {
	maybeAddr := c.Params("address")

	if !common.IsHexAddress(maybeAddr) {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse device address.")
	}

	addr := common.HexToAddress(maybeAddr)

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.EthereumAddress.EQ(addr.Bytes()),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that address.")
		}
		return err
	}

	var name string
	if three, err := mnemonic.EntropyToMnemonicThreeWords(ad.EthereumAddress); err == nil {
		name = strings.Join(three, " ")
	}

	return c.JSON(NFTMetadataResp{
		Name:        name,
		Description: name + ", a DIMO hardware device.",
		Image:       fmt.Sprintf("%s/v1/aftermarket/device/%s/image", nc.Settings.DeploymentBaseURL, ad.TokenID),
		Attributes: []NFTAttribute{
			{TraitType: "Ethereum Address", Value: common.BytesToAddress(ad.EthereumAddress).String()},
			{TraitType: "Serial Number", Value: ad.Serial},
		},
	})
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

	unit, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.TokenID.EQ(utils.BigToDecimal(tid)),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id.")
		}
		return err
	}
	var name string
	if three, err := mnemonic.EntropyToMnemonicThreeWords(unit.EthereumAddress); err == nil {
		name = strings.Join(three, " ")
	}

	return c.JSON(NFTMetadataResp{
		Name:        name,
		Description: name + ", a DIMO hardware device.",
		Image:       fmt.Sprintf("%s/v1/aftermarket/device/%s/image", nc.Settings.DeploymentBaseURL, tid),
		Attributes: []NFTAttribute{
			{TraitType: "Ethereum Address", Value: common.BytesToAddress(unit.EthereumAddress).String()},
			{TraitType: "Serial Number", Value: unit.Serial},
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

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.TokenID.EQ(utils.BigToDecimal(ti)),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "No device with id.")
		}
		return err
	}

	dm, err := nc.deviceDefSvc.GetMakeByTokenID(c.Context(), ad.DeviceManufacturerTokenID.Int(nil))
	if err != nil {
		return err
	}

	var key string

	switch dm.Name {
	case constants.AutoPiVendor:
		key = nc.Settings.AutoPiNFTImage
	case "Hashdog":
		key = nc.Settings.MacaronNFTImage
	default:
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("No NFT image for manufacturer %s.", dm.Name))
	}

	s3o, err := nc.s3.GetObject(c.Context(), &s3.GetObjectInput{
		Bucket: aws.String(nc.Settings.NFTS3Bucket),
		Key:    aws.String(key),
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
		return shared.GrpcErrorToFiber(err, "deviceDefSvc error getting integration id: "+udai.IntegrationID)
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

// GetVinCredential godoc
// @Description Returns the vin credential for the vehicle with a given token id.
// @Tags        permission
// @Param       tokenId path int true "token id"
// @Produce     json
// @Success     200 {object} map[string]any
// @Failure     404
// @Router      /vehicle/{tokenId}/vin-credential [get]
func (nc *NFTController) GetVinCredential(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	ud, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(tid),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
		}
		nc.log.Err(err).Msg("Database error retrieving NFT metadata.")
		return opaqueInternalError
	}

	if ud.R.Claim == nil {
		return fiber.NewError(fiber.StatusNotFound, "Credential associated with NFT not found.")
	}

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return c.Send(ud.R.Claim.Credential.JSON)
}

// GetBurnDevice godoc
// @Description Returns the data the user must sign in order to burn the device.
// @Param       tokenID path int true "token id"
// @Success     200          {object} apitypes.TypedData
// @Security    BearerAuth
// @Router     /user/vehicle/{tokenID}/commands/burn [get]
func (udc *UserDevicesController) GetBurnDevice(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("failed to parse token id %q", tis))
	}
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

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

	tx, err := udc.DBS().Reader.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(tid),
		qm.Load(models.UserDeviceRels.BurnRequest),
		qm.Load(models.UserDeviceRels.VehicleTokenAftermarketDevice),
		qm.Load(models.UserDeviceRels.VehicleTokenSyntheticDevice),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No vehicle NFT with that token id.")
		}
		return err
	}

	bvs, _, err := udc.checkDeviceBurn(c.Context(), userDevice)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(client.GetPayload(&bvs))
}

// PostBurnDevice godoc
// @Description Sends a burn device request to the blockchain
// @Param       tokenID path int true "token id"
// @Param       burnRequest  body controllers.BurnRequest true "Signature"
// @Success     200
// @Security    BearerAuth
// @Router      /user/vehicle/{tokenID}/commands/burn [post]
func (udc *UserDevicesController) PostBurnDevice(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("failed to parse token id %q", tis))
	}
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

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

	tx, err := udc.DBS().Reader.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(tid),
		qm.Load(models.UserDeviceRels.BurnRequest),
		qm.Load(models.UserDeviceRels.VehicleTokenAftermarketDevice),
		qm.Load(models.UserDeviceRels.VehicleTokenSyntheticDevice),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No vehicle NFT with that token id.")
		}
		return err
	}

	bvs, user, err := udc.checkDeviceBurn(c.Context(), userDevice)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var br BurnRequest
	if err := c.BodyParser(&br); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	udc.log.Info().
		Interface("httpRequestBody", br).
		Interface("client", client).
		Interface("burnVehicleSign", bvs).
		Interface("typedData", client.GetPayload(&bvs)).
		Msg("Got request.")

	hash, err := client.Hash(&bvs)
	if err != nil {
		return fmt.Errorf("could not hash bvs: %w", err)
	}

	sigBytes := common.FromHex(br.Signature)
	recAddr, err := helpers.Ecrecover(hash, sigBytes)
	if err != nil {
		return fmt.Errorf("could not recover signature: %w", err)
	}

	realAddr := common.HexToAddress(*user.EthereumAddress)
	if recAddr != realAddr {
		return fiber.NewError(fiber.StatusBadRequest, "signature incorrect")
	}

	requestID := ksuid.New().String()

	mtr := models.MetaTransactionRequest{
		ID:     requestID,
		Status: models.MetaTransactionRequestStatusUnsubmitted,
	}

	if err := mtr.Insert(c.Context(), tx, boil.Infer()); err != nil {
		return fmt.Errorf("failed to insert metatransaction request: %w", err)
	}

	userDevice.BurnRequestID = null.StringFrom(requestID)
	if _, err := userDevice.Update(c.Context(), tx, boil.Infer()); err != nil {
		return fmt.Errorf("failed to update vehicle nft: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	udc.log.Info().Msgf("submitted metatransaction request %s", requestID)
	return client.BurnVehicleSign(requestID, bvs.TokenID, sigBytes)
}

func (udc *UserDevicesController) checkDeviceBurn(ctx context.Context, userDevice *models.UserDevice) (registry.BurnVehicleSign, *pb.User, error) {
	var bvs registry.BurnVehicleSign

	if userDevice.R.BurnRequest != nil && userDevice.R.BurnRequest.Status != models.MetaTransactionRequestStatusFailed {
		return bvs, nil, errors.New("burning already in progress")
	}

	if userDevice.R.VehicleTokenAftermarketDevice != nil || userDevice.R.VehicleTokenSyntheticDevice != nil {
		return bvs, nil, errors.New("vehicle must be unpaired to burn")
	}

	user, err := udc.usersClient.GetUser(ctx, &pb.GetUserRequest{Id: userDevice.UserID})
	if err != nil {
		return bvs, nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	if user.EthereumAddress == nil {
		return bvs, nil, errors.New("user does not have an Ethereum address on file")
	}

	return registry.BurnVehicleSign{
		TokenID: userDevice.TokenID.Int(nil),
	}, user, nil
}

// BurnRequest contains the user's signature for the burn request.
type BurnRequest struct {
	// Signature is the hex encoding of the EIP-712 signature result.
	Signature string `json:"signature" validate:"required"`
}
