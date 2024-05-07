package services

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/DIMO-Network/shared/kafka"
	"github.com/segmentio/ksuid"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/services/dex"
	"github.com/DIMO-Network/devices-api/internal/utils"
	"google.golang.org/protobuf/proto"

	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type Integration interface {
	Pair(ctx context.Context, autoPiTokenID, vehicleTokenID *big.Int) error
	Unpair(ctx context.Context, autoPiTokenID, vehicleTokenID *big.Int) error
}

type ContractsEventsConsumer struct {
	db           db.Store
	log          *zerolog.Logger
	settings     *config.Settings
	registryAddr common.Address
	apInt        Integration
	mcInt        Integration
	ddSvc        DeviceDefinitionService
}

type EventName string

const (
	PrivilegeSet                          EventName = "PrivilegeSet"
	AftermarketDeviceNodeMinted           EventName = "AftermarketDeviceNodeMinted"
	Transfer                              EventName = "Transfer"
	BeneficiarySet                        EventName = "BeneficiarySet"
	DCNNameChanged                        EventName = "NameChanged"
	DCNNewNode                            EventName = "NewNode"
	DCNNewExpiration                      EventName = "NewExpiration"
	AftermarketDeviceClaimed              EventName = "AftermarketDeviceClaimed"
	AftermarketDevicePaired               EventName = "AftermarketDevicePaired"
	AftermarketDeviceUnpaired             EventName = "AftermarketDeviceUnpaired"
	AftermarketDeviceAttributeSet         EventName = "AftermarketDeviceAttributeSet"
	AftermarketDeviceAddressReset         EventName = "AftermarketDeviceAddressReset"
	VehicleNodeBurned                     EventName = "VehicleNodeBurned"
	VehicleNodeMintedWithDeviceDefinition EventName = "VehicleNodeMintedWithDeviceDefinition"
)

func (r EventName) String() string {
	return string(r)
}

const contractEventCEType = "zone.dimo.contract.event"

type ContractEventData struct {
	ChainID         int64           `json:"chainId"`
	EventName       string          `json:"eventName"`
	Block           Block           `json:"block,omitempty"`
	Contract        common.Address  `json:"contract"`
	TransactionHash common.Hash     `json:"transactionHash"`
	EventSignature  common.Hash     `json:"eventSignature"`
	Arguments       json.RawMessage `json:"arguments"`
	// TODO(elffjs): chainID. Don't repeat this struct everywhere.
}

type Block struct {
	Number *big.Int    `json:"number,omitempty"`
	Hash   common.Hash `json:"hash,omitempty"`
	Time   time.Time   `json:"time,omitempty"`
}

func NewContractsEventsConsumer(pdb db.Store, log *zerolog.Logger, settings *config.Settings, apInt Integration, mcInt Integration, ddSvc DeviceDefinitionService) *ContractsEventsConsumer {
	return &ContractsEventsConsumer{
		db:           pdb,
		log:          log,
		settings:     settings,
		registryAddr: common.HexToAddress(settings.DIMORegistryAddr),
		apInt:        apInt,
		mcInt:        mcInt,
		ddSvc:        ddSvc,
	}
}

func (c *ContractsEventsConsumer) RunConsumer() error {
	ctx := context.Background()

	if err := kafka.Consume[*shared.CloudEvent[json.RawMessage]](ctx, kafka.Config{
		Brokers: strings.Split(c.settings.KafkaBrokers, ","),
		Topic:   c.settings.ContractsEventTopic,
		Group:   "user-devices",
	}, c.processEvent, c.log); err != nil {
		c.log.Error().Err(err).Msg("error starting contracts events consumer")
		return err
	}

	c.log.Info().Msg("Starting contracts event consumer.")

	return nil
}

