package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/shared"
	pb_oracle "github.com/DIMO-Network/tesla-oracle/pkg/grpc"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	sig2 "github.com/DIMO-Network/devices-api/internal/contracts/signature"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/ethclient"
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
	walletSvc      services.SyntheticWalletInstanceService
	registryClient registry.Client
	teslaOracle    pb_oracle.TeslaOracleClient
}

type MintSyntheticDeviceRequest struct {
	Signature string `json:"signature" example:"0xc565d38982e1a5004efb5ee390fba0a08bb5e72b3f3e91094c66bc395c324f785425d58d5c1a601372d9c16164e380c63e89f1e0ea95fdefdf7b2854c4f938e81b"`
}

type SyntheticDeviceSequence struct {
	NextVal int `boil:"nextval"`
}

func NewSyntheticDevicesController(
	settings *config.Settings,
	dbs func() *db.ReaderWriter,
	logger *zerolog.Logger,
	deviceDefSvc services.DeviceDefinitionService,
	walletSvc services.SyntheticWalletInstanceService,
	registryClient registry.Client,
	teslaOracle pb_oracle.TeslaOracleClient,
) SyntheticDevicesController {
	return SyntheticDevicesController{
		Settings:       settings,
		DBS:            dbs,
		log:            logger,
		deviceDefSvc:   deviceDefSvc,
		walletSvc:      walletSvc,
		registryClient: registryClient,
		teslaOracle:    teslaOracle,
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

// getEIP712MintV2 produces the "new-style" EIP-712 payload for synthetic device minting, the
// one that uses connection id (which is typically a very large number) instead of integration id
// or node. Unlike in other places, this has a real effect at the byte level because EIP-712
// type hashes include names.
func (sdc *SyntheticDevicesController) getEIP712MintV2(connectionID *big.Int, vehicleNode int64) *signer.TypedData {
	return &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"MintSyntheticDeviceSign": []signer.Type{
				{Name: "connectionId", Type: "uint256"},
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
			"connectionId": math.HexOrDecimal256(*connectionID),
			"vehicleNode":  math.NewHexOrDecimal256(vehicleNode),
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
	userAddr, err := helpers.GetJWTEthAddr(c)
	if err != nil {
		return err
	}

	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	newIntegIDs, ok := syntheticIntegrationKSUIDToOtherIDs[integrationID]
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot mint this integration with devices-api.")
	}

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleTokenSyntheticDevice, models.SyntheticDeviceRels.MintRequest)),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID)),
		qm.Load(models.UserDeviceRels.BurnRequest),
	).One(c.Context(), sdc.DBS().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "No vehicle with that id found.")
		}
		return err
	}

	if ud.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not minted.")
	}

	if burn := ud.R.BurnRequest; burn != nil && burn.Status != models.MetaTransactionRequestStatusFailed {
		return fiber.NewError(fiber.StatusConflict, "Vehicle is being burned.")
	}

	if ownerAddr := common.BytesToAddress(ud.OwnerAddress.Bytes); userAddr != ownerAddr {
		return fiber.NewError(fiber.StatusUnauthorized, fmt.Sprintf("User's address %s does not match vehicle owner %s.", userAddr, ownerAddr))
	}

	if sd := ud.R.VehicleTokenSyntheticDevice; sd != nil {
		if !sd.TokenID.IsZero() {
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("Vehicle already paired with synthetic device %d.", sd.TokenID.Big))
		}
		if sd.R.MintRequest.Status != models.MetaTransactionRequestStatusFailed {
			return fiber.NewError(fiber.StatusConflict, "There is already a synthetic device mint in progress for this vehicle.")
		}
	}

	if len(ud.R.UserDeviceAPIIntegrations) == 0 {
		return fiber.NewError(fiber.StatusConflict, "Vehicle does not have this kind of connection.")
	}

	vid, ok := ud.TokenID.Int64()
	if !ok {
		return fmt.Errorf("vehicle token id invalid, this should never happen %d", ud.TokenID)
	}

	var response *signer.TypedData

	if sdc.Settings.ConnectionsReplacedIntegrations {
		fmt.Println("XDD1")
		response = sdc.getEIP712MintV2(newIntegIDs.ConnectionID, vid)
	} else {
		fmt.Println("XDD2")

		response = sdc.getEIP712Mint(newIntegIDs.IntegrationNode.Int64(), vid)
	}

	return c.JSON(response)
}

