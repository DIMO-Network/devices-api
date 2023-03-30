package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/contracts"

	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type ContractsEventsConsumer struct {
	db               db.Store
	log              *zerolog.Logger
	settings         *config.Settings
	registryAddr     common.Address
	autopiAPIService AutoPiAPIService
}

type EventName string

const (
	PrivilegeSet                EventName = "PrivilegeSet"
	AftermarketDeviceNodeMinted EventName = "AftermarketDeviceNodeMinted"
	Transfer                    EventName = "Transfer"
)

func (r EventName) String() string {
	return string(r)
}

const contractEventCEType = "zone.dimo.contract.event"

type ContractEventData struct {
	Contract        common.Address  `json:"contract"`
	TransactionHash common.Hash     `json:"transactionHash"`
	Arguments       json.RawMessage `json:"arguments"`
	EventSignature  common.Hash     `json:"eventSignature"`
	EventName       string          `json:"eventName"`
	// TODO(elffjs): chainID. Don't repeat this struct everywhere.
}

func NewContractsEventsConsumer(pdb db.Store, log *zerolog.Logger, settings *config.Settings) *ContractsEventsConsumer {
	autopiAPIService := NewAutoPiAPIService(settings, pdb.DBS)

	return &ContractsEventsConsumer{
		db:               pdb,
		log:              log,
		settings:         settings,
		registryAddr:     common.HexToAddress(settings.DIMORegistryAddr),
		autopiAPIService: autopiAPIService,
	}
}

func (c *ContractsEventsConsumer) ProcessContractsEventsMessages(messages <-chan *message.Message) {
	for msg := range messages {
		err := c.processMessage(msg)
		if err != nil {
			c.log.Err(err).Msg("error processing credential msg")
		}
	}
}

func (c *ContractsEventsConsumer) processMessage(msg *message.Message) error {
	// Keep the pipeline moving no matter what.
	defer func() { msg.Ack() }()

	// Deletion messages. We're the only actor that produces these, so ignore them.
	if msg.Payload == nil {
		return nil
	}

	event := new(shared.CloudEvent[json.RawMessage])
	if err := json.Unmarshal(msg.Payload, event); err != nil {
		return errors.Wrap(err, "error parsing device event payload")
	}

	return c.processEvent(event)
}

func (c *ContractsEventsConsumer) processEvent(event *shared.CloudEvent[json.RawMessage]) error {
	if event.Type != contractEventCEType {
		return nil
	}

	var data ContractEventData

	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	switch data.EventName {
	case PrivilegeSet.String():
		c.log.Info().Str("event", data.EventName).Msg("Event received")
		return c.setPrivilegeHandler(&data)
	case Transfer.String():
		if event.Source == fmt.Sprintf("chain/%s", c.settings.PolygonChainID) {
			c.log.Info().Str("event", data.EventName).Msg("Event received")
			return c.routeTransferEvent(&data)
		}
		c.log.Debug().Str("event", data.EventName).Interface("event data", event).Msg("Handler not provided for event.")
		return nil
	case AftermarketDeviceNodeMinted.String():
		if data.Contract == c.registryAddr {
			c.log.Info().Str("event", data.EventName).Msg("Event received")
			return c.setMintedAfterMarketDevice(&data)
		}
		fallthrough // TODO(elffjs): Danger!
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
	case common.HexToAddress(c.settings.AfterMarketContractAddress):
		return c.handleAfterMarketTransferEvent(e)
	default:
		c.log.Debug().Str("event", e.EventName).Interface("full event data", e).Msg("Handler not provided for event.")
	}

	return nil
}

func (c *ContractsEventsConsumer) handleAfterMarketTransferEvent(e *ContractEventData) error {
	ctx := context.Background()
	var args contracts.ContractsTransfer
	err := json.Unmarshal(e.Arguments, &args)
	if err != nil {
		return err
	}

	if !IsZeroAddress(args.From) { // This is not a mint
		tkID := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(1), 0))

		apUnit, err := models.AutopiUnits(models.AutopiUnitWhere.TokenID.EQ(tkID)).One(context.Background(), c.db.DBS().Reader)
		if err != nil || !apUnit.OwnerAddress.Valid {
			if errors.Is(err, sql.ErrNoRows) {
				c.log.Err(err).Str("tokenID", tkID.String()).Msg("Could not find device")
				return nil
			}
			c.log.Err(err).Str("tokenID", tkID.String()).Msg("Error occurred transferring device")
			return nil
		}

		apUnit.UserID = null.String{}

		cols := models.AutopiUnitColumns

		_, err = apUnit.Update(ctx, c.db.DBS().Writer, boil.Whitelist(cols.UserID))
		if err != nil {
			c.log.Err(err).Str("tokenID", tkID.String()).Msg("Error occurred transferring device")
			return nil
		}

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
		TokenID:         types.NewDecimal(new(decimal.Big).SetBigMantScale(args.TokenId, 0)),
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

	device, err := c.autopiAPIService.GetDeviceByEthAddress(args.AftermarketDeviceAddress.Hex())
	if err != nil {
		return fmt.Errorf("couldn't fetch dongle with address %s: %w", args.AftermarketDeviceAddress, err)
	}

	c.log.Info().Str("serial", device.UnitID).Msgf("Aftermarket device minted with address %s, token id %d.", args.AftermarketDeviceAddress, args.TokenId)

	ap := models.AutopiUnit{
		AutopiUnitID:    device.UnitID,
		AutopiDeviceID:  null.StringFrom(device.ID),
		EthereumAddress: null.BytesFrom(args.AftermarketDeviceAddress.Bytes()),
		TokenID:         types.NewNullDecimal(new(decimal.Big).SetBigMantScale(args.TokenId, 0)),
	}

	cols := models.AutopiUnitColumns

	err = ap.Upsert(context.Background(), c.db.DBS().Writer, true, []string{cols.AutopiUnitID}, boil.Whitelist(cols.AutopiDeviceID, cols.EthereumAddress, cols.TokenID, cols.UpdatedAt), boil.Infer())
	if err != nil {
		c.log.Error().Err(err).Msg("Failed to insert privilege record.")
		return err
	}

	return nil
}