func (c *ContractsEventsConsumer) processEvent(_ context.Context, event *shared.CloudEvent[json.RawMessage]) error {
	if event == nil || event.Type != contractEventCEType {
		return nil
	}

	var data ContractEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	if event.Source != fmt.Sprintf("chain/%d", c.settings.DIMORegistryChainID) {
		c.log.Debug().Str("event", data.EventName).Interface("event data", event).Msg("Handler not provided for event.")
		return nil
	}
	switch data.EventName {
	case PrivilegeSet.String():
		c.log.Info().Str("event", data.EventName).Msg("Event received")
		return c.setPrivilegeHandler(&data)
	case Transfer.String():
		return c.routeTransferEvent(&data)
	case AftermarketDeviceNodeMinted.String():
		if data.Contract == c.registryAddr {
			c.log.Info().Str("event", data.EventName).Msg("Event received")
			return c.setMintedAfterMarketDevice(&data)
		}
	case BeneficiarySet.String():
		if data.Contract == c.registryAddr {
			c.log.Info().Str("event", data.EventName).Msg("Event received")
			return c.beneficiarySet(&data)
		}
	case DCNNameChanged.String():
		c.log.Info().Str("event", data.EventName).Msg("Event received")
		return c.dcnNameChanged(&data)
	case DCNNewNode.String():
		c.log.Info().Str("event", data.EventName).Msg("Event received")
		return c.dcnNewNode(&data)
	case DCNNewExpiration.String():
		c.log.Info().Str("event", data.EventName).Msg("Event received")
		return c.dcnNewExpiration(&data)
	case AftermarketDeviceClaimed.String():
		return c.aftermarketDeviceClaimed(&data)
	case AftermarketDevicePaired.String():
		return c.aftermarketDevicePaired(&data)
	case AftermarketDeviceUnpaired.String():
		return c.aftermarketDeviceUnpaired(&data)
	case AftermarketDeviceAttributeSet.String():
		return c.aftermarketDeviceAttributeSet(&data)
	case AftermarketDeviceAddressReset.String():
		return c.aftermarketDeviceAddressReset(&data)
	case VehicleNodeBurned.String():
		return c.vehicleNodeBurned(&data)
	case VehicleNodeMintedWithDeviceDefinition.String():
		return c.vehicleNodeMintedWithDeviceDefinition(&data)
	default:
		c.log.Debug().Str("event", data.EventName).Msg("Handler not provided for event.")
	}

	return nil
}

type PrivilegeArgs struct {
	TokenID     string
	Version     int64
	PrivilegeID int64  `mapstructure:"privId"`
	UserAddress string `mapstructure:"user"`
	ExpiresAt   string `mapstructure:"expires"`
}

func (c *ContractsEventsConsumer) routeTransferEvent(e *ContractEventData) error {
	switch e.Contract {
	case common.HexToAddress(c.settings.AftermarketDeviceContractAddress):
		c.log.Info().Str("event", e.EventName).Msg("Event received")
		return c.handleAfterMarketTransferEvent(e)
	case common.HexToAddress(c.settings.VehicleNFTAddress):
		c.log.Info().Str("event", e.EventName).Msg("Event received")
		return c.handleVehicleTransfer(e)
	default:
		c.log.Debug().Str("event", e.EventName).Interface("fullEventData", e).Msg("Handler not provided for contract")
	}

	return nil
}

