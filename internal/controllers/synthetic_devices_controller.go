package controllers

import (
	"context"
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
	"github.com/savsgio/gotils/bytes"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type SyntheticDevicesController struct {
	Settings     *config.Settings
	DBS          db.Store
	log          *zerolog.Logger
	deviceDefSvc services.DeviceDefinitionService
	usersClient  pb.UserServiceClient
}

type SyntheticDeviceMintingRequest struct {
	VehicleNode int `json:"vehicleNode"`
	Credentials struct {
		AuthorizationCode string `json:"authorizationCode"`
	} `json:"credentials"`
	OwnerSignature string `json:"ownerSignature"`
}

func NewVirtualDeviceController(
	settings *config.Settings, dbs db.Store, logger *zerolog.Logger, deviceDefSvc services.DeviceDefinitionService, usersClient pb.UserServiceClient,
) SyntheticDevicesController {
	return SyntheticDevicesController{
		Settings:     settings,
		DBS:          dbs,
		log:          logger,
		usersClient:  usersClient,
		deviceDefSvc: deviceDefSvc,
	}
}

func (vc *SyntheticDevicesController) getVirtualDeviceMintPayload(integrationID, vehicleNode int64) *signer.TypedData {
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

func (vc *SyntheticDevicesController) verifyUserAddressAndNFTExist(ctx context.Context, user *pb.User, vehicleNode int64, integrationNode string) error {
	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "User does not have an Ethereum address on file.")
	}

	vnID := types.NewNullDecimal(decimal.New(vehicleNode, 0))
	vehicleNFT, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(vnID),
		models.VehicleNFTWhere.OwnerAddress.EQ(null.BytesFrom(common.HexToAddress(*user.EthereumAddress).Bytes())),
	).Exists(ctx, vc.DBS.DBS().Reader)
	if err != nil {
		vc.log.Error().Err(err).Int64("vehicleNode", vehicleNode).Str("integrationNode", integrationNode).Msg("Could not fetch minting payload for device")
		return fiber.NewError(fiber.StatusInternalServerError, "error generating device mint payload")
	}

	if !vehicleNFT {
		return fiber.NewError(fiber.StatusNotFound, "user does not own vehicle node")
	}

	return nil
}

// GetSyntheticDeviceMintingMessage godoc
// @Description Gets the synthetic device minting payload for the user to sign and submit.
// @Tags        integrations
// @Produce     json
// @Param       integrationNode path int true "integration node id"
// @Param       vehicleNode path int true "vehicle node id"
// @Success     200 {array} signer.TypedData
// @Router      synthetic/device/mint/:integrationNode/:vehicleNode [get]
func (vc *SyntheticDevicesController) GetSyntheticDeviceMintingMessage(c *fiber.Ctx) error {
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

	response := vc.getVirtualDeviceMintPayload(int64(integration.TokenId), vid)

	return c.JSON(response)
}

// MintSyntheticDevice godoc
// @Description Validate and submit to the blockchain a synthetic device minting request.
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

	req := &SyntheticDeviceMintingRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request.")
	}

	signature := common.FromHex(req.OwnerSignature)
	if len(signature) != 65 {
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

	payloadVerified := bytes.Equal(crypto.PubkeyToAddress(*pubRaw).Bytes(), userAddr.Bytes())

	if !payloadVerified {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signature provided")
	}

	return c.Send([]byte("signature is valid"))
}
