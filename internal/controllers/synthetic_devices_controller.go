package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

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
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
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
	smartcarClient services.SmartcarClient
	teslaService   services.TeslaService
	cipher         shared.Cipher
}

type MintSyntheticDeviceRequest struct {
	Credentials    credentials `json:"credentials"`
	OwnerSignature string      `json:"ownerSignature"`
}

type credentials struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	ExternalID   string `json:"externalId"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirectUri"`
}

type SyntheticDeviceSequence struct {
	NextVal int `boil:"nextval"`
}

func NewSyntheticDevicesController(
	settings *config.Settings, dbs func() *db.ReaderWriter, logger *zerolog.Logger, deviceDefSvc services.DeviceDefinitionService, usersClient pb.UserServiceClient, walletSvc services.SyntheticWalletInstanceService, registryClient registry.Client, smartcarClient services.SmartcarClient, teslaSvc services.TeslaService, cipher shared.Cipher,

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
		teslaService:   teslaSvc,
		cipher:         cipher,
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

type BurnSyntheticDeviceRequest struct {
	OwnerSignature string `json:"ownerSignature"`
}

func (sdc *SyntheticDevicesController) verifyUserAddressAndNFTExist(ctx context.Context, user *pb.User, vehicleNode int64) (*models.VehicleNFT, error) {
	if user.EthereumAddress == nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "User does not have an Ethereum address.")
	}

	uadd := common.HexToAddress(*user.EthereumAddress)

	vnID := types.NewNullDecimal(decimal.New(vehicleNode, 0))
	vehicleNFT, err := models.VehicleNFTS(models.VehicleNFTWhere.TokenID.EQ(vnID)).One(ctx, sdc.DBS().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fiber.NewError(fiber.StatusNotFound, "No vehicle with that id found.")
		}
		return nil, err
	}

	vown := common.BytesToAddress(vehicleNFT.OwnerAddress.Bytes)
	if uadd != vown {
		return nil, fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("Vehicle owned by another address, %s.", vown))
	}

	if !vehicleNFT.UserDeviceID.Valid {
		return nil, fiber.NewError(fiber.StatusConflict, "Vehicle deleted.")
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
// @Router 	    /synthetic/device/mint/{integrationNode}/{vehicleNode} [get]
func (sdc *SyntheticDevicesController) GetSyntheticDeviceMintingPayload(c *fiber.Ctx) error {
	rawIntegrationNode := c.Params("integrationNode")
	vehicleNode := c.Params("vehicleNode")
	userID := helpers.GetUserID(c)

	integrationNode, err := strconv.ParseUint(rawIntegrationNode, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid integrationNode provided")
	}

	user, err := sdc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		sdc.log.Debug().Err(err).Msg("error occurred when fetching user")
		return helpers.GrpcErrorToFiber(err, "error occurred when fetching user")
	}

	vid, err := strconv.ParseInt(vehicleNode, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid vehicleNode provided")
	}

	if _, err := sdc.verifyUserAddressAndNFTExist(c.Context(), user, vid); err != nil {
		return err
	}

	integration, err := sdc.deviceDefSvc.GetIntegrationByTokenID(c.Context(), integrationNode)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get integration")
	}

	response := sdc.getEIP712Mint(int64(integration.TokenId), vid)

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
func (sdc *SyntheticDevicesController) MintSyntheticDevice(c *fiber.Ctx) error {
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

	integration, err := sdc.deviceDefSvc.GetIntegrationByTokenID(c.Context(), integrationNode)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get integration")
	}

	if integration.Vendor == constants.TeslaVendor && req.Credentials.AccessToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid access token")
	}

	if integration.Vendor == constants.SmartCarVendor && req.Credentials.Code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid authorization code")
	}

	user, err := sdc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "error occurred when fetching user")
	}

	vid, err := strconv.ParseInt(vehicleNode, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid vehicle id provided")
	}

	vNFT, err := sdc.verifyUserAddressAndNFTExist(c.Context(), user, vid)
	if err != nil {
		return err
	}

	userAddr := common.HexToAddress(*user.EthereumAddress)
	rawPayload := sdc.getEIP712Mint(int64(integration.TokenId), vid)

	tdHash, _, err := signer.TypedDataAndHash(*rawPayload)
	if err != nil {
		sdc.log.Err(err).Msg("Error occurred creating hash of payload")
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't verify signature.")
	}

	ownerSignature := common.FromHex(req.OwnerSignature)
	recAddr, err := helpers.Ecrecover(tdHash, ownerSignature)
	if err != nil {
		sdc.log.Err(err).Msg("unable to validate signature")
		return err
	}

	if recAddr != userAddr {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signature.")
	}

	udai, err := sdc.generateUDAI(c.Context(), req.Credentials, vNFT.UserDeviceID.String, integration)
	if err != nil {
		return err
	}

	childKeyNumber, err := sdc.generateNextChildKeyNumber(c.Context())
	if err != nil {
		sdc.log.Err(err).Msg("failed to generate sequence from database")
		return fiber.NewError(fiber.StatusInternalServerError, "synthetic device minting request failed")
	}

	requestID := ksuid.New().String()

	syntheticDeviceAddr, err := sdc.sendSyntheticDeviceMintPayload(c.Context(), requestID, tdHash, int(vid), integration.TokenId, ownerSignature, childKeyNumber)
	if err != nil {
		sdc.log.Err(err).Msg("synthetic device minting request failed")
		return fiber.NewError(fiber.StatusInternalServerError, "synthetic device minting request failed")
	}

	tx, err := sdc.DBS().Writer.DB.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}

	if err := udai.Insert(c.Context(), tx, boil.Infer()); err != nil {
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

	vnID := types.NewDecimal(decimal.New(vid, 0))
	syntheticDevice := &models.SyntheticDevice{
		VehicleTokenID:     vnID,
		IntegrationTokenID: types.NewDecimal(decimal.New(int64(integrationNode), 0)),
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

func (sdc *SyntheticDevicesController) generateUDAI(ctx context.Context, creds credentials, userDeviceID string, integration *grpc.Integration) (*models.UserDeviceAPIIntegration, error) {
	udi := models.UserDeviceAPIIntegration{
		IntegrationID: integration.Id,
		UserDeviceID:  userDeviceID,
		Status:        models.UserDeviceAPIIntegrationStatusPending,
		TaskID:        null.StringFrom(ksuid.New().String()),
	}

	switch integration.Vendor {
	case constants.SmartCarVendor:
		token, err := sdc.exchangeSmartCarCode(ctx, creds.Code, creds.RedirectURI)
		if err != nil {
			return nil, err
		}

		encAccess, err := sdc.cipher.Encrypt(token.Access)
		if err != nil {
			return nil, opaqueInternalError
		}

		encRefresh, err := sdc.cipher.Encrypt(token.Refresh)
		if err != nil {
			return nil, opaqueInternalError
		}

		udi.AccessToken = null.StringFrom(encAccess)
		udi.AccessExpiresAt = null.TimeFrom(token.AccessExpiry)
		udi.RefreshToken = null.StringFrom(encRefresh)

		externalID, err := sdc.smartcarClient.GetExternalID(ctx, token.Access)
		if err != nil {
			return nil, err
		}

		endpoints, err := sdc.smartcarClient.GetEndpoints(ctx, token.Access, externalID)
		if err != nil {
			sdc.log.Err(err).Msg("Failed to retrieve permissions from Smartcar.")
			return nil, err
		}

		meta := services.UserDeviceAPIIntegrationsMetadata{
			SmartcarEndpoints: endpoints,
		}

		mb, _ := json.Marshal(meta)
		udi.Metadata = null.JSONFrom(mb)
		udi.ExternalID = null.StringFrom(externalID)
	case constants.TeslaVendor:
		teslaID, err := strconv.Atoi(creds.ExternalID)
		if err != nil {
			sdc.log.Err(err).Msgf("unable to parse external id %q as integer", creds.ExternalID)
			return nil, err
		}

		v, err := sdc.teslaService.GetVehicle(creds.AccessToken, teslaID)
		if err != nil {
			sdc.log.Err(err).Msg("unable to retrieve vehicle from Tesla")
			return nil, err
		}

		encAccess, err := sdc.cipher.Encrypt(creds.AccessToken)
		if err != nil {
			return nil, opaqueInternalError
		}
		encRefresh, err := sdc.cipher.Encrypt(creds.RefreshToken)
		if err != nil {
			return nil, opaqueInternalError
		}
		udi.AccessToken = null.StringFrom(encAccess)
		udi.RefreshToken = null.StringFrom(encRefresh)

		meta := services.UserDeviceAPIIntegrationsMetadata{
			Commands: &services.UserDeviceAPIIntegrationsMetadataCommands{
				Enabled: []string{"doors/unlock", "doors/lock", "trunk/open", "frunk/open", "charge/limit"},
			},
			TeslaVehicleID: v.ID,
		}

		mb, _ := json.Marshal(meta)

		udi.ExternalID = null.StringFrom(creds.ExternalID)
		udi.AccessExpiresAt = null.TimeFrom(time.Now().Add(time.Duration(creds.ExpiresIn) * time.Second))
		udi.Metadata = null.JSONFrom(mb)
	default:
		return nil, fmt.Errorf("unrecognized vendor %s", integration.Vendor)
	}

	return &udi, nil
}

func (sdc *SyntheticDevicesController) exchangeSmartCarCode(ctx context.Context, authCode, redirectURI string) (*smartcar.Token, error) {
	token, err := sdc.smartcarClient.ExchangeCode(ctx, authCode, redirectURI)
	if err != nil {
		var scErr *services.SmartcarError
		if errors.As(err, &scErr) {
			sdc.log.Err(err).Str("function", "syntheticDeviceController.exchangeSmartCarCode").Msgf("Failed exchanging Authorization code. Status code %d, request id %s`.", scErr.Code, scErr.RequestID)
		} else {
			sdc.log.Err(err).Str("function", "syntheticDeviceController.exchangeSmartCarCode").Msg("Failed to exchange authorization code with Smartcar.")
		}

		return nil, errors.New("failed to exchange authorization code with smartcar")
	}

	return token, nil
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