func (c *ContractsEventsConsumer) handleVehicleTransfer(e *ContractEventData) error {
	ctx := context.Background()
	var args contracts.MultiPrivilegeTransfer
	err := json.Unmarshal(e.Arguments, &args)
	if err != nil {
		return err
	}

	tkID := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(args.TokenId, 0))

	if IsZeroAddress(args.From) {
		c.log.Debug().Str("tokenID", tkID.String()).Msg("Ignoring mint event")
		return nil
	}

	tx, err := c.db.DBS().Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback() //nolint

	rowsAff, err := models.NFTPrivileges(
		models.NFTPrivilegeWhere.TokenID.EQ(types.Decimal(tkID)),
	).DeleteAll(ctx, tx)
	if err != nil {
		return err
	}

	c.log.Info().Str("tokenId", tkID.String()).Msgf("Cleared %d privileges upon vehicle transfer.", rowsAff)

	ud, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(tkID),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	if IsZeroAddress(args.To) {
		_, err = ud.Delete(ctx, tx)
		if err != nil {
			return err
		}
	} else {
		s := dex.IDTokenSubject{
			UserId: args.To.Hex(),
			ConnId: "web3",
		}
		b, err := proto.Marshal(&s)
		if err != nil {
			return err
		}

		ud.UserID = base64.RawURLEncoding.EncodeToString(b)
		ud.OwnerAddress = null.BytesFrom(args.To.Bytes())

		if _, err := ud.Update(ctx, tx, boil.Whitelist(models.UserDeviceColumns.OwnerAddress, models.UserDeviceColumns.UserID)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (c *ContractsEventsConsumer) handleAfterMarketTransferEvent(e *ContractEventData) error {
	ctx := context.Background()
	var args contracts.AftermarketDeviceIdTransfer
	err := json.Unmarshal(e.Arguments, &args)
	if err != nil {
		return err
	}

	tkID := utils.BigToDecimal(args.TokenId)

	if IsZeroAddress(args.From) {
		// Handled in setMintedAfterMarketDevice.
		c.log.Debug().Str("tokenID", tkID.String()).Msg("ignoring mint event")
		return nil
	}

	apUnit, err := models.AftermarketDevices(models.AftermarketDeviceWhere.TokenID.EQ(tkID)).One(context.Background(), c.db.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.log.Err(err).Str("tokenID", tkID.String()).Msg("record not found as this might be a newly minted device")
			return errors.New("record not found as this might be a newly minted device")
		}
		c.log.Err(err).Str("tokenID", tkID.String()).Msg("error occurred transferring device")
		return errors.New("error occurred transferring device")
	}

	if IsZeroAddress(args.To) {
		// Burn.
		c.log.Info().Msgf("Burning aftermarket device %d.", tkID)
		_, err := models.AutopiJobs(models.AutopiJobWhere.AutopiUnitID.EQ(null.StringFrom(apUnit.Serial))).DeleteAll(ctx, c.db.DBS().Writer)
		if err != nil {
			return fmt.Errorf("error deleting jobs associated with aftermarket device: %w", err)
		}

		_, err = apUnit.Delete(ctx, c.db.DBS().Writer)
		if err != nil {
			return fmt.Errorf("error deleting aftermarket device: %w", err)
		}

		return nil
	}

	if !apUnit.OwnerAddress.Valid {
		c.log.Debug().Str("tokenID", tkID.String()).Msg("device has not been claimed yet")
		return nil
	}

	apUnit.UserID = null.String{}
	apUnit.OwnerAddress = null.BytesFrom(args.To.Bytes())
	apUnit.Beneficiary = null.Bytes{}

	cols := models.AftermarketDeviceColumns

	if _, err = apUnit.Update(ctx, c.db.DBS().Writer, boil.Whitelist(cols.UserID, cols.OwnerAddress, cols.Beneficiary, cols.UpdatedAt)); err != nil {
		c.log.Err(err).Str("tokenID", tkID.String()).Msg("error occurred transferring device")
		return errors.New("error occurred transferring device")
	}

	return nil
}

func (c *ContractsEventsConsumer) setPrivilegeHandler(e *ContractEventData) error {
	var args contracts.MultiPrivilegeSetPrivilegeData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	udp := models.NFTPrivilege{
		UserAddress:     args.User.Bytes(),
		ContractAddress: e.Contract.Bytes(),
		TokenID:         utils.BigToDecimal(args.TokenId),
		Privilege:       args.PrivId.Int64(),
		Expiry:          time.Unix(args.Expires.Int64(), 0),
	}

	cols := models.NFTPrivilegeColumns

	return udp.Upsert(context.Background(), c.db.DBS().Writer, true, []string{cols.ContractAddress, cols.TokenID, cols.Privilege, cols.UserAddress}, boil.Whitelist(cols.Expiry, cols.UpdatedAt), boil.Infer())
}

func (c *ContractsEventsConsumer) setMintedAfterMarketDevice(e *ContractEventData) error {
	var args contracts.RegistryAftermarketDeviceNodeMinted
	err := json.Unmarshal(e.Arguments, &args)
	if err != nil {
		return err
	}

	// Place this in a holding table until we receive AftermarketDeviceAttributeSet with the serial.
	pad := models.PartialAftermarketDevice{
		EthereumAddress:     args.AftermarketDeviceAddress.Bytes(),
		TokenID:             utils.BigToDecimal(args.TokenId),
		ManufacturerTokenID: utils.BigToDecimal(args.ManufacturerId),
	}

	if err := pad.Upsert(context.TODO(), c.db.DBS().Writer, false, []string{models.PartialAftermarketDeviceColumns.TokenID}, boil.Infer(), boil.Infer()); err != nil {
		return err
	}

	c.log.Info().Str("address", args.AftermarketDeviceAddress.Hex()).Msgf("Aftermarket device %d minted under manufacturer %d. Waiting for serial.", args.TokenId, args.ManufacturerId)

	return nil
}

func (c *ContractsEventsConsumer) aftermarketDeviceClaimed(e *ContractEventData) error {
	if e.ChainID != c.settings.DIMORegistryChainID || e.Contract != common.HexToAddress(c.settings.DIMORegistryAddr) {
		return fmt.Errorf("aftermarket claim from unexpected source %d/%s", e.ChainID, e.Contract)
	}

	var args contracts.RegistryAftermarketDeviceClaimed
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	am, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.TokenID.EQ(utils.BigToDecimal(args.AftermarketDeviceNode)),
	).One(context.TODO(), c.db.DBS().Reader)
	if err != nil {
		return err
	}

	c.log.Info().Int64("aftermarketDeviceNode", args.AftermarketDeviceNode.Int64()).Str("owner", args.Owner.Hex()).Msg("Claiming aftermarket device.")

	am.OwnerAddress = null.BytesFrom(args.Owner.Bytes())
	_, err = am.Update(context.TODO(), c.db.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.OwnerAddress))

	return err
}

