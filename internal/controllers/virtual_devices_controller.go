package controllers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"strconv"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type VirtualDeviceController struct {
	Settings     *config.Settings
	DBS          func() *db.ReaderWriter
	log          *zerolog.Logger
	integSvc     services.DeviceDefinitionIntegrationService
	deviceDefSvc services.DeviceDefinitionService
	usersClient  pb.UserServiceClient
}

type SignVirtualDeviceMintingPayloadRequest struct {
	VehicleNode int `json:"vehicleNode"`
	Credentials struct {
		AuthorizationCode string `json:"authorizationCode"`
	} `json:"credentials"`
	OwnerSignature string `json:"ownerSignature"`
}

func NewVirtualDeviceController(
	settings *config.Settings, dbs func() *db.ReaderWriter, logger *zerolog.Logger, integSvc services.DeviceDefinitionIntegrationService, deviceDefSvc services.DeviceDefinitionService, usersClient pb.UserServiceClient,
) VirtualDeviceController {
	return VirtualDeviceController{
		Settings:    settings,
		DBS:         dbs,
		log:         logger,
		integSvc:    integSvc,
		usersClient: usersClient,
	}
}

func (vc *VirtualDeviceController) getVirtualDeviceMintPayload(tokenID int64, vehicleNode int64) *signer.TypedData {
	return &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"MintDevice": {
				{Name: "integrationNode", Type: "uint256"},
				{Name: "vehicleNode", Type: "uint256"},
			},
		},
		PrimaryType: "MintVirtualDeviceSign",
		Domain: signer.TypedDataDomain{
			Name:              "DIMO",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(vc.Settings.DeviceMintingChainID),
			VerifyingContract: vc.Settings.DeviceMintingVerifyingContract,
		},
		Message: signer.TypedDataMessage{
			"integrationNode": math.NewHexOrDecimal256(tokenID),
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
	vehicleNode := c.Query("vehicle_id")
	userID := helpers.GetUserID(c)

	uTokenID, err := strconv.ParseUint(tokenID, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid tokenID provided")
	}

	user, err := vc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "could not get user info")
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
		return fiber.NewError(fiber.StatusBadRequest, "invalid vehicleNode provided")
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
	vehicleNode := c.Query("vehicle_id")
	userID := helpers.GetUserID(c)

	req := &SignVirtualDeviceMintingPayloadRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request.")
	}

	signature, err := hex.DecodeString(req.OwnerSignature)
	if err != nil {
		log.Println("invalid hex string", err)
		return fiber.NewError(fiber.StatusBadRequest, "invalid signature provided")
	}

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
		return helpers.GrpcErrorToFiber(err, "could not get user info")
	}

	vid, err := strconv.ParseInt(vehicleNode, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid vehicle id provided")
	}

	rawPayload := vc.getVirtualDeviceMintPayload(int64(integration.TokenId), vid)
	payloadJSON, err := json.Marshal(rawPayload)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request.")
	}

	hash := crypto.Keccak256Hash(payloadJSON)

	sigPublicKey, err := crypto.Ecrecover(hash.Bytes(), signature)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Couldn't not verify signature.")
	}

	pubKey, err := crypto.UnmarshalPubkey(sigPublicKey)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Couldn't not verify signature.")
	}

	payloadVerified := bytes.Equal(crypto.PubkeyToAddress(*pubKey).Bytes(), common.HexToAddress(*user.EthereumAddress).Bytes())
	if !payloadVerified {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signature provided")
	}
	return c.Send([]byte("signature is valid"))
}