// MintSyntheticDevice godoc
// @Description Submit a metadata
// @Tags        integrations
// @Produce     json
// @Param       userDeviceID path int true "user device KSUID, must be minted"
// @Param       integrationID path int true "integration KSUD, must be software-based"
// @Param       signed body controllers.MintSyntheticDeviceRequest true "only field is the signed EIP-712"
// @Success     204
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/mint [post]
func (sdc *SyntheticDevicesController) MintSyntheticDevice(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	newIntegIDs, ok := syntheticIntegrationKSUIDToOtherIDs[integrationID]
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot mint this integration with devices-api.")
	}

	tx, err := sdc.DBS().Writer.BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleTokenSyntheticDevice, models.SyntheticDeviceRels.MintRequest)),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID)),
		qm.Load(models.UserDeviceRels.BurnRequest),
	).One(c.Context(), tx)
	if err != nil {
		return err
	}

	if ud.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not minted.")
	}

	if burn := ud.R.BurnRequest; burn != nil && burn.Status != models.MetaTransactionRequestStatusFailed {
		return fiber.NewError(fiber.StatusConflict, "Vehicle is being burned.")
	}

	if sd := ud.R.VehicleTokenSyntheticDevice; sd != nil {
		if !sd.TokenID.IsZero() {
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("Vehicle already paired with synthetic device %d.", sd.TokenID.Big))
		}
		if sd.R.MintRequest.Status != models.MetaTransactionRequestStatusFailed {
			return fiber.NewError(fiber.StatusConflict, "There is already a synthetic device mint in progress for this vehicle.")
		}
		_, err := ud.R.VehicleTokenSyntheticDevice.Delete(c.Context(), tx)
		if err != nil {
			return fmt.Errorf("failed to delete failed synthetic minting: %v", err)
		}
	}

	if len(ud.R.UserDeviceAPIIntegrations) == 0 {
		return fiber.NewError(fiber.StatusConflict, "Vehicle does not have this kind of connection.")
	}

	vid, ok := ud.TokenID.Int64()
	if !ok {
		return fmt.Errorf("vehicle token id invalid, this should never happen %d", ud.TokenID)
	}

	var req MintSyntheticDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request.")
	}

	userAddr, err := helpers.GetJWTEthAddr(c)
	if err != nil {
		return err
	}

	rawPayload := sdc.getEIP712Mint(newIntegIDs.IntegrationNode.Int64(), vid)

	tdHash, _, err := signer.TypedDataAndHash(*rawPayload)
	if err != nil {
		sdc.log.Err(err).Msg("Error occurred creating hash of payload")
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't verify signature.")
	}

	ownerSignature := common.FromHex(req.Signature)

	recAddr, origErr := helpers.Ecrecover(tdHash, ownerSignature)
	if origErr != nil || recAddr != userAddr {
		ethClient, err := ethclient.Dial(sdc.Settings.MainRPCURL)
		if err != nil {
			return err
		}

		sigCon, err := sig2.NewErc1271(userAddr, ethClient)
		if err != nil {
			return err
		}

		ret, err := sigCon.IsValidSignature(nil, common.BytesToHash(tdHash), ownerSignature)
		if err != nil {
			return err
		}

		if ret != erc1271magicValue {
			return fiber.NewError(fiber.StatusBadRequest, "Could not verify ERC-1271 signature.")
		}
	}

	childKeyNumber, err := sdc.generateNextChildKeyNumber(c.Context())
	if err != nil {
		sdc.log.Err(err).Msg("failed to generate sequence from database")
		return fiber.NewError(fiber.StatusInternalServerError, "synthetic device minting request failed")
	}

	requestID := ksuid.New().String()

	syntheticDeviceAddr, err := sdc.walletSvc.GetAddress(c.Context(), uint32(childKeyNumber))
	if err != nil {
		sdc.log.Err(err).
			Str("function-name", "SyntheticWallet.GetAddress").
			Int("childKeyNumber", childKeyNumber).
			Msg("Error occurred getting synthetic wallet address")
		return err
	}

	virtSig, err := sdc.walletSvc.SignHash(c.Context(), uint32(childKeyNumber), tdHash)
	if err != nil {
		sdc.log.Err(err).
			Str("function-name", "SyntheticWallet.SignHash").
			Bytes("Hash", tdHash).
			Int("childKeyNumber", childKeyNumber).
			Msg("Error occurred signing message hash")
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
		IntegrationTokenID: types.NewDecimal(new(decimal.Big).SetBigMantScale(newIntegIDs.IntegrationNode, 0)),
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

	var realID *big.Int
	if sdc.Settings.ConnectionsReplacedIntegrations {
		realID = newIntegIDs.ConnectionID
	} else {
		realID = newIntegIDs.IntegrationNode
	}

	mvt := contracts.MintSyntheticDeviceInput{
		IntegrationNode:     realID,
		VehicleNode:         new(big.Int).SetInt64(vid),
		VehicleOwnerSig:     ownerSignature,
		SyntheticDeviceAddr: common.BytesToAddress(syntheticDeviceAddr),
		SyntheticDeviceSig:  virtSig,
	}

	if err := sdc.registryClient.MintSyntheticDeviceSign(requestID, mvt); err != nil {
		return err
	}

	// register synthetic device with tesla oracle
	if newIntegIDs.Name == "Tesla" {
		if _, err := sdc.teslaOracle.RegisterNewSyntheticDevice(c.Context(), &pb_oracle.RegisterNewSyntheticDeviceRequest{
			Vin:                    ud.VinIdentifier.String,
			SyntheticDeviceAddress: syntheticDeviceAddr,
			WalletChildNum:         uint64(childKeyNumber),
		}); err != nil {
			sdc.log.Err(err).Msg("failed to register synthetic device with tesla oracle")
		}
	}

	return c.JSON(fiber.Map{"message": "Submitted synthetic device mint request."})
}