func (c *ContractsEventsConsumer) aftermarketDevicePaired(e *ContractEventData) error {
	if e.ChainID != c.settings.DIMORegistryChainID || e.Contract != common.HexToAddress(c.settings.DIMORegistryAddr) {
		return fmt.Errorf("aftermarket claim from unexpected source %d/%s", e.ChainID, e.Contract)
	}

	var args contracts.RegistryAftermarketDevicePaired
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	log := c.log.With().Int64("vehicleNode", args.VehicleNode.Int64()).Int64("aftermarketDeviceNode", args.AftermarketDeviceNode.Int64()).Logger()
	log.Info().Msg("Pairing aftermarket device and vehicle.")

	am, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.TokenID.EQ(utils.BigToDecimal(args.AftermarketDeviceNode)),
	).One(context.TODO(), c.db.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed to retrieve aftermarket device: %w", err)
	}

	dm, err := c.ddSvc.GetMakeByTokenID(context.TODO(), am.DeviceManufacturerTokenID.Int(nil))
	if err != nil {
		return fmt.Errorf("error retrieving manufacturer %d: %w", am.DeviceManufacturerTokenID, err)
	}

	am.VehicleTokenID = types.NewNullDecimal(utils.BigToDecimal(args.VehicleNode).Big)
	_, err = am.Update(context.TODO(), c.db.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.VehicleTokenID))
	if err != nil {
		return fmt.Errorf("failed to update aftermarket device: %w", err)
	}

	switch dm.Name {
	case constants.AutoPiVendor:
		err = c.apInt.Pair(context.TODO(), args.AftermarketDeviceNode, args.VehicleNode)
	case "Hashdog":
		err = c.mcInt.Pair(context.TODO(), args.AftermarketDeviceNode, args.VehicleNode)
	default:
		err = fmt.Errorf("unexpected aftermarket device manufacturer vendor %s", dm.Name)
	}

	log.Info().Msg("aftermarket device pairing completed")
	return err

}

