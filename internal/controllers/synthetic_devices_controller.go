package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"strconv"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type SyntheticDevicesController struct {
	Settings       *config.Settings
	DBS            func() *db.ReaderWriter
	log            *zerolog.Logger
	deviceDefSvc   services.DeviceDefinitionService
	usersClient    pb.UserServiceClient
	walletSvc      services.SyntheticWalletInstanceService
	registryClient registry.Client
}

type MintSyntheticDeviceRequest struct {
	Signature string `json:"signature"`
}

type SyntheticDeviceSequence struct {
	NextVal int `boil:"nextval"`
}

func NewSyntheticDevicesController(
	settings *config.Settings,
	dbs func() *db.ReaderWriter,
	logger *zerolog.Logger,
	deviceDefSvc services.DeviceDefinitionService,
	usersClient pb.UserServiceClient,
	walletSvc services.SyntheticWalletInstanceService,
	registryClient registry.Client,
) SyntheticDevicesController {
	return SyntheticDevicesController{
		Settings:       settings,
		DBS:            dbs,
		log:            logger,
		usersClient:    usersClient,
		deviceDefSvc:   deviceDefSvc,
		walletSvc:      walletSvc,
		registryClient: registryClient,
	}
}

func (sdc *SyntheticDevicesController) getEIP712Mint(integrationID, vehicleNode int64) *signer.TypedData {
	return &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			// Need to keep this name until the contract changes.
			"MintSyntheticDeviceSign": []signer.Type{
				{Name: "integrationNode", Type: "uint256"},
				{Name: "vehicleNode", Type: "uint256"},
			},
		},
		PrimaryType: "MintSyntheticDeviceSign",
		Domain: signer.TypedDataDomain{
			Name:              "DIMO",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(sdc.Settings.DIMORegistryChainID),
			VerifyingContract: sdc.Settings.DIMORegistryAddr,
		},
		Message: signer.TypedDataMessage{
			"integrationNode": math.NewHexOrDecimal256(integrationID),
			"vehicleNode":     math.NewHexOrDecimal256(vehicleNode),
		},
	}
}

// GetSyntheticDeviceMintingPayload godoc
// @Description Produces the payload that the user signs and submits to mint a synthetic device for
// @Description the given vehicle and integration.
// @Tags        integrations
// @Produce     json
// @Param       userDeviceID path int true "user device KSUID"
// @Param       integrationID path int true "integration KSUD, must be software-based"
// @Success     200 {array} signer.TypedData
// @Router 	    /user/devices/{userDeviceID}/integrations/{integrationID}/commands/mint [get]
func (sdc *SyntheticDevicesController) GetSyntheticDeviceMintingPayload(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	user, err := sdc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "error occurred when fetching user")
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "User does not have an Ethereum address.")
	}

	userAddr := common.HexToAddress(*user.EthereumAddress)

	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")
	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenSyntheticDevice)),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID)),
	).One(c.Context(), sdc.DBS().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "No vehicle with that id found.")
		}
		return err
	}

	if ud.R.VehicleNFT == nil || ud.R.VehicleNFT.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not minted.")
	}

	if userAddr != common.BytesToAddress(ud.R.VehicleNFT.OwnerAddress.Bytes) {
		return fiber.NewError(fiber.StatusUnauthorized, "User's address does not control this device.")
	}

	if ud.R.VehicleNFT.R.VehicleTokenSyntheticDevice != nil {
		return fiber.NewError(fiber.StatusConflict, "Vehicle already paired with a synthetic device.")
	}

	if len(ud.R.UserDeviceAPIIntegrations) == 0 {
		return fiber.NewError(fiber.StatusConflict, "Vehicle does not have this kind of connection.")
	}

	in, err := sdc.deviceDefSvc.GetIntegrationByID(c.Context(), integrationID)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get integration")
	}

	if in.Vendor != constants.SmartCarVendor && in.Vendor != constants.TeslaVendor {
		return fiber.NewError(fiber.StatusConflict, "This is not a software connection.")
	}

	if in.TokenId == 0 {
		return fiber.NewError(fiber.StatusConflict, "Connection type not yet minted.")
	}

	vid, ok := ud.R.VehicleNFT.TokenID.Int64()
	if !ok {
		return fmt.Errorf("vehicle token id invalid, this should never happen %d", ud.R.VehicleNFT.TokenID)
	}

	response := sdc.getEIP712Mint(int64(in.TokenId), vid)

	return c.JSON(response)
}

