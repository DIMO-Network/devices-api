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

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/services/dex"
	"github.com/DIMO-Network/devices-api/internal/utils"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/dbtypes"
	"github.com/DIMO-Network/shared/kafka"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/protobuf/proto"
)

type Integration interface {
	Pair(ctx context.Context, autoPiTokenID, vehicleTokenID *big.Int) error
	Unpair(ctx context.Context, autoPiTokenID, vehicleTokenID *big.Int) error
}

//go:generate mockgen -source=./contracts_events_consumer.go -destination=./contract_events_consumer_mocks_test.go -package=services
type SyntheticTaskService interface {
	StopPoll(udai *models.UserDeviceAPIIntegration) error
}

type ContractsEventsConsumer struct {
	db           db.Store
	log          *zerolog.Logger
	settings     *config.Settings
	registryAddr common.Address
	genericInt   Integration
	ddSvc        DeviceDefinitionService
	evtSvc       EventService

	scTask    SyntheticTaskService
	teslaTask SyntheticTaskService
}

type EventName string

const (
	PrivilegeSet                          EventName = "PrivilegeSet"
	AftermarketDeviceNodeMinted           EventName = "AftermarketDeviceNodeMinted"
	Transfer                              EventName = "Transfer"
	BeneficiarySet                        EventName = "BeneficiarySet"
	AftermarketDeviceClaimed              EventName = "AftermarketDeviceClaimed"
	AftermarketDevicePaired               EventName = "AftermarketDevicePaired"
	AftermarketDeviceUnpaired             EventName = "AftermarketDeviceUnpaired"
	AftermarketDeviceAttributeSet         EventName = "AftermarketDeviceAttributeSet"
	AftermarketDeviceAddressReset         EventName = "AftermarketDeviceAddressReset"
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

func NewContractsEventsConsumer(pdb db.Store, log *zerolog.Logger, settings *config.Settings, genericInt Integration, ddSvc DeviceDefinitionService, evtSvc EventService, scTask SyntheticTaskService, teslaTask SyntheticTaskService) *ContractsEventsConsumer {
	return &ContractsEventsConsumer{
		db:           pdb,
		log:          log,
		settings:     settings,
		registryAddr: common.HexToAddress(settings.DIMORegistryAddr),
		genericInt:   genericInt,
		ddSvc:        ddSvc,
		evtSvc:       evtSvc,
		scTask:       scTask,
		teslaTask:    teslaTask,
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

func (c *ContractsEventsConsumer) processEvent(ctx context.Context, event *shared.CloudEvent[json.RawMessage]) error {
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
		return c.routeTransferEvent(ctx, &data)
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
	case VehicleNodeMintedWithDeviceDefinition.String():
		return c.vehicleNodeMintedWithDeviceDefinition(&data)
	default:
		c.log.Debug().Str("event", data.EventName).Msg("Handler not provided for event.")
	}

	return nil
}

func (c *ContractsEventsConsumer) routeTransferEvent(ctx context.Context, e *ContractEventData) error {
	switch e.Contract {
	case common.HexToAddress(c.settings.AftermarketDeviceContractAddress):
		c.log.Info().Str("event", e.EventName).Msg("Event received")
		return c.handleAfterMarketTransferEvent(e)
	case common.HexToAddress(c.settings.VehicleNFTAddress):
		c.log.Info().Str("event", e.EventName).Msg("Event received")
		return c.handleVehicleTransfer(ctx, e)
	case common.HexToAddress(c.settings.SyntheticDeviceNFTAddress):
		c.log.Info().Str("event", e.EventName).Msg("Event received")
		return c.handleSyntheticTransfer(ctx, e)
	default:
		c.log.Debug().Str("event", e.EventName).Interface("fullEventData", e).Msg("Handler not provided for contract")
	}

	return nil
}

func (c *ContractsEventsConsumer) handleSyntheticTransfer(ctx context.Context, e *ContractEventData) error {
	var args contracts.MultiPrivilegeTransfer
	err := json.Unmarshal(e.Arguments, &args)
	if err != nil {
		return err
	}

	// Only interested in burns. Mints are handled as meta-transcations and for all other
	// transfers, synthetics transfer together with the vehicle, so no need to track ownership.
	if !IsZeroAddress(args.To) {
		return nil
	}

	sd, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.TokenID.EQ(dbtypes.NullIntToDecimal(args.TokenId)),
		qm.Load(models.SyntheticDeviceRels.VehicleToken),
	).One(ctx, c.db.DBS().Writer)
	if err != nil {
		return fmt.Errorf("couldn't find synthetic device %d to burn: %w", args.TokenId, err)
	}

	// The most important thing is to delete the database rows to free things up.
	_, err = sd.Delete(ctx, c.db.DBS().Writer)
	if err != nil {
		return fmt.Errorf("failed to delete synthetic device %d row: %w", sd.TokenID, err)
	}

	ud := sd.R.VehicleToken
	if ud == nil {
		return fmt.Errorf("burning synthetic device %d with no paired vehicle", sd.TokenID)
	}

	intID, _ := sd.IntegrationTokenID.Uint64()
	integ, err := c.ddSvc.GetIntegrationByTokenID(ctx, intID)
	if err != nil {
		return err
	}

	udai, err := models.FindUserDeviceAPIIntegration(ctx, c.db.DBS().Reader.DB, ud.ID, integ.Id)
	if err != nil {
		return fmt.Errorf("failed to find job backing burned synthetic device %d: %w", sd.TokenID, err)
	}

	_, err = udai.Delete(ctx, c.db.DBS().Writer)
	if err != nil {
		return fmt.Errorf("failed to delete job backing synthetic device %d: %w", sd.TokenID, err)
	}

	if udai.TaskID.Valid {
		switch integ.Vendor {
		case constants.SmartCarVendor:
			err := c.scTask.StopPoll(udai)
			if err != nil {
				return err
			}
		case constants.TeslaVendor:
			err := c.teslaTask.StopPoll(udai)
			if err != nil {
				return err
			}
		default:
			c.log.Warn().Msgf("Unexpected integration %s.", integ.Vendor)
		}
	}

	// Need this for the event.
	dd, err := c.ddSvc.GetDeviceDefinitionBySlug(ctx, ud.DefinitionID)
	if err != nil {
		return err
	}

	err = c.evtSvc.Emit(&shared.CloudEvent[any]{
		Type:    "com.dimo.zone.device.integration.delete",
		Source:  "devices-api",
		Subject: ud.ID,
		Data: UserDeviceIntegrationEvent{
			Timestamp: time.Now(),
			UserID:    ud.UserID,
			Device: UserDeviceEventDevice{
				ID:           ud.ID,
				Make:         dd.Make.Name,
				Model:        dd.Model,
				Year:         int(dd.Year),
				VIN:          ud.VinIdentifier.String,
				DefinitionID: dd.Id,
			},
			Integration: UserDeviceEventIntegration{
				ID:     integ.Id,
				Type:   integ.Type,
				Style:  integ.Style,
				Vendor: integ.Vendor,
			},
		},
	})
	if err != nil {
		c.log.Info().Int64("syntheticDeviceTokenId", args.TokenId.Int64()).Str("owner", args.From.Hex()).Msg("Couldn't send out integration deletion event.")
	}

	c.log.Info().Int64("syntheticDeviceTokenId", args.TokenId.Int64()).Str("owner", args.From.Hex()).Msg("Burned synthetic device.")

	return nil
}

func (c *ContractsEventsConsumer) handleVehicleTransfer(ctx context.Context, e *ContractEventData) error {
	var args contracts.MultiPrivilegeTransfer
	err := json.Unmarshal(e.Arguments, &args)
	if err != nil {
		return err
	}

	// Handle mints in the meta-transaction handler.
	if IsZeroAddress(args.From) {
		return nil
	}

	tx, err := c.db.DBS().Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	rowsAff, err := models.NFTPrivileges(
		models.NFTPrivilegeWhere.TokenID.EQ(dbtypes.IntToDecimal(args.TokenId)),
	).DeleteAll(ctx, tx)
	if err != nil {
		return err
	}

	if rowsAff != 0 {
		c.log.Info().Int64("vehicleTokenId", args.TokenId.Int64()).Msgf("Cleared %d privileges upon vehicle transfer.", rowsAff)
	}

	ud, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(dbtypes.NullIntToDecimal(args.TokenId)),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	if IsZeroAddress(args.To) {
		_, err = ud.Delete(ctx, tx)
		if err != nil {
			return err
		}

		c.log.Info().Int64("vehicleTokenId", args.TokenId.Int64()).Str("owner", args.From.Hex()).Msg("Burned vehicle.")
	} else {
		// Faking a user id for a web3 user with the new owner address.
		userID, err := addressToUserID(args.To)
		if err != nil {
			return fmt.Errorf("failed to convert address to user id: %w", err)
		}

		cols := models.UserDeviceColumns
		ud.UserID = userID
		ud.OwnerAddress = null.BytesFrom(args.To.Bytes())

		if _, err := ud.Update(ctx, tx, boil.Whitelist(cols.UserID, cols.OwnerAddress)); err != nil {
			return err
		}

		c.log.Info().Int64("vehicleTokenId", args.TokenId.Int64()).Msgf("Transferred vehicle from %s to %s.", args.From, args.To)
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

	cols := models.AftermarketDeviceColumns

	am.VehicleTokenID = types.NewNullDecimal(utils.BigToDecimal(args.VehicleNode).Big)
	_, err = am.Update(context.TODO(), c.db.DBS().Writer, boil.Whitelist(cols.VehicleTokenID, cols.UpdatedAt))
	if err != nil {
		return fmt.Errorf("failed to update aftermarket device: %w", err)
	}

	return c.genericInt.Pair(context.TODO(), args.AftermarketDeviceNode, args.VehicleNode)
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

	am.VehicleTokenID = types.NullDecimal{}
	am.PairRequestID = null.String{}

	if _, err := am.Update(context.TODO(), c.db.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.VehicleTokenID, models.AftermarketDeviceColumns.PairRequestID)); err != nil {
		return err
	}

	return c.genericInt.Unpair(context.TODO(), args.AftermarketDeviceNode, args.VehicleNode)
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

	if check, err := models.MetaTransactionRequests(
		models.MetaTransactionRequestWhere.Hash.EQ(null.BytesFrom(e.TransactionHash.Bytes())),
	).Exists(ctx, tx); err != nil {
		// if we cannot confirm whether event has been handled, log error and then create vehicle
		log.Err(err).Msg("failed to check if vehicle mint event has already been handled")
	} else if check {
		log.Info().Msgf("vehicle mint event has already been handled. tx hash: %s", e.TransactionHash.String())
		return nil
	}

	userID, err := addressToUserID(args.Owner)
	if err != nil {
		return fmt.Errorf("failed to convert address to user id: %w", err)
	}

	ud := models.UserDevice{
		ID:           ksuid.New().String(),
		UserID:       userID,
		OwnerAddress: null.BytesFrom(args.Owner.Bytes()),
		TokenID:      dbtypes.NullIntToDecimal(args.VehicleId),
		DefinitionID: args.DeviceDefinitionId,
	}

	if err := ud.Insert(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("failed to insert new user device: %w", err)
	}

	c.evtSvc.Emit(&shared.CloudEvent[any]{ //nolint
		Type:    "com.dimo.zone.device.mint",
		Subject: ud.ID,
		Source:  "devices-api",
		Data: UserDeviceMintEvent{
			Timestamp: time.Now(),
			UserID:    ud.UserID,
			Device: UserDeviceEventDevice{
				ID:           ud.ID,
				VIN:          ud.VinIdentifier.String,
				DefinitionID: args.DeviceDefinitionId,
			},
			NFT: UserDeviceEventNFT{
				TokenID: args.VehicleId,
				Owner:   args.Owner,
				TxHash:  e.TransactionHash,
			},
		},
	})

	return tx.Commit()

}

func addressToUserID(addr common.Address) (string, error) {
	userIDArgs := dex.IDTokenSubject{
		UserId: addr.Hex(),
		ConnId: "web3",
	}

	ub, err := proto.Marshal(&userIDArgs)
	return base64.RawURLEncoding.EncodeToString(ub), err
}