// aftermarketDeviceAttributeSet handles the event of the same name from the registry contract.
// At present this is only used to grab the serial number for Macarons AND AutoPi's
func (c *ContractsEventsConsumer) aftermarketDeviceAttributeSet(e *ContractEventData) error {
	// TODO(elffjs): Stop repeating the next eight lines in every handler.
	if e.ChainID != c.settings.DIMORegistryChainID || e.Contract != common.HexToAddress(c.settings.DIMORegistryAddr) {
		return fmt.Errorf("aftermarket claim from unexpected source %d/%s", e.ChainID, e.Contract)
	}

	var args contracts.RegistryAftermarketDeviceAttributeSet
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	if args.Attribute != "Serial" {
		return nil
	}

	tx, err := c.db.DBS().Writer.BeginTx(context.TODO(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	pad, err := models.PartialAftermarketDevices(
		models.PartialAftermarketDeviceWhere.TokenID.EQ(utils.BigToDecimal(args.TokenId)),
	).One(context.TODO(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	ad := models.AftermarketDevice{
		Serial:                    args.Info,
		EthereumAddress:           pad.EthereumAddress,
		TokenID:                   pad.TokenID,
		DeviceManufacturerTokenID: pad.ManufacturerTokenID,
	}

	err = ad.Upsert(context.TODO(), tx, false, []string{models.AftermarketDeviceColumns.EthereumAddress}, boil.Infer(), boil.Infer())
	if err != nil {
		return err
	}

	c.log.Info().Str("address", common.BytesToAddress(ad.EthereumAddress).Hex()).Msgf("Aftermarket device serial set to %s.", args.Info)

	_, err = pad.Delete(context.TODO(), tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *ContractsEventsConsumer) aftermarketDeviceUnpaired(e *ContractEventData) error {
	if e.ChainID != c.settings.DIMORegistryChainID || e.Contract != common.HexToAddress(c.settings.DIMORegistryAddr) {
		return fmt.Errorf("aftermarket claim from unexpected source %d/%s", e.ChainID, e.Contract)
	}

	var args contracts.RegistryAftermarketDeviceUnpaired
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	c.log.Info().Int64("vehicleNode", args.VehicleNode.Int64()).Int64("aftermarketDeviceNode", args.AftermarketDeviceNode.Int64()).Msg("Unpairing aftermarket device and vehicle.")

	am, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.TokenID.EQ(utils.BigToDecimal(args.AftermarketDeviceNode)),
	).One(context.TODO(), c.db.DBS().Reader)
	if err != nil {
		return err
	}

	dm, err := c.ddSvc.GetMakeByTokenID(context.TODO(), am.DeviceManufacturerTokenID.Int(nil))
	if err != nil {
		return fmt.Errorf("error retrieving manufacturer %d: %w", am.DeviceManufacturerTokenID, err)
	}

	am.VehicleTokenID = types.NullDecimal{}
	am.PairRequestID = null.String{}

	if _, err := am.Update(context.TODO(), c.db.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.VehicleTokenID, models.AftermarketDeviceColumns.PairRequestID)); err != nil {
		return err
	}

	switch dm.Name {
	case constants.AutoPiVendor:
		err = c.apInt.Unpair(context.TODO(), args.AftermarketDeviceNode, args.VehicleNode)
	case "Hashdog":
		err = c.mcInt.Unpair(context.TODO(), args.AftermarketDeviceNode, args.VehicleNode)
	default:
		err = fmt.Errorf("unexpected aftermarket device manufacturer vendor %s", dm.Name)
	}

	return err
}

func (c *ContractsEventsConsumer) beneficiarySet(e *ContractEventData) error {
	var args contracts.RegistryBeneficiarySet
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	if args.IdProxyAddress != common.HexToAddress(c.settings.AftermarketDeviceContractAddress) {
		c.log.Warn().Msgf("Beneficiary set on an unexpected contract: %s.", args.IdProxyAddress)
		return nil
	}

	c.log.Info().Int64("nodeID", args.NodeId.Int64()).Msgf("Aftermarket beneficiary set: %s.", args.Beneficiary)

	device, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.TokenID.EQ(utils.BigToDecimal(args.NodeId)),
	).One(context.Background(), c.db.DBS().Reader)
	if err != nil {
		return err
	}

	cols := models.AftermarketDeviceColumns

	if IsZeroAddress(args.Beneficiary) {
		device.Beneficiary = null.Bytes{}
	} else {
		device.Beneficiary = null.BytesFrom(args.Beneficiary[:])
	}

	if _, err = device.Update(context.Background(), c.db.DBS().Writer, boil.Whitelist(cols.Beneficiary, cols.UpdatedAt)); err != nil {
		c.log.Error().Err(err).Msg("Failed to set beneficiary.")
		return err
	}

	return nil
}

// dcnNameChanged processes an event of type NameChanged. Upserts DCN record, setting the Name
func (c *ContractsEventsConsumer) dcnNameChanged(e *ContractEventData) error {
	var args DCNNameChangedContract
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}
	// see if it exists first
	dcn, err := models.DCNS(models.DCNWhere.NFTNodeID.EQ(args.Node[:])).One(context.Background(), c.db.DBS().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrap(err, "failed to query for existing dcn")
	}
	if dcn == nil {
		dcn = &models.DCN{
			NFTNodeID: args.Node[:],
		}
	}
	if len(args.Name) == 0 {
		j, _ := e.Arguments.MarshalJSON()
		c.log.Warn().Str("handler", "dcnNameChanged").Str("eventPayload", string(j)).Msg("DCN Name Change argument is empty")
	}
	dcn.Name = null.StringFrom(args.Name)

	err = dcn.Upsert(context.Background(), c.db.DBS().Writer, true, []string{models.DCNColumns.NFTNodeID},
		boil.Whitelist(models.DCNColumns.Name, models.DCNColumns.UpdatedAt), boil.Infer())
	if err != nil {
		return errors.Wrapf(err, "failed to upsert dcn with name: %s", args.Name)
	}

	return nil
}

// dcnNewNode processes an event of type NewNode. Upserts DCN record, setting the Owner Address and Block creation time
func (c *ContractsEventsConsumer) dcnNewNode(e *ContractEventData) error {
	var args contracts.DcnRegistryNewNode
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}
	//question: should this be an insert always?
	dcn, err := models.DCNS(models.DCNWhere.NFTNodeID.EQ(args.Node[:])).One(context.Background(), c.db.DBS().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrap(err, "failed to query for existing dcn")
	}
	if dcn == nil {
		dcn = &models.DCN{
			NFTNodeID: args.Node[:],
		}
	}
	dcn.OwnerAddress = null.BytesFrom(args.Owner.Bytes())
	dcn.NFTNodeBlockCreateTime = null.TimeFrom(e.Block.Time)

	err = dcn.Upsert(context.Background(), c.db.DBS().Writer, true, []string{models.DCNColumns.NFTNodeID},
		boil.Whitelist(models.DCNColumns.OwnerAddress, models.DCNColumns.NFTNodeBlockCreateTime, models.DCNColumns.UpdatedAt), boil.Infer())
	if err != nil {
		return errors.Wrapf(err, "failed to upsert dcn with node: %s", args.Node)
	}

	return nil
}

