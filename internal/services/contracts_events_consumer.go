package services

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/services/dex"
	"google.golang.org/protobuf/proto"

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
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
	BeneficiarySet              EventName = "BeneficiarySet"
	DCNNameChanged              EventName = "NameChanged"
	DCNNewNode                  EventName = "NewNode"
	DCNNewExpiration            EventName = "NewExpiration"
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
			c.log.Err(err).Msg("error processing contract events messages")
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

	if event.Source != fmt.Sprintf("chain/%d", c.settings.DIMORegistryChainID) {
		c.log.Debug().Str("event", data.EventName).Interface("event data", event).Msg("Handler not provided for event.")
		return nil
	}
	switch data.EventName {
	case PrivilegeSet.String():
		c.log.Info().Str("event", data.EventName).Msg("Event received")
		return c.setPrivilegeHandler(&data)
	case Transfer.String():
		c.log.Info().Str("event", data.EventName).Msg("Event received")
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
		return c.handleAfterMarketTransferEvent(e)
	case common.HexToAddress(c.settings.VehicleNFTAddress):
		return c.handleVehicleTransfer(e)
	default:
		c.log.Debug().Str("event", e.EventName).Interface("fullEventData", e).Msg("Handler not provided for contract")
	}

	return errors.New("Handler not provided for contract")
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

	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(tkID),
		qm.Load(models.VehicleNFTRels.UserDevice),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	nft.OwnerAddress = null.BytesFrom(args.To.Bytes())
	if _, err := nft.Update(ctx, tx, boil.Whitelist(models.VehicleNFTColumns.OwnerAddress)); err != nil {
		return err
	}

	if ud := nft.R.UserDevice; ud != nil {
		s := dex.IDTokenSubject{
			UserId: args.To.Hex(),
			ConnId: "web3",
		}
		b, err := proto.Marshal(&s)
		if err != nil {
			return err
		}

		ud.UserID = base64.RawURLEncoding.EncodeToString(b)
		if _, err := ud.Update(ctx, tx, boil.Whitelist(models.UserDeviceColumns.UserID)); err != nil {
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

	tkID := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(args.TokenId, 0))

	if IsZeroAddress(args.From) {
		c.log.Debug().Str("tokenID", tkID.String()).Msg("ignoring mint event")
		return nil
	}

	apUnit, err := models.AutopiUnits(models.AutopiUnitWhere.TokenID.EQ(tkID)).One(context.Background(), c.db.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.log.Err(err).Str("tokenID", tkID.String()).Msg("record not found as this might be a newly minted device")
			return errors.New("record not found as this might be a newly minted device")
		}
		c.log.Err(err).Str("tokenID", tkID.String()).Msg("error occurred transferring device")
		return errors.New("error occurred transferring device")
	}

	if !apUnit.OwnerAddress.Valid {
		c.log.Debug().Str("tokenID", tkID.String()).Msg("device has not been claimed yet")
		return nil
	}

	apUnit.UserID = null.String{}
	apUnit.OwnerAddress = null.BytesFrom(args.To.Bytes())
	apUnit.Beneficiary = null.Bytes{}

	cols := models.AutopiUnitColumns

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

func (c *ContractsEventsConsumer) beneficiarySet(e *ContractEventData) error {
	var args contracts.RegistryBeneficiarySet
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	c.log.Info().Int64("nodeID", args.NodeId.Int64()).Msgf("Aftermarket beneficiary set: %s.", args.Beneficiary)

	device, err := models.AutopiUnits(
		models.AutopiUnitWhere.TokenID.EQ(types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(args.NodeId.Int64()), 0))),
	).One(context.Background(), c.db.DBS().Reader)
	if err != nil {
		return err
	}

	cols := models.AutopiUnitColumns

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
		c.log.Warn().Str("handler", "dcnNameChanged").Msg("DCN Name Change argument is empty: args.name")
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

// DCNNameChangedContract represents a NameChanged event raised by the FullAbi contract.
type DCNNameChangedContract struct {
	Node [32]byte
	Name string `json:"name_"`
	//Raw  types.Log // Blockchain specific contextual infos
}