func nameToConnectionID(name string) *big.Int {
	paddedBytes := make([]byte, 32)
	copy(paddedBytes, []byte(name))

	return new(big.Int).SetBytes(paddedBytes)
}

type newIDs struct {
	IntegrationNode *big.Int
	ConnectionID    *big.Int
	Name            string // LOL
}

var (
	syntheticIntegrationKSUIDToOtherIDs = map[string]*newIDs{
		"22N2xaPOq2WW2gAHBHd0Ikn4Zob": {
			IntegrationNode: big.NewInt(1),
			ConnectionID:    nameToConnectionID("Smartcar"),
			Name:            "Smartcar",
		},
		"26A5Dk3vvvQutjSyF0Jka2DP5lg": {
			IntegrationNode: big.NewInt(2),
			ConnectionID:    nameToConnectionID("Tesla"),
			Name:            "Tesla",
		},
	}
)

// GetSyntheticDeviceBurnPayload godoc
// @Description Produces the payload that the user signs and submits to burn a synthetic device.
// @Produce     json
// @Param       userDeviceID path int true "user device KSUID, must be minted"
// @Param       integrationID path int true "integration KSUD, must be software-based and active"
// @Success     200 {array} signer.TypedData
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/burn [get]
func (sdc *SyntheticDevicesController) GetSyntheticDeviceBurnPayload(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleTokenSyntheticDevice, models.SyntheticDeviceRels.MintRequest)),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleTokenSyntheticDevice, models.SyntheticDeviceRels.BurnRequest)),
	).One(c.Context(), sdc.DBS().Reader)
	if err != nil {
		return err
	}

	sd := ud.R.VehicleTokenSyntheticDevice

	if sd == nil {
		return fiber.NewError(fiber.StatusBadRequest, "No synthetic device associated with this vehicle.")
	}

	// Check that the integration id in the path matches the synthetic's integration.
	in, err := sdc.deviceDefSvc.GetIntegrationByID(c.Context(), integrationID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "failed to get integration")
	}

	if intNode, _ := sd.IntegrationTokenID.Uint64(); intNode != in.TokenId {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Associated synthetic device is not under integration %s.", integrationID))
	}

	// Check if minting is in progress.
	if sd.TokenID.IsZero() {
		if sd.R.MintRequest.Status == models.MetaTransactionRequestStatusFailed {
			return fiber.NewError(fiber.StatusConflict, "Synthetic device previously failed to mint; there is nothing to burn.")
		}
		return fiber.NewError(fiber.StatusConflict, "Synthetic device is currently minting; wait for this to complete.")
	}

	if br := sd.R.BurnRequest; br != nil && br.Status != models.MetaTransactionRequestStatusFailed {
		return fiber.NewError(fiber.StatusConflict, "Burning already in progress.")
	}

	vehicleNode, _ := ud.TokenID.Int64()
	syntheticDeviceNode, _ := sd.TokenID.Int64()

	return c.JSON(sdc.getEIP712Burn(vehicleNode, syntheticDeviceNode))
}

