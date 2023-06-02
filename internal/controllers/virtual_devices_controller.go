package controllers

import (
	"context"
	"log"
	"math/big"
	"math/rand"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	registryContract "github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/Shopify/sarama"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type VirtualDeviceController struct {
	Settings      *config.Settings
	DBS           func() *db.ReaderWriter
	log           *zerolog.Logger
	integSvc      services.DeviceDefinitionIntegrationService
	deviceDefSvc  services.DeviceDefinitionService
	usersClient   pb.UserServiceClient
	virtDeviceSvc services.VirtualDeviceInstanceService
	producer      sarama.SyncProducer
}

type SignVirtualDeviceMintingPayloadRequest struct {
	VehicleNode int `json:"vehicleNode"`
	Credentials struct {
		AuthorizationCode string `json:"authorizationCode"`
	} `json:"credentials"`
	OwnerSignature string `json:"ownerSignature"`
}

func NewVirtualDeviceController(
	settings *config.Settings, dbs func() *db.ReaderWriter, logger *zerolog.Logger, integSvc services.DeviceDefinitionIntegrationService, deviceDefSvc services.DeviceDefinitionService, usersClient pb.UserServiceClient, virtDeviceSvc services.VirtualDeviceInstanceService, producer sarama.SyncProducer,
) VirtualDeviceController {
	return VirtualDeviceController{
		Settings:      settings,
		DBS:           dbs,
		log:           logger,
		integSvc:      integSvc,
		usersClient:   usersClient,
		deviceDefSvc:  deviceDefSvc,
		virtDeviceSvc: virtDeviceSvc,
		producer:      producer,
	}
}

func (vc *VirtualDeviceController) getVirtualDeviceMintPayload(integrationID int64, vehicleNode int64) *signer.TypedData {
	return &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
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

// GetVirtualDeviceMintingPayload godoc
// @Description gets the payload for to mint virtual device given an integration token ID
// @Tags        integrations
// @Produce     json
// @Success     200 {array} signer.TypedData
// @Router      /integration/:tokenID/mint-virtual-device [get]
func (vc *VirtualDeviceController) GetVirtualDeviceMintingPayload(c *fiber.Ctx) error {
	tokenID := c.Params("tokenID")
	vehicleNode := c.Params("vehicleID")
	userID := helpers.GetUserID(c)

	uTokenID, err := strconv.ParseUint(tokenID, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid tokenID provided")
	}

	user, err := vc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		vc.log.Debug().Err(err).Msg("error occurred when fetching user")
		return helpers.GrpcErrorToFiber(err, "error occurred when fetching user")
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "User does not have an Ethereum address on file.")
	}

	integration, err := vc.deviceDefSvc.GetIntegrationByTokenID(c.Context(), uTokenID)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get integration")
	}

	vid, err := strconv.ParseInt(vehicleNode, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid vehicleNode provided")
	}
	vnID := types.NewNullDecimal(decimal.New(vid, 0))
	vehicleNFT, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(vnID),
		models.VehicleNFTWhere.OwnerAddress.EQ(null.BytesFrom(common.HexToAddress(*user.EthereumAddress).Bytes())),
	).Exists(c.Context(), vc.DBS().Reader)
	if err != nil {
		vc.log.Error().Err(err).Str("vehicleNode", vehicleNode).Str("tokenID", tokenID).Msg("Could not fetch minting payload for device")
		return fiber.NewError(fiber.StatusInternalServerError, "error generating device mint payload")
	}

	if !vehicleNFT {
		return fiber.NewError(fiber.StatusNotFound, "user does not own vehicle node")
	}

	response := vc.getVirtualDeviceMintPayload(int64(integration.TokenId), vid)

	return c.JSON(response)
}

// SignVirtualDeviceMintingPayload godoc
// @Description validate signed signature for vehicle minting
// @Tags        integrations
// @Produce     json
// @Success     200 {array}
// @Router      /integration/:tokenID/mint-virtual-device [post]
func (vc *VirtualDeviceController) SignVirtualDeviceMintingPayload(c *fiber.Ctx) error {
	tokenID := c.Params("tokenID")
	vehicleNode := c.Params("vehicleID")
	userID := helpers.GetUserID(c)

	req := &SignVirtualDeviceMintingPayloadRequest{}
	if err := c.BodyParser(req); err != nil {
		log.Println(err)
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request.")
	}

	signature := common.FromHex(req.OwnerSignature)
	if len(signature) != 65 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid signature provided")
	}

	uTokenID, err := strconv.ParseUint(tokenID, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid tokenID provided")
	}

	integration, err := vc.deviceDefSvc.GetIntegrationByTokenID(c.Context(), uTokenID)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get integration")
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

	rawPayload := vc.getVirtualDeviceMintPayload(int64(integration.TokenId), vid)

	hash, _, err := signer.TypedDataAndHash(*rawPayload)
	if err != nil {
		vc.log.Err(err).Msg("Error occurred creating has of payload")
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't verify signature.")
	}

	signature[64] -= 27

	pub, err := crypto.Ecrecover(hash, signature)
	if err != nil {
		vc.log.Err(err).Msg("Error occurred while trying to recover public key from signature")
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't verify signature.")
	}

	pubRaw, err := crypto.UnmarshalPubkey(pub)
	if err != nil {
		vc.log.Err(err).Msg("Error occurred marshalling recovered public public key")
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't verify signature.")
	}

	addr := crypto.PubkeyToAddress(*pubRaw)
	userAddr := common.HexToAddress(*user.EthereumAddress)
	payloadVerified := addr == userAddr

	if !payloadVerified {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signature provided")
	}

	err = vc.sendVirtualDeviceMintPayload(c.Context(), hash, req.VehicleNode, integration.TokenId, userAddr, signature)
	if err != nil {
		vc.log.Err(err).Msg("virtual device minting request failed")
		return fiber.NewError(fiber.StatusInternalServerError, "virtual device minting request failed")
	}

	return c.Send([]byte("virtual device mint request successful"))
}

func (vc *VirtualDeviceController) sendVirtualDeviceMintPayload(ctx context.Context, hash []byte, vehicleNode int, intTokenID uint64, userAddr common.Address, signature []byte) error {
	childKeyNumber := generateRandomNumber()
	virtSig, err := vc.virtDeviceSvc.SignHash(ctx, uint32(childKeyNumber), hash)
	if err != nil {
		vc.log.Err(err).
			Str("function-name", "SyntheticWallet.SignHash").
			Bytes("Hash", hash).
			Int("childKeyNumber", childKeyNumber).
			Msg("Error occurred signing message hash")
		return err
	}

	client := registry.Client{
		Producer:     vc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(vc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(vc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	requestID := ksuid.New().String()

	vNode := new(big.Int).SetInt64(int64(vehicleNode))
	mvt := registryContract.MintVirtualDeviceInput{
		IntegrationNode:   new(big.Int).SetUint64(intTokenID),
		VehicleNode:       vNode,
		VehicleOwnerSig:   signature,
		VirtualDeviceAddr: userAddr,
		VirtualDeviceSig:  virtSig,
	}

	client.MintVirtualDeviceSign(requestID, mvt)
	return nil
}

func generateRandomNumber() int {
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 1000
	return rand.Intn(max-min+1) + min
}
