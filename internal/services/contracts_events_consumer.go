package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type ContractsEventsConsumer struct {
	db  db.Store
	log *zerolog.Logger
}

type EventName string

const (
	PrivilegeSet EventName = "PrivilegeSet"
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

func NewContractsEventsConsumer(pdb db.Store, log *zerolog.Logger) *ContractsEventsConsumer {
	return &ContractsEventsConsumer{db: pdb, log: log}
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