type BurnSyntheticDeviceRequest struct {
	Signature string `json:"signature"`
}

// BurnSyntheticDevice godoc
// @Description Submit the signature required for the synthetic device burning meta-transaction.
// @Produce     json
// @Param       userDeviceID path int true "user device KSUID, must be minted"
// @Param       integrationID path int true "integration KSUD, must be software-based and active"
// @Param       signed body controllers.BurnSyntheticDeviceRequest true "only field is the signed EIP-712"
// @Success     200 {array} signer.TypedData
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/burn [post]
func (sdc *SyntheticDevicesController) BurnSyntheticDevice(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	tx, err := sdc.DBS().Writer.BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleTokenSyntheticDevice, models.SyntheticDeviceRels.MintRequest)),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleTokenSyntheticDevice, models.SyntheticDeviceRels.BurnRequest)),
	).One(c.Context(), sdc.DBS().Reader)
	if err != nil {
		return err
	}

	sd := ud.R.VehicleTokenSyntheticDevice

	if sd == nil {
		return fiber.NewError(fiber.StatusBadRequest, "No synthetic device associated with this vehicle.")
	}

	// Check that the integration id in the path matches the synthetic's integration.
	in, err := sdc.deviceDefSvc.GetIntegrationByID(c.Context(), integrationID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "failed to get integration")
	}

	if intNode, _ := sd.IntegrationTokenID.Uint64(); intNode != in.TokenId {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Associated synthetic device is not under integration %s.", integrationID))
	}

	// Check if minting is in progress.
	if sd.TokenID.IsZero() {
		if sd.R.MintRequest.Status == models.MetaTransactionRequestStatusFailed {
			return fiber.NewError(fiber.StatusConflict, "Synthetic device previously failed to mint; there is nothing to burn.")
		}
		return fiber.NewError(fiber.StatusConflict, "Synthetic device is currently minting; wait for this to complete.")
	}

	if br := sd.R.BurnRequest; br != nil && br.Status != models.MetaTransactionRequestStatusFailed {
		return fiber.NewError(fiber.StatusConflict, "Burning already in progress.")
	}

	ownerAddr := common.BytesToAddress(ud.OwnerAddress.Bytes)

	vehicleNode, _ := ud.TokenID.Int64()
	syntheticDeviceNode, _ := sd.TokenID.Int64()

	td := sdc.getEIP712Burn(vehicleNode, syntheticDeviceNode)

	var req BurnSyntheticDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	ownerSignature := common.FromHex(req.Signature)

	hash, _, err := signer.TypedDataAndHash(*td)
	if err != nil {
		sdc.log.Err(err).Msg("Error occurred creating has of payload")
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't verify signature.")
	}

	recAddr, origErr := helpers.Ecrecover(hash, ownerSignature)
	if origErr != nil || recAddr != ownerAddr {
		ethClient, err := ethclient.Dial(sdc.Settings.MainRPCURL)
		if err != nil {
			return err
		}

		sigCon, err := sig2.NewErc1271(ownerAddr, ethClient)
		if err != nil {
			return err
		}

		ret, err := sigCon.IsValidSignature(nil, common.BytesToHash(hash), ownerSignature)
		if err != nil {
			return err
		}

		if ret != erc1271magicValue {
			return fiber.NewError(fiber.StatusBadRequest, "Could not verify ERC-1271 signature.")
		}
	}

	reqID := ksuid.New().String()

	mtr := models.MetaTransactionRequest{
		ID:     reqID,
		Status: models.MetaTransactionRequestStatusUnsubmitted,
	}

	if err := mtr.Insert(c.Context(), tx, boil.Infer()); err != nil {
		return err
	}

	sd.BurnRequestID = null.StringFrom(reqID)
	_, err = sd.Update(c.Context(), tx, boil.Whitelist(models.SyntheticDeviceColumns.BurnRequestID))
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return sdc.registryClient.BurnSyntheticDeviceSign(reqID, big.NewInt(vehicleNode), big.NewInt(syntheticDeviceNode), ownerSignature)
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