// MintSyntheticDevice godoc
// @Description Submit a metadata
// @Tags        integrations
// @Produce     json
// @Param       userDeviceID path int true "user device KSUID, must be minted"
// @Param       integrationID path int true "integration KSUD, must be software-based"
// @Success     204
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/mint [post]
func (sdc *SyntheticDevicesController) MintSyntheticDevice(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenSyntheticDevice)),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID)),
	).One(c.Context(), sdc.DBS().Reader)
	if err != nil {
		return err
	}

	if ud.R.VehicleNFT == nil || ud.R.VehicleNFT.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not minted.")
	}

	if ud.R.VehicleNFT.R.VehicleTokenSyntheticDevice != nil {
		return fiber.NewError(fiber.StatusConflict, "Vehicle already paired with a synthetic device.")
	}

	if len(ud.R.UserDeviceAPIIntegrations) == 0 {
		return fiber.NewError(fiber.StatusConflict, "Vehicle does not have this kind of connection.")
	}

	in, err := sdc.deviceDefSvc.GetIntegrationByID(c.Context(), integrationID)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get integration")
	}

	if in.Vendor != constants.SmartCarVendor && in.Vendor != constants.TeslaVendor {
		return fiber.NewError(fiber.StatusConflict, "This is not a software connection.")
	}

	if in.TokenId == 0 {
		return fiber.NewError(fiber.StatusConflict, "Connection type not yet minted.")
	}

	vid, ok := ud.R.VehicleNFT.TokenID.Int64()
	if !ok {
		return fmt.Errorf("vehicle token id invalid, this should never happen %d", ud.R.VehicleNFT.TokenID)
	}

	userID := helpers.GetUserID(c)
	var req MintSyntheticDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request.")
	}

	user, err := sdc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "error occurred when fetching user")
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusConflict, "User does not have an Ethereum address.")
	}

	userAddr := common.HexToAddress(*user.EthereumAddress)
	rawPayload := sdc.getEIP712Mint(int64(in.TokenId), vid)

	tdHash, _, err := signer.TypedDataAndHash(*rawPayload)
	if err != nil {
		sdc.log.Err(err).Msg("Error occurred creating hash of payload")
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't verify signature.")
	}

	ownerSignature := common.FromHex(req.Signature)
	recAddr, err := helpers.Ecrecover(tdHash, ownerSignature)
	if err != nil {
		sdc.log.Err(err).Msg("unable to validate signature")
		return err
	}

	if recAddr != userAddr {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signature.")
	}

	childKeyNumber, err := sdc.generateNextChildKeyNumber(c.Context())
	if err != nil {
		sdc.log.Err(err).Msg("failed to generate sequence from database")
		return fiber.NewError(fiber.StatusInternalServerError, "synthetic device minting request failed")
	}

	requestID := ksuid.New().String()

	syntheticDeviceAddr, err := sdc.sendSyntheticDeviceMintPayload(c.Context(), requestID, tdHash, int(vid), in.TokenId, ownerSignature, childKeyNumber)
	if err != nil {
		sdc.log.Err(err).Msg("synthetic device minting request failed")
		return fiber.NewError(fiber.StatusInternalServerError, "synthetic device minting request failed")
	}

	tx, err := sdc.DBS().Writer.DB.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}

	metaReq := &models.MetaTransactionRequest{
		ID:     requestID,
		Status: models.MetaTransactionRequestStatusUnsubmitted,
	}

	if err = metaReq.Insert(c.Context(), tx, boil.Infer()); err != nil {
		sdc.log.Err(err).Msg("error occurred creating meta transaction request")
		return fiber.NewError(fiber.StatusInternalServerError, "synthetic device minting request failed")
	}

	syntheticDevice := &models.SyntheticDevice{
		VehicleTokenID:     types.NewNullDecimal(decimal.New(vid, 0)),
		IntegrationTokenID: types.NewDecimal(decimal.New(int64(in.TokenId), 0)),
		WalletChildNumber:  childKeyNumber,
		WalletAddress:      syntheticDeviceAddr,
		MintRequestID:      requestID,
	}

	if err = syntheticDevice.Insert(c.Context(), tx, boil.Infer()); err != nil {
		sdc.log.Err(err).Msg("error occurred saving synthetic device")
		return fiber.NewError(fiber.StatusInternalServerError, "synthetic device minting request failed")
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Submitted synthetic device mint request."})
}

