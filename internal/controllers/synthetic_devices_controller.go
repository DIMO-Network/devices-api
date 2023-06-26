package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
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
	smartcar "github.com/smartcar/go-sdk"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
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
	smartcarClient services.SmartcarClient
	cipher         shared.Cipher
}

type MintSyntheticDeviceRequest struct {
	VehicleNode int `json:"vehicleNode"`
	Credentials struct {
		AuthorizationCode string `json:"authorizationCode"`
	} `json:"credentials"`
	OwnerSignature string `json:"ownerSignature"`
}

type SyntheticDeviceSequence struct {
	NextVal int `boil:"nextval"`
}

func NewSyntheticDevicesController(
	settings *config.Settings, dbs func() *db.ReaderWriter, logger *zerolog.Logger, deviceDefSvc services.DeviceDefinitionService, usersClient pb.UserServiceClient, walletSvc services.SyntheticWalletInstanceService, registryClient registry.Client, smartcarClient services.SmartcarClient, cipher shared.Cipher,

) SyntheticDevicesController {
	return SyntheticDevicesController{
		Settings:       settings,
		DBS:            dbs,
		log:            logger,
		usersClient:    usersClient,
		deviceDefSvc:   deviceDefSvc,
		walletSvc:      walletSvc,
		registryClient: registryClient,
		smartcarClient: smartcarClient,
		cipher:         cipher,
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
			"MintSyntheticDeviceSign": []signer.Type{
				{Name: "integrationNode", Type: "uint256"},
				{Name: "vehicleNode", Type: "uint256"},
			},
		},
		PrimaryType: "MintSyntheticDeviceSign",
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

func (vc *SyntheticDevicesController) verifyUserAddressAndNFTExist(ctx context.Context, user *pb.User, vehicleNode int64, integrationNode string) (*models.VehicleNFT, error) {
	if user.EthereumAddress == nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "User does not have an Ethereum address on file.")
	}

	vnID := types.NewNullDecimal(decimal.New(vehicleNode, 0))
	vehicleNFT, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(vnID),
		models.VehicleNFTWhere.OwnerAddress.EQ(null.BytesFrom(common.HexToAddress(*user.EthereumAddress).Bytes())),
	).One(ctx, vc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fiber.NewError(fiber.StatusNotFound, "user does not own vehicle node")
		}
		vc.log.Error().Err(err).Int64("vehicleNode", vehicleNode).Str("integrationNode", integrationNode).Msg("Could not fetch minting payload for device")
		return nil, fiber.NewError(fiber.StatusInternalServerError, "error generating device mint payload")
	}

	return vehicleNFT, nil
}