// dcnNewExpiration processes an event of type NewExpiration. Upserts DCN record, setting the Expiration
func (c *ContractsEventsConsumer) dcnNewExpiration(e *ContractEventData) error {
	var args contracts.DcnRegistryNewExpiration
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}
	dcn, err := models.DCNS(models.DCNWhere.NFTNodeID.EQ(args.Node[:])).One(context.Background(), c.db.DBS().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrap(err, "failed to query for existing dcn")
	}
	if dcn == nil {
		dcn = &models.DCN{
			NFTNodeID: args.Node[:],
		}
	}
	t := time.Unix(args.Expiration.Int64(), 0)
	dcn.Expiration = null.TimeFrom(t)

	err = dcn.Upsert(context.Background(), c.db.DBS().Writer, true, []string{models.DCNColumns.NFTNodeID},
		boil.Whitelist(models.DCNColumns.Expiration, models.DCNColumns.UpdatedAt), boil.Infer())
	if err != nil {
		return errors.Wrapf(err, "failed to upsert dcn with node: %s", args.Node)
	}

	return nil
}

// aftermarketDeviceAddressReset handles the event of the same name from the registry contract.
func (c *ContractsEventsConsumer) aftermarketDeviceAddressReset(e *ContractEventData) error {
	if e.ChainID != c.settings.DIMORegistryChainID || e.Contract != common.HexToAddress(c.settings.DIMORegistryAddr) {
		return fmt.Errorf("aftermarket device address reset from unexpected source %d/%s", e.ChainID, e.Contract)
	}

	var args contracts.RegistryAftermarketDeviceAddressReset
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	_, err := c.db.DBS().Writer.Exec(
		`UPDATE devices_api.aftermarket_devices 
		SET ethereum_address = decode($1, 'hex')
		WHERE token_id = $2;`,
		strings.TrimPrefix(args.AftermarketDeviceAddress.String(), "0x"),
		args.TokenId.Int64())
	return err
}