// GetSyntheticDeviceBurnPayload godoc
// @Description Produces the payload that the user signs and submits to burn a synthetic device.
// @Produce     json
// @Param       syntheticDeviceNode path int true "synthetic device token id"
// @Success     200 {array} signer.TypedData
// @Router      /synthetic/device/{syntheticDeviceNode}/burn [get]
func (sdc *SyntheticDevicesController) GetSyntheticDeviceBurnPayload(c *fiber.Ctx) error {
	syntheticDeviceNodeRaw := c.Params("syntheticDeviceNode")
	userID := helpers.GetUserID(c)

	syntheticDeviceNode, err := strconv.ParseInt(syntheticDeviceNodeRaw, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Node id %q must be an integer.", syntheticDeviceNodeRaw))
	}

	sd, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(syntheticDeviceNode, 0))),
		qm.Load(models.SyntheticDeviceRels.VehicleToken),
	).One(c.Context(), sdc.DBS().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("No device with id %d found.", syntheticDeviceNode))
		}
		return err
	}

	vOwn := common.BytesToAddress(sd.R.VehicleToken.OwnerAddress.Bytes)

	user, err := sdc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		sdc.log.Debug().Err(err).Msg("error occurred when fetching user")
		return helpers.GrpcErrorToFiber(err, "error occurred when fetching user")
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusForbidden, "No Ethereum address on file for user.")
	}

	addr := common.HexToAddress(*user.EthereumAddress)

	if vOwn != addr {
		return fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("Vehicle is owned by %s, your address is %s.", vOwn, addr))
	}

	vehicleNode, _ := sd.VehicleTokenID.Int64()

	return c.JSON(sdc.getEIP712Burn(vehicleNode, syntheticDeviceNode))
}

type BurnSyntheticDeviceRequest struct {
	OwnerSignature string `json:"ownerSignature"`
}

// BurnSyntheticDevice godoc
// @Description Submit the signature required for the synthetic device burning meta-transaction.
// @Produce     json
// @Param       syntheticDeviceNode path int true "synthetic device token id"
// @Success     200
// @Router      /synthetic/device/{syntheticDeviceNode}/burn [post]
func (sdc *SyntheticDevicesController) BurnSyntheticDevice(c *fiber.Ctx) error {
	syntheticDeviceNodeRaw := c.Params("syntheticDeviceNode")
	userID := helpers.GetUserID(c)

	syntheticDeviceNode, err := strconv.ParseInt(syntheticDeviceNodeRaw, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Node id %q must be an integer.", syntheticDeviceNodeRaw))
	}

	sd, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(syntheticDeviceNode, 0))),
		qm.Load(models.SyntheticDeviceRels.VehicleToken),
	).One(c.Context(), sdc.DBS().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("No device with id %d found.", syntheticDeviceNode))
		}
		return err
	}

	vOwn := common.BytesToAddress(sd.R.VehicleToken.OwnerAddress.Bytes)

	user, err := sdc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		sdc.log.Debug().Err(err).Msg("error occurred when fetching user")
		return helpers.GrpcErrorToFiber(err, "error occurred when fetching user")
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusForbidden, "No Ethereum address on file for user.")
	}

	addr := common.HexToAddress(*user.EthereumAddress)

	if vOwn != addr {
		return fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("Vehicle is owned by %s, your address is %s.", vOwn, addr))
	}

	vehicleNode, _ := sd.VehicleTokenID.Int64()

	td := sdc.getEIP712Burn(vehicleNode, syntheticDeviceNode)

	var req BurnSyntheticDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	ownerSignature := common.FromHex(req.OwnerSignature)
	if len(ownerSignature) != 65 {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Owner signature was length %d, not 65.", len(ownerSignature)))
	}

	hash, _, err := signer.TypedDataAndHash(*td)
	if err != nil {
		sdc.log.Err(err).Msg("Error occurred creating has of payload")
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't verify signature.")
	}

	if recAddr, err := helpers.Ecrecover(hash, ownerSignature); err != nil {
		return err
	} else if recAddr != addr {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signature.")
	}

	reqID := ksuid.New().String()

	if err := sdc.registryClient.BurnSyntheticDeviceSign(reqID, big.NewInt(vehicleNode), big.NewInt(syntheticDeviceNode), ownerSignature); err != nil {
		return err
	}

	mtr := models.MetaTransactionRequest{
		ID:     reqID,
		Status: models.MetaTransactionRequestStatusUnsubmitted,
	}

	if err := mtr.Insert(c.Context(), sdc.DBS().Writer, boil.Infer()); err != nil {
		return err
	}

	sd.BurnRequestID = null.StringFrom(reqID)
	_, err = sd.Update(c.Context(), sdc.DBS().Writer, boil.Infer())

	return err
}

