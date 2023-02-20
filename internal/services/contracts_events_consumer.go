package services

import (
	"context"
	"encoding/json"
	"fmt"
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
	db           db.Store
	log          *zerolog.Logger
	settings     *config.Settings
	registryAddr common.Address
}

type EventName string

const (
	PrivilegeSet                EventName = "PrivilegeSet"
	AftermarketDeviceNodeMinted EventName = "AftermarketDeviceNodeMinted"
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
	return &ContractsEventsConsumer{db: pdb, log: log, settings: settings, registryAddr: common.HexToAddress(settings.DIMORegistryAddr)}
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

	// TODO(elffjs): Don't initialize this every time.
	autopiAPIService := NewAutoPiAPIService(c.settings, c.db.DBS)

	device, err := autopiAPIService.GetDeviceByEthAddress(args.AftermarketDeviceAddress.Hex())
	if err != nil {
		return fmt.Errorf("couldn't fetch dongle with address %s: %w", args.AftermarketDeviceAddress, err)
	}

	c.log.Info().Msgf("Device minted with unit id %s, address %s, token id %d.", device.UnitID, args.AftermarketDeviceAddress, args.TokenId)

	ap := models.AutopiUnit{
		AutopiUnitID:    device.UnitID,
		AutopiDeviceID:  null.StringFrom(device.ID),
		EthereumAddress: null.BytesFrom(args.AftermarketDeviceAddress.Bytes()),
		TokenID:         types.NewNullDecimal(new(decimal.Big).SetBigMantScale(args.TokenId, 0)),
	}

	cols := models.AutopiUnitColumns

	err = ap.Upsert(context.Background(), c.db.DBS().Writer, true, []string{cols.AutopiUnitID}, boil.Whitelist(cols.AutopiDeviceID, cols.EthereumAddress, cols.TokenID), boil.Infer())
	if err != nil {
		c.log.Error().Err(err).Msg("Failed to insert privilege record.")
		return err
	}

	return nil
}