// vehicleNodeBurned handles the event of the same name from the registry contract.
func (c *ContractsEventsConsumer) vehicleNodeBurned(e *ContractEventData) error {
	ctx := context.Background()
	if e.ChainID != c.settings.DIMORegistryChainID || e.Contract != common.HexToAddress(c.settings.DIMORegistryAddr) {
		return fmt.Errorf("vehicle burn from unexpected source %d/%s", e.ChainID, e.Contract)
	}

	var args contracts.RegistryVehicleNodeBurned
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	c.log.Info().Int64("vehicleNode", args.VehicleNode.Int64()).Msg("burning vehicle node")
	tx, err := c.db.DBS().Reader.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	ud, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(
			types.NewNullDecimal(new(decimal.Big).SetBigMantScale(args.VehicleNode, 0))),
		qm.Load(models.UserDeviceRels.BurnRequest),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	if _, err := ud.Delete(ctx, tx); err != nil {
		return err
	}

	if mtr := ud.R.BurnRequest; mtr != nil {
		mtr.Hash = null.BytesFrom(e.TransactionHash.Bytes())
		mtr.Status = models.MetaTransactionRequestStatusConfirmed
		if _, err := mtr.Update(ctx, tx, boil.Infer()); err != nil {
			return err
		}
	}

	c.log.Info().Int64("vehicleNode", args.VehicleNode.Int64()).Msg("vehicle node burned")
	return tx.Commit()
}

func (c *ContractsEventsConsumer) vehicleNodeMintedWithDeviceDefinition(e *ContractEventData) error {
	if e.ChainID != c.settings.DIMORegistryChainID || e.Contract != common.HexToAddress(c.settings.DIMORegistryAddr) {
		return fmt.Errorf("vehicle mint from unexpected source %d/%s", e.ChainID, e.Contract)
	}

	ctx := context.Background()
	var args contracts.RegistryVehicleNodeMintedWithDeviceDefinition
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return fmt.Errorf("failed to unmarshal arguments from mint event: %w", err)
	}

	log := c.log.With().Int64("vehicleNode", args.VehicleId.Int64()).Int64("manufacturerId", args.ManufacturerId.Int64()).Str("deviceDefinitionId", args.DeviceDefinitionId).Logger()
	log.Info().Msg("Minting vehicle with device definition")

	tx, err := c.db.DBS().Writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback() //nolint

	// TODO (AE): how to avoid a potential race here?
	if check, err := models.MetaTransactionRequests(
		models.MetaTransactionRequestWhere.Hash.EQ(null.BytesFrom(e.TransactionHash.Bytes())),
	).Exists(ctx, tx); err != nil {
		return fmt.Errorf("failed to check if vehicle mint event has already been handled: %w", err)
	} else if check {
		return fmt.Errorf("vehicle mint event has already been handled. tx hash: %s", e.TransactionHash.String())
	}

	userIDArgs := dex.IDTokenSubject{
		UserId: args.Owner.Hex(),
		ConnId: "web3",
	}
	userID, err := proto.Marshal(&userIDArgs)
	if err != nil {
		return fmt.Errorf("failed to marshal user id: %w", err)
	}

	// TODO (AE): query tableland for device style id?
	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             base64.RawURLEncoding.EncodeToString(userID),
		DeviceDefinitionID: args.DeviceDefinitionId,
		OwnerAddress:       null.BytesFrom(args.Owner.Bytes()),
		TokenID:            types.NewNullDecimal(new(decimal.Big).SetBigMantScale(args.VehicleId, 0)),
	}

	if err := ud.Insert(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("failed to insert new user device: %w", err)
	}

	return tx.Commit()

}

// DCNNameChangedContract represents a NameChanged event raised by the FullAbi contract.
// Could not use abigen struct because it did not consider the underscore in the name property for serialization
type DCNNameChangedContract struct {
	Node [32]byte
	Name string `json:"name_"`
	//Raw  types.Log // Blockchain specific contextual infos
}
