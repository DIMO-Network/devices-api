package controllers

import (
	"context"
	"math/big"
	"math/rand"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
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
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/savsgio/gotils/bytes"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
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
	VehicleNode int `json:"vehicleNode"`
	Credentials struct {
		AuthorizationCode string `json:"authorizationCode"`
	} `json:"credentials"`
	OwnerSignature string `json:"ownerSignature"`
}

func NewSyntheticDevicesController(
	settings *config.Settings, dbs func() *db.ReaderWriter, logger *zerolog.Logger, deviceDefSvc services.DeviceDefinitionService, usersClient pb.UserServiceClient, walletSvc services.SyntheticWalletInstanceService, registryClient registry.Client,
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

func (vc *SyntheticDevicesController) getEIP712(integrationID, vehicleNode int64) *signer.TypedData {
	return &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			// Need to keep this name until the contract changes.
			"MintVirtualDeviceSign": []signer.Type{
				{Name: "integrationNode", Type: "uint256"},
				{Name: "vehicleNode", Type: "uint256"},
			},
		},
		PrimaryType: "MintVirtualDeviceSign",
		Domain: signer.TypedDataDomain{
			Name:              "DIMO",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(vc.Settings.DIMORegistryChainID),
			VerifyingContract: vc.Settings.DIMORegistryAddr,
		},
		Message: signer.TypedDataMessage{
			"integrationNode": math.NewHexOrDecimal256(integrationID),
			"vehicleNode":     math.NewHexOrDecimal256(vehicleNode),
		},
	}
}

func (vc *SyntheticDevicesController) verifyUserAddressAndNFTExist(ctx context.Context, user *pb.User, vehicleNode int64, integrationNode string) error {
	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "User does not have an Ethereum address on file.")
	}

	vnID := types.NewNullDecimal(decimal.New(vehicleNode, 0))
	vehicleNFT, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(vnID),
		models.VehicleNFTWhere.OwnerAddress.EQ(null.BytesFrom(common.HexToAddress(*user.EthereumAddress).Bytes())),
	).Exists(ctx, vc.DBS().Reader)
	if err != nil {
		vc.log.Error().Err(err).Int64("vehicleNode", vehicleNode).Str("integrationNode", integrationNode).Msg("Could not fetch minting payload for device")
		return fiber.NewError(fiber.StatusInternalServerError, "error generating device mint payload")
	}

	if !vehicleNFT {
		return fiber.NewError(fiber.StatusNotFound, "user does not own vehicle node")
	}

	return nil
}

// GetSyntheticDeviceMintingPayload godoc
// @Description Produces the payload that the user signs and submits to mint a synthetic device for
// @Description the given vehicle and integration.
// @Tags        integrations
// @Produce     json
// @Param       integrationNode path int true "token ID"
// @Param       vehicleNode path int true "vehicle ID"
// @Success     200 {array} signer.TypedData
// @Router      synthetic/device/mint/:integrationNode/:vehicleNode [get]
func (vc *SyntheticDevicesController) GetSyntheticDeviceMintingPayload(c *fiber.Ctx) error {
	rawIntegrationNode := c.Params("integrationNode")
	vehicleNode := c.Params("vehicleNode")
	userID := helpers.GetUserID(c)

	integrationNode, err := strconv.ParseUint(rawIntegrationNode, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid integrationNode provided")
	}

	user, err := vc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		vc.log.Debug().Err(err).Msg("error occurred when fetching user")
		return helpers.GrpcErrorToFiber(err, "error occurred when fetching user")
	}

	vid, err := strconv.ParseInt(vehicleNode, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid vehicleNode provided")
	}

	if err = vc.verifyUserAddressAndNFTExist(c.Context(), user, vid, rawIntegrationNode); err != nil {
		return err
	}

	integration, err := vc.deviceDefSvc.GetIntegrationByTokenID(c.Context(), integrationNode)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get integration")
	}

	response := vc.getEIP712(int64(integration.TokenId), vid)

	return c.JSON(response)
}

