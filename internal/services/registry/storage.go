package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/DIMO-Network/shared"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/dbtypes"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type StatusProcessor interface {
	Handle(ctx context.Context, data *ceData) error
}

type proc struct {
	ABI      *abi.ABI
	DB       func() *db.ReaderWriter
	Logger   *zerolog.Logger
	settings *config.Settings
	Eventer  services.EventService
}

func (p *proc) Handle(ctx context.Context, data *ceData) error {
	logger := p.Logger.With().
		Str("requestId", data.RequestID).
		Str("status", data.Type).
		Str("hash", data.Transaction.Hash).
		Logger()

	logger.Info().Msg("Got transaction status.")

	tx, err := p.DB().Writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback() //nolint

	mtr, err := models.MetaTransactionRequests(
		models.MetaTransactionRequestWhere.ID.EQ(data.RequestID),
		// This is really ugly. We should probably link back to the type instead of doing this.
		qm.Load(models.MetaTransactionRequestRels.MintRequestUserDevice),
		qm.Load(models.MetaTransactionRequestRels.MintRequestSyntheticDevice),
	).One(context.Background(), tx)
	if err != nil {
		return err
	}

	mtr.Status = data.Type

	if data.Type != models.MetaTransactionRequestStatusFailed {
		mtr.Hash = null.BytesFrom(common.FromHex(data.Transaction.Hash))
	}

	_, err = mtr.Update(ctx, p.DB().Writer, boil.Infer())
	if err != nil {
		return err
	}

	if mtr.Status != models.MetaTransactionRequestStatusConfirmed {
		return tx.Commit()
	}

	vehicleNodeMintedWithDeviceDefinition := p.ABI.Events["VehicleNodeMintedWithDeviceDefinition"]
	syntheticDeviceMintedEvent := p.ABI.Events["SyntheticDeviceNodeMinted"]

	switch {
	case mtr.R.MintRequestUserDevice != nil:
		for _, logs := range data.Transaction.Logs {
			if logs.Topics[0] == vehicleNodeMintedWithDeviceDefinition.ID {
				var event contracts.RegistryVehicleNodeMintedWithDeviceDefinition
				err := p.parseLog(&event, vehicleNodeMintedWithDeviceDefinition, logs)
				if err != nil {
					return fmt.Errorf("failed to parse VehicleNodeMintedWithDeviceDefinition event: %w", err)
				}

				ud := mtr.R.MintRequestUserDevice
				cols := models.UserDeviceColumns

				ud.TokenID = dbtypes.NullIntToDecimal(event.VehicleId)
				ud.OwnerAddress = null.BytesFrom(event.Owner.Bytes())
				_, err = ud.Update(ctx, tx, boil.Whitelist(cols.TokenID, cols.OwnerAddress))
				if err != nil {
					return fmt.Errorf("failed to update vehicle record: %w", err)
				}

				p.Eventer.Emit(&shared.CloudEvent[any]{ //nolint
					Type:    "com.dimo.zone.device.mint",
					Subject: ud.ID,
					Source:  "devices-api",
					Data: services.UserDeviceMintEvent{
						Timestamp: time.Now(),
						UserID:    ud.UserID,
						Device: services.UserDeviceEventDevice{
							ID:                 ud.ID,
							VIN:                ud.VinIdentifier.String,
							DeviceDefinitionID: ud.DeviceDefinitionID,
						},
						NFT: services.UserDeviceEventNFT{
							TokenID: event.VehicleId,
							Owner:   event.Owner,
							TxHash:  common.HexToHash(data.Transaction.Hash),
						},
					},
				})

				logger.Info().
					Str("userDeviceId", mtr.R.MintRequestUserDevice.ID).
					Int64("vehicleTokenId", event.VehicleId.Int64()).
					Str("owner", event.Owner.Hex()).
					Msg("Vehicle minted.")
			} else if logs.Topics[0] == syntheticDeviceMintedEvent.ID {
				// We must be doing a combined vehicle and SD mint. This always comes second.
				var event contracts.RegistrySyntheticDeviceNodeMinted
				err := p.parseLog(&event, syntheticDeviceMintedEvent, logs)
				if err != nil {
					return fmt.Errorf("failed to parse SyntheticDeviceNodeMinted event: %w", err)
				}

				cols := models.SyntheticDeviceColumns
				sd := mtr.R.MintRequestSyntheticDevice
				if sd == nil {
					return fmt.Errorf("vehicle mint has a SyntheticDeviceNodeMinted log but there's no synthetic device attached to the meta-transaction")
				}

				sd.VehicleTokenID = mtr.R.MintRequestUserDevice.TokenID
				sd.TokenID = dbtypes.NullIntToDecimal(event.SyntheticDeviceNode)
				_, err = sd.Update(ctx, tx, boil.Whitelist(cols.VehicleTokenID, cols.TokenID))
				if err != nil {
					return fmt.Errorf("failed to update synthetic device record: %w", err)
				}

				logger.Info().
					Int64("vehicleTokenId", event.VehicleNode.Int64()).
					Int64("syntheticDeviceTokenId", event.SyntheticDeviceNode.Int64()).
					Str("owner", event.Owner.Hex()).
					Msg("Synthetic device minted.")
			}
		}
	// It's very important that this be after the case for VehicleNodeMinted.
	case mtr.R.MintRequestSyntheticDevice != nil:
		for _, log := range data.Transaction.Logs {
			if log.Topics[0] == syntheticDeviceMintedEvent.ID {
				var event contracts.RegistrySyntheticDeviceNodeMinted
				err := p.parseLog(&event, syntheticDeviceMintedEvent, log)
				if err != nil {
					return fmt.Errorf("failed to parse SyntheticDeviceNodeMinted event: %w", err)
				}

				sd := mtr.R.MintRequestSyntheticDevice
				sd.TokenID = dbtypes.NullIntToDecimal(event.SyntheticDeviceNode)
				if _, err := sd.Update(ctx, p.DB().Writer, boil.Whitelist(models.SyntheticDeviceColumns.TokenID)); err != nil {
					return fmt.Errorf("failed to update synthetic device record: %w", err)
				}

				logger.Info().
					Int64("vehicleTokenId", event.VehicleNode.Int64()).
					Int64("syntheticDeviceTokenId", event.SyntheticDeviceNode.Int64()).
					Str("owner", event.Owner.Hex()).
					Msg("Synthetic device minted.")
			}
		}
	}

	return tx.Commit()
}

func (p *proc) parseLog(out any, event abi.Event, log ceLog) error {
	if len(log.Data) > 0 {
		if err := p.ABI.UnpackIntoInterface(out, event.Name, log.Data); err != nil {
			return err
		}
	}

	var indexed abi.Arguments
	for _, arg := range event.Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}

	return abi.ParseTopics(out, indexed, log.Topics[1:])
}

func NewProcessor(
	db func() *db.ReaderWriter,
	logger *zerolog.Logger,
	settings *config.Settings,
	eventer services.EventService,
) (StatusProcessor, error) {
	regABI, err := contracts.RegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	return &proc{
		ABI:      regABI,
		DB:       db,
		Logger:   logger,
		settings: settings,
		Eventer:  eventer,
	}, nil
}