// GetSyntheticDeviceMintingPayload godoc
// @Description Produces the payload that the user signs and submits to mint a synthetic device for
// @Description the given vehicle and integration.
// @Tags        integrations
// @Produce     json
// @Param       integrationNode path int true "token ID"
// @Param       vehicleNode path int true "vehicle ID"
// @Success     200 {array} signer.TypedData
// @Router /synthetic/device/mint/{integrationNode}/{vehicleNode} [get]
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

	if _, err := vc.verifyUserAddressAndNFTExist(c.Context(), user, vid, rawIntegrationNode); err != nil {
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
// @Success     204
// @Router      /synthetic/device/mint/{integrationNode}/{vehicleNode} [post]
func (vc *SyntheticDevicesController) MintSyntheticDevice(c *fiber.Ctx) error {
	rawIntegrationNode := c.Params("integrationNode")
	vehicleNode := c.Params("vehicleNode")
	userID := helpers.GetUserID(c)

	req := &MintSyntheticDeviceRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request.")
	}

	integrationNode, err := strconv.ParseUint(rawIntegrationNode, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid integrationNode provided")
	}

	integration, err := vc.deviceDefSvc.GetIntegrationByTokenID(c.Context(), integrationNode)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get integration")
	}

	if integration.Vendor == constants.SmartCarVendor && req.Credentials.AuthorizationCode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "please provide authorization code")
	}

	ownerSignature := common.FromHex(req.OwnerSignature)
	if len(ownerSignature) != 65 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid signature provided")
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

	vNFT, err := vc.verifyUserAddressAndNFTExist(c.Context(), user, vid, rawIntegrationNode)
	if err != nil {
		return err
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

	ownerSignature[64] += 27

	childKeyNumber, err := vc.generateNextChildKeyNumber(c.Context())
	if err != nil {
		vc.log.Err(err).Msg("failed to generate sequence from database")
		return fiber.NewError(fiber.StatusInternalServerError, "synthetic device minting request failed")
	}

	syntheticDeviceAddr, err := vc.sendSyntheticDeviceMintPayload(c.Context(), hash, req.VehicleNode, integration.TokenId, ownerSignature, childKeyNumber)
	if err != nil {
		vc.log.Err(err).Msg("synthetic device minting request failed")
		return fiber.NewError(fiber.StatusInternalServerError, "synthetic device minting request failed")
	}

	tx, err := vc.DBS().Writer.DB.BeginTx(c.Context(), nil)
	if err != nil {
		vc.log.Err(err).Msg("error creating database transaction")
		return fiber.NewError(fiber.StatusInternalServerError, "synthetic device minting request failed")
	}

	if err := vc.handleDeviceAPIIntegrationCreation(c.Context(), tx, req, vNFT.UserDeviceID.String, integration); err != nil {
		vc.log.Err(err).Str("UserDeviceID", vNFT.UserDeviceID.String).Msg("error creating userDeviceAPiIntegration record")
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	requestID := ksuid.New().String()
	metaReq := &models.MetaTransactionRequest{
		ID:     requestID,
		Status: models.MetaTransactionRequestStatusUnsubmitted,
	}

	if err = metaReq.Insert(c.Context(), tx, boil.Infer()); err != nil {
		vc.log.Err(err).Msg("error occurred creating meta transaction request")
		return fiber.NewError(fiber.StatusInternalServerError, "synthetic device minting request failed")
	}

	vnID := types.NewDecimal(decimal.New(vid, 0))
	syntheticDevice := &models.SyntheticDevice{
		VehicleTokenID:     vnID,
		IntegrationTokenID: types.NewDecimal(decimal.New(int64(integrationNode), 0)),
		WalletChildNumber:  childKeyNumber,
		WalletAddress:      syntheticDeviceAddr,
		MintRequestID:      requestID,
	}

	if err = syntheticDevice.Insert(c.Context(), tx, boil.Infer()); err != nil {
		vc.log.Err(err).Msg("error occurred saving synthetic device")
		return fiber.NewError(fiber.StatusInternalServerError, "synthetic device minting request failed")
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return c.Send([]byte("synthetic device mint request successful"))
}

func (vc *SyntheticDevicesController) sendSyntheticDeviceMintPayload(ctx context.Context, hash []byte, vehicleNode int, intTokenID uint64, ownerSignature []byte, childKeyNumber int) ([]byte, error) {
	syntheticDeviceAddr, err := vc.walletSvc.GetAddress(ctx, uint32(childKeyNumber))
	if err != nil {
		vc.log.Err(err).
			Str("function-name", "SyntheticWallet.GetAddress").
			Int("childKeyNumber", childKeyNumber).
			Msg("Error occurred getting synthetic wallet address")
		return nil, err
	}

	virtSig, err := vc.walletSvc.SignHash(ctx, uint32(childKeyNumber), hash)
	if err != nil {
		vc.log.Err(err).
			Str("function-name", "SyntheticWallet.SignHash").
			Bytes("Hash", hash).
			Int("childKeyNumber", childKeyNumber).
			Msg("Error occurred signing message hash")
		return nil, err
	}

	requestID := ksuid.New().String()

	vNode := new(big.Int).SetInt64(int64(vehicleNode))
	mvt := contracts.MintSyntheticDeviceInput{
		IntegrationNode:     new(big.Int).SetUint64(intTokenID),
		VehicleNode:         vNode,
		VehicleOwnerSig:     ownerSignature,
		SyntheticDeviceAddr: common.BytesToAddress(syntheticDeviceAddr),
		SyntheticDeviceSig:  virtSig,
	}

	return syntheticDeviceAddr, vc.registryClient.MintSyntheticDeviceSign(requestID, mvt)
}

func (vc *SyntheticDevicesController) generateNextChildKeyNumber(ctx context.Context) (int, error) {
	seq := SyntheticDeviceSequence{}

	qry := fmt.Sprintf("SELECT nextval('%s.synthetic_devices_serial_sequence');", vc.Settings.DB.Name)
	err := queries.Raw(qry).Bind(ctx, vc.DBS().Reader, &seq)
	if err != nil {
		return 0, err
	}

	return seq.NextVal, nil
}

func (vc *SyntheticDevicesController) handleDeviceAPIIntegrationCreation(ctx context.Context, tx *sql.Tx, req *MintSyntheticDeviceRequest, userDeviceID string, integration *grpc.Integration) error {
	udi := models.UserDeviceAPIIntegration{
		IntegrationID: integration.Id,
		UserDeviceID:  userDeviceID,
		Status:        models.UserDeviceAPIIntegrationStatusPending,
	}
	switch integration.Vendor {
	case constants.SmartCarVendor:
		token, err := vc.exchangeSmartCarCode(ctx, req.Credentials.AuthorizationCode, "")
		if err != nil {
			return err
		}
		encAccess, err := vc.cipher.Encrypt(token.Access)
		if err != nil {
			return opaqueInternalError
		}
		encRefresh, err := vc.cipher.Encrypt(token.Refresh)
		if err != nil {
			return opaqueInternalError
		}
		udi.AccessToken = null.StringFrom(encAccess)
		udi.AccessExpiresAt = null.TimeFrom(token.AccessExpiry)
		udi.RefreshToken = null.StringFrom(encRefresh)

		externalID, err := vc.smartcarClient.GetExternalID(ctx, token.Access)
		if err != nil {
			return err
		}

		endpoints, err := vc.smartcarClient.GetEndpoints(ctx, token.Access, externalID)
		if err != nil {
			vc.log.Err(err).Msg("Failed to retrieve permissions from Smartcar.")
			return err
		}

		meta := services.UserDeviceAPIIntegrationsMetadata{
			SmartcarEndpoints: endpoints,
		}

		mb, _ := json.Marshal(meta)
		udi.Metadata = null.JSONFrom(mb)
		udi.ExternalID = null.StringFrom(externalID)
	default:
		return nil
	}

	return udi.Insert(ctx, tx, boil.Infer())
}

func (vc *SyntheticDevicesController) exchangeSmartCarCode(ctx context.Context, authCode, redirectURI string) (*smartcar.Token, error) {
	token, err := vc.smartcarClient.ExchangeCode(ctx, authCode, redirectURI)
	if err != nil {
		var scErr *services.SmartcarError
		if errors.As(err, &scErr) {
			vc.log.Err(err).Str("function", "syntheticDeviceController.exchangeSmartCarCode").Msgf("Failed exchanging Authorization code. Status code %d, request id %s`.", scErr.Code, scErr.RequestID)
		} else {
			vc.log.Err(err).Str("function", "syntheticDeviceController.exchangeSmartCarCode").Msg("Failed to exchange authorization code with Smartcar.")
		}

		return nil, errors.New("failed to exchange authorization code with smartcar")
	}

	return token, nil
}