func (sdc *SyntheticDevicesController) sendSyntheticDeviceMintPayload(ctx context.Context, requestID string, hash []byte, vehicleNode int, intTokenID uint64, ownerSignature []byte, childKeyNumber int) ([]byte, error) {
	syntheticDeviceAddr, err := sdc.walletSvc.GetAddress(ctx, uint32(childKeyNumber))
	if err != nil {
		sdc.log.Err(err).
			Str("function-name", "SyntheticWallet.GetAddress").
			Int("childKeyNumber", childKeyNumber).
			Msg("Error occurred getting synthetic wallet address")
		return nil, err
	}

	virtSig, err := sdc.walletSvc.SignHash(ctx, uint32(childKeyNumber), hash)
	if err != nil {
		sdc.log.Err(err).
			Str("function-name", "SyntheticWallet.SignHash").
			Bytes("Hash", hash).
			Int("childKeyNumber", childKeyNumber).
			Msg("Error occurred signing message hash")
		return nil, err
	}

	vNode := new(big.Int).SetInt64(int64(vehicleNode))
	mvt := contracts.MintSyntheticDeviceInput{
		IntegrationNode:     new(big.Int).SetUint64(intTokenID),
		VehicleNode:         vNode,
		VehicleOwnerSig:     ownerSignature,
		SyntheticDeviceAddr: common.BytesToAddress(syntheticDeviceAddr),
		SyntheticDeviceSig:  virtSig,
	}

	return syntheticDeviceAddr, sdc.registryClient.MintSyntheticDeviceSign(requestID, mvt)
}

func (sdc *SyntheticDevicesController) generateNextChildKeyNumber(ctx context.Context) (int, error) {
	seq := SyntheticDeviceSequence{}

	qry := fmt.Sprintf("SELECT nextval('%s.synthetic_devices_serial_sequence');", sdc.Settings.DB.Name)
	err := queries.Raw(qry).Bind(ctx, sdc.DBS().Reader, &seq)
	if err != nil {
		return 0, err
	}

	return seq.NextVal, nil
}

func (sdc *SyntheticDevicesController) getEIP712Burn(vehicleNode, syntheticDeviceNode int64) *signer.TypedData {
	return &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"BurnSyntheticDeviceSign": []signer.Type{
				{Name: "vehicleNode", Type: "uint256"},
				{Name: "syntheticDeviceNode", Type: "uint256"},
			},
		},
		PrimaryType: "BurnSyntheticDeviceSign",
		Domain: signer.TypedDataDomain{
			Name:              "DIMO",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(sdc.Settings.DIMORegistryChainID),
			VerifyingContract: sdc.Settings.DIMORegistryAddr,
		},
		Message: signer.TypedDataMessage{
			"vehicleNode":         math.NewHexOrDecimal256(vehicleNode),
			"syntheticDeviceNode": math.NewHexOrDecimal256(syntheticDeviceNode),
		},
	}
}