// MintSyntheticDevice godoc
// @Description Submit a metadata
// @Tags        integrations
// @Produce     json
// @Param       integrationNode path int true "token ID"
// @Param       vehicleNode path int true "vehicle ID"
// @Success     200 {array}
// @Router      synthetic/device/mint/:integrationNode/:vehicleNode [post]
func (vc *SyntheticDevicesController) MintSyntheticDevice(c *fiber.Ctx) error {
	rawIntegrationNode := c.Params("integrationNode")
	vehicleNode := c.Params("vehicleNode")
	userID := helpers.GetUserID(c)

	req := &MintSyntheticDeviceRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request.")
	}

	ownerSignature := common.FromHex(req.OwnerSignature)
	if len(ownerSignature) != 65 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid signature provided")
	}

	integrationNode, err := strconv.ParseUint(rawIntegrationNode, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid integrationNode provided")
	}

	user, err := vc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "error occurred when fetching user")
	}

	vid, err := strconv.ParseInt(vehicleNode, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid vehicle id provided")
	}

	if err = vc.verifyUserAddressAndNFTExist(c.Context(), user, vid, rawIntegrationNode); err != nil {
		return err
	}

	integration, err := vc.deviceDefSvc.GetIntegrationByTokenID(c.Context(), integrationNode)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get integration")
	}

	userAddr := common.HexToAddress(*user.EthereumAddress)

	rawPayload := vc.getEIP712(int64(integration.TokenId), vid)

	hash, _, err := signer.TypedDataAndHash(*rawPayload)
	if err != nil {
		vc.log.Err(err).Msg("Error occurred creating has of payload")
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't verify signature.")
	}

	ownerSignature[64] -= 27

	pub, err := crypto.Ecrecover(hash, ownerSignature)
	if err != nil {
		vc.log.Err(err).Msg("Error occurred while trying to recover public key from signature")
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't verify signature.")
	}

	pubRaw, err := crypto.UnmarshalPubkey(pub)
	if err != nil {
		vc.log.Err(err).Msg("Error occurred marshalling recovered public public key")
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't verify signature.")
	}

	payloadVerified := bytes.Equal(crypto.PubkeyToAddress(*pubRaw).Bytes(), userAddr.Bytes())
	if !payloadVerified {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signature provided")
	}

	err = vc.sendVirtualDeviceMintPayload(c.Context(), hash, req.VehicleNode, integration.TokenId, ownerSignature)
	if err != nil {
		vc.log.Err(err).Msg("virtual device minting request failed")
		return fiber.NewError(fiber.StatusInternalServerError, "virtual device minting request failed")
	}

	return c.Send([]byte("virtual device mint request successful"))
}

func (vc *SyntheticDevicesController) sendVirtualDeviceMintPayload(ctx context.Context, hash []byte, vehicleNode int, intTokenID uint64, ownerSignature []byte) error {
	childKeyNumber := generateRandomNumber()

	syntheticDeviceAddr, err := vc.walletSvc.GetAddress(ctx, uint32(childKeyNumber))
	if err != nil {
		vc.log.Err(err).
			Str("function-name", "SyntheticWallet.GetAddress").
			Int("childKeyNumber", childKeyNumber).
			Msg("Error occurred getting synthetic wallet address")
		return err
	}

	virtSig, err := vc.walletSvc.SignHash(ctx, uint32(childKeyNumber), hash)
	if err != nil {
		vc.log.Err(err).
			Str("function-name", "SyntheticWallet.SignHash").
			Bytes("Hash", hash).
			Int("childKeyNumber", childKeyNumber).
			Msg("Error occurred signing message hash")
		return err
	}

	requestID := ksuid.New().String()

	vNode := new(big.Int).SetInt64(int64(vehicleNode))
	mvt := contracts.MintVirtualDeviceInput{
		IntegrationNode:   new(big.Int).SetUint64(intTokenID),
		VehicleNode:       vNode,
		VehicleOwnerSig:   ownerSignature,
		VirtualDeviceAddr: common.BytesToAddress(syntheticDeviceAddr),
		VirtualDeviceSig:  virtSig,
	}

	return vc.registryClient.MintVirtualDeviceSign(requestID, mvt)
}

func generateRandomNumber() int {
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 1000
	return rand.Intn(max-min+1) + min
}
