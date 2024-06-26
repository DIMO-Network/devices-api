package registry

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/dbtypes"
	"github.com/DIMO-Network/shared/event/sdmint"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gopkg.in/yaml.v3"
)

//go:embed dimo_registry_error_translations.yaml
var dimoRegistryErrorTranslationsRaw []byte

type StatusProcessor interface {
	Handle(ctx context.Context, data *ceData) error
}

type proc struct {
	ABI             *abi.ABI
	DB              func() *db.ReaderWriter
	Logger          *zerolog.Logger
	settings        *config.Settings
	Eventer         services.EventService
	ErrorTranslator *ABIErrorTranslator
	smartcarTask    services.SmartcarTaskService
	teslaTask       services.TeslaTaskService
	ddSvc           services.DeviceDefinitionService
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

	if data.Type == models.MetaTransactionRequestStatusFailed {
		errData := common.FromHex(data.Reason.Data)
		if len(errData) != 0 {
			friendlyError, err := p.ErrorTranslator.Decode(errData)
			if err != nil {
				logger.Err(err).Msg("Error decoding revert data.")
			} else {
				mtr.FailureReason = null.StringFrom(friendlyError)
			}
		}
	} else {
		mtr.Hash = null.BytesFrom(common.FromHex(data.Transaction.Hash))
	}

	_, err = mtr.Update(ctx, tx, boil.Infer())
	if err != nil {
		return err
	}

	if mtr.Status != models.MetaTransactionRequestStatusConfirmed {
		return tx.Commit()
	}

	vehicleNodeMintedWithDeviceDefinition := p.ABI.Events["VehicleNodeMintedWithDeviceDefinition"]
	vehicleNodeMinted := p.ABI.Events["VehicleNodeMinted"]
	syntheticDeviceMintedEvent := p.ABI.Events["SyntheticDeviceNodeMinted"]

	if ud := mtr.R.MintRequestUserDevice; ud != nil {
		for _, logs := range data.Transaction.Logs {
			if logs.Topics[0] == vehicleNodeMintedWithDeviceDefinition.ID {
				var event contracts.RegistryVehicleNodeMintedWithDeviceDefinition
				err := p.parseLog(&event, vehicleNodeMintedWithDeviceDefinition, logs)
				if err != nil {
					return fmt.Errorf("failed to parse VehicleNodeMintedWithDeviceDefinition event: %w", err)
				}

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
			} else if logs.Topics[0] == vehicleNodeMinted.ID {
				var event contracts.RegistryVehicleNodeMinted
				err := p.parseLog(&event, vehicleNodeMinted, logs)
				if err != nil {
					return fmt.Errorf("failed to parse VehicleNodeMinted event: %w", err)
				}

				ud := mtr.R.MintRequestUserDevice
				cols := models.UserDeviceColumns

				ud.TokenID = dbtypes.NullIntToDecimal(event.TokenId)
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
							ID:  ud.ID,
							VIN: ud.VinIdentifier.String,
						},
						NFT: services.UserDeviceEventNFT{
							TokenID: event.TokenId,
							Owner:   event.Owner,
							TxHash:  common.HexToHash(data.Transaction.Hash),
						},
					},
				})

				logger.Info().
					Str("userDeviceId", mtr.R.MintRequestUserDevice.ID).
					Int64("vehicleTokenId", event.TokenId.Int64()).
					Str("owner", event.Owner.Hex()).
					Msg("Vehicle minted.")
			}
		}
	}

	if sd := mtr.R.MintRequestSyntheticDevice; sd != nil {
		for _, log := range data.Transaction.Logs {
			if log.Topics[0] == syntheticDeviceMintedEvent.ID {
				var event contracts.RegistrySyntheticDeviceNodeMinted
				err := p.parseLog(&event, syntheticDeviceMintedEvent, log)
				if err != nil {
					return fmt.Errorf("failed to parse SyntheticDeviceNodeMinted event: %w", err)
				}

				integ, err := p.ddSvc.GetIntegrationByTokenID(ctx, event.IntegrationNode.Uint64())
				if err != nil {
					return fmt.Errorf("couldn't retrieve integration %d: %w", event.IntegrationNode, err)
				}

				cols := models.SyntheticDeviceColumns

				sd.TokenID = dbtypes.NullIntToDecimal(event.SyntheticDeviceNode)
				sd.VehicleTokenID = dbtypes.NullIntToDecimal(event.VehicleNode)

				if _, err := sd.Update(ctx, tx, boil.Whitelist(cols.TokenID, cols.VehicleTokenID)); err != nil {
					return fmt.Errorf("failed to update synthetic device record: %w", err)
				}

				ud, err := models.UserDevices(
					models.UserDeviceWhere.TokenID.EQ(dbtypes.NullIntToDecimal(event.VehicleNode)),
					qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integ.Id)),
				).One(ctx, tx)
				if err != nil {
					return fmt.Errorf("couldn't retrieve vehicle %d: %w", event.VehicleNode, err)
				}

				if len(ud.R.UserDeviceAPIIntegrations) == 0 {
					return fmt.Errorf("vehicle %d does not have integration %d being minted", event.VehicleNode, event.IntegrationNode)
				}

				switch integ.Vendor {
				case constants.SmartCarVendor:
					err := p.smartcarTask.StartPoll(ud.R.UserDeviceAPIIntegrations[0], sd)
					if err != nil {
						return err
					}
				case constants.TeslaVendor:
					err := p.teslaTask.StartPoll(ud.R.UserDeviceAPIIntegrations[0], sd)
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf("unexpected integration vendor %s", integ.Vendor)
				}

				p.Eventer.Emit(&shared.CloudEvent[any]{ //nolint
					ID:          ksuid.New().String(),
					Source:      "devices-api",
					SpecVersion: "1.0",
					Subject:     ud.ID,
					Time:        time.Now(),
					Type:        sdmint.Type,
					Data: sdmint.Data{
						Integration: sdmint.Integration{
							TokenID:       int(event.IntegrationNode.Int64()),
							IntegrationID: integ.Id,
						},
						Vehicle: sdmint.Vehicle{
							TokenID:      int(event.VehicleNode.Int64()),
							UserDeviceID: ud.ID,
						},
						Device: sdmint.Device{
							TokenID:           int(event.SyntheticDeviceNode.Int64()),
							ExternalID:        ud.R.UserDeviceAPIIntegrations[0].ExternalID.String,
							Address:           common.BytesToAddress(sd.WalletAddress),
							WalletChildNumber: uint32(sd.WalletChildNumber),
						},
					},
				})

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
	smartcarTask services.SmartcarTaskService,
	teslaTask services.TeslaTaskService,
	ddSvc services.DeviceDefinitionService,
) (StatusProcessor, error) {
	regABI, err := contracts.RegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	var errorTranslationMap map[string]string
	err = yaml.Unmarshal(dimoRegistryErrorTranslationsRaw, &errorTranslationMap)
	if err != nil {
		return nil, fmt.Errorf("error parsing error translation file: %w", err)
	}

	errorTranslator, err := NewABIErrorTranslator(regABI, errorTranslationMap)
	if err != nil {
		return nil, fmt.Errorf("error constructing error translater: %w", err)
	}

	return &proc{
		ABI:             regABI,
		DB:              db,
		Logger:          logger,
		settings:        settings,
		Eventer:         eventer,
		ErrorTranslator: errorTranslator,
		smartcarTask:    smartcarTask,
		teslaTask:       teslaTask,
		ddSvc:           ddSvc,
	}, nil
}
