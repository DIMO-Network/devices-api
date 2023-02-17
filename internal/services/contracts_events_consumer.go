package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type ContractsEventsConsumer struct {
	db       db.Store
	log      *zerolog.Logger
	settings *config.Settings
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
	Contract        string         `json:"contract,omitempty"`
	TransactionHash string         `json:"transactionHash,omitempty"`
	Arguments       map[string]any `json:"arguments,omitempty"`
	BlockCompleted  bool           `json:"blockCompleted,omitempty"`
	EventSignature  string         `json:"eventSignature,omitempty"`
	EventName       string         `json:"eventName,omitempty"`
}

func NewContractsEventsConsumer(pdb db.Store, log *zerolog.Logger, settings *config.Settings) *ContractsEventsConsumer {
	return &ContractsEventsConsumer{db: pdb, log: log, settings: settings}
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

	event := new(shared.CloudEvent[map[string]any])
	if err := json.Unmarshal(msg.Payload, event); err != nil {
		return errors.Wrap(err, "error parsing device event payload")
	}

	return c.processEvent(event)
}

func (c *ContractsEventsConsumer) processEvent(event *shared.CloudEvent[map[string]any]) error {
	if event.Type != contractEventCEType {
		return nil
	}

	var data ContractEventData

	err := mapstructure.Decode(event.Data, &data)
	if err != nil {
		return err
	}

	switch data.EventName {
	case PrivilegeSet.String():
		c.log.Info().Str("event", data.EventName).Msg("Event received")
		return c.setPrivilegeHandler(&data)
	case AftermarketDeviceNodeMinted.String():
		c.log.Info().Str("event", data.EventName).Msg("Event received")
		return c.setMintedAfterMarketDevice(&data)
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
	p := PrivilegeArgs{}
	err := mapstructure.WeakDecode(e.Arguments, &p)
	if err != nil {
		return err
	}

	t, err := strconv.ParseInt(p.ExpiresAt, 10, 64)
	if err != nil {
		panic(err) // TODO(elffjs): Get rid of this.
	}
	tm := time.Unix(t, 0)

	ti, ok := new(decimal.Big).SetString(p.TokenID)
	if !ok {
		c.log.Error().Msg(fmt.Sprintf("Couldn't parse token id %q.", ti))
		return fmt.Errorf("couldn't parse token id %q", p.TokenID)
	}

	tid := types.NewDecimal(ti)

	udp := models.NFTPrivilege{
		UserAddress:     common.FromHex(p.UserAddress),
		ContractAddress: common.FromHex(e.Contract),
		TokenID:         tid,
		Privilege:       p.PrivilegeID,
		Expiry:          tm,
	}

	nftCols := models.NFTPrivilegeColumns

	err = udp.Upsert(context.Background(), c.db.DBS().Writer, true, []string{nftCols.ContractAddress, nftCols.TokenID, nftCols.Privilege, nftCols.UserAddress}, boil.Infer(), boil.Infer())
	if err != nil {
		c.log.Error().Err(err).Msg("Failed to insert privilege record.")
		return err
	}

	return nil
}

type MintedAftermarketDeviceArgs struct {
	TokenID       string `mapstructure:"newTokenId"`
	DeviceAddress string
	Owner         string // TODO - confirm who the minter is
}

func (c *ContractsEventsConsumer) setMintedAfterMarketDevice(e *ContractEventData) error {
	p := MintedAftermarketDeviceArgs{}
	err := mapstructure.WeakDecode(e.Arguments, &p)
	if err != nil {
		return err
	}

	autopiApiService := NewAutoPiAPIService(c.settings, c.db.DBS)
	device, err := autopiApiService.GetDeviceByEthAddress(p.DeviceAddress)
	if err != nil {
		c.log.Error().Msg(fmt.Sprintf("Couldn't fetch dongle with eth_address %s.", p.DeviceAddress))
		return fmt.Errorf("couldn't fetch dongle with eth_address %s", p.DeviceAddress)
	}

	var maybeAddr null.Bytes

	if strAddr := device.EthereumAddress; common.IsHexAddress(p.DeviceAddress) {
		maybeAddr = null.BytesFrom(common.FromHex(strAddr))
	} else {
		c.log.Warn().Str("address", device.EthereumAddress).Msg("Invalid device Ethereum address from AutoPi.")
	}

	tokenID, _ := new(big.Int).SetString(p.TokenID, 16)

	ap := models.AutopiUnit{
		AutopiUnitID:    device.UnitID,
		AutopiDeviceID:  null.StringFrom(device.ID),
		EthereumAddress: maybeAddr,
		TokenID:         types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0)),
	}

	apCols := models.AutopiUnitColumns

	err = ap.Upsert(context.Background(), c.db.DBS().Writer, true, []string{apCols.AutopiUnitID, apCols.EthereumAddress}, boil.Infer(), boil.Infer())
	if err != nil {
		c.log.Error().Err(err).Msg("Failed to insert privilege record.")
		return err
	}

	return nil
}
