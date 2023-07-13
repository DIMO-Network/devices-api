package registry

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/autopi"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type StatusProcessor interface {
	Handle(ctx context.Context, data *ceData) error
}

type proc struct {
	ABI             *abi.ABI
	DB              func() *db.ReaderWriter
	Logger          *zerolog.Logger
	ap              *autopi.Integration
	settings        *config.Settings
	smartcarTaskSvc services.SmartcarTaskService
	teslaTaskSvc    services.TeslaTaskService
	deviceDefSvc    services.DeviceDefinitionService
}

func (p *proc) Handle(ctx context.Context, data *ceData) error {
	logger := p.Logger.With().
		Str("requestId", data.RequestID).
		Str("status", data.Type).
		Str("hash", data.Transaction.Hash).
		Logger()

	logger.Info().Msg("Got transaction status.")

	mtr, err := models.MetaTransactionRequests(
		models.MetaTransactionRequestWhere.ID.EQ(data.RequestID),
		// This is really ugly. We should probably link back to the type instead of doing this.
		qm.Load(models.MetaTransactionRequestRels.MintRequestVehicleNFT),
		qm.Load(models.MetaTransactionRequestRels.ClaimMetaTransactionRequestAftermarketDevice),
		qm.Load(models.MetaTransactionRequestRels.PairRequestAftermarketDevice),
		qm.Load(models.MetaTransactionRequestRels.UnpairRequestAftermarketDevice),
		qm.Load(qm.Rels(models.MetaTransactionRequestRels.MintRequestSyntheticDevice, models.SyntheticDeviceRels.VehicleToken)),
		qm.Load(models.MetaTransactionRequestRels.BurnRequestSyntheticDevice),
	).One(context.Background(), p.DB().Reader)
	if err != nil {
		return err
	}

	mtr.Status = data.Type
	mtr.Hash = null.BytesFrom(common.FromHex(data.Transaction.Hash))

	_, err = mtr.Update(ctx, p.DB().Writer, boil.Infer())
	if err != nil {
		return err
	}

	if mtr.Status != models.MetaTransactionRequestStatusConfirmed {
		return nil
	}

	vehicleMintedEvent := p.ABI.Events["VehicleNodeMinted"]
	deviceClaimedEvent := p.ABI.Events["AftermarketDeviceClaimed"]
	devicePairedEvent := p.ABI.Events["AftermarketDevicePaired"]
	deviceUnpairedEvent := p.ABI.Events["AftermarketDeviceUnpaired"]
	syntheticDeviceMintedEvent := p.ABI.Events["SyntheticDeviceNodeMinted"]
	sdBurnEvent := p.ABI.Events["SyntheticDeviceNodeBurned"]

	switch {
	case mtr.R.MintRequestVehicleNFT != nil:
		for _, l1 := range data.Transaction.Logs {
			if l1.Topics[0] == vehicleMintedEvent.ID {
				out := new(contracts.RegistryVehicleNodeMinted)
				err := p.parseLog(out, vehicleMintedEvent, l1)
				if err != nil {
					return err
				}

				mtr.R.MintRequestVehicleNFT.TokenID = types.NewNullDecimal(new(decimal.Big).SetBigMantScale(out.TokenId, 0))
				mtr.R.MintRequestVehicleNFT.OwnerAddress = null.BytesFrom(out.Owner.Bytes())
				_, err = mtr.R.MintRequestVehicleNFT.Update(ctx, p.DB().Writer, boil.Infer())
				if err != nil {
					return err
				}

				logger.Info().Str("userDeviceId", mtr.R.MintRequestVehicleNFT.UserDeviceID.String).Msg("Vehicle minted.")
			}
		}
		// Other soon.
	case mtr.R.ClaimMetaTransactionRequestAftermarketDevice != nil:
		for _, l1 := range data.Transaction.Logs {
			if l1.Topics[0] == deviceClaimedEvent.ID {
				out := new(contracts.RegistryAftermarketDeviceClaimed)
				err := p.parseLog(out, deviceClaimedEvent, l1)
				if err != nil {
					return err
				}

				mtr.R.ClaimMetaTransactionRequestAftermarketDevice.OwnerAddress = null.BytesFrom(out.Owner[:])
				_, err = mtr.R.ClaimMetaTransactionRequestAftermarketDevice.Update(ctx, p.DB().Writer, boil.Infer())
				if err != nil {
					return err
				}

				logger.Info().Str("autoPiTokenId", mtr.R.ClaimMetaTransactionRequestAftermarketDevice.TokenID.String()).Str("owner", out.Owner.String()).Msg("Device claimed.")
			}
		}
	case mtr.R.PairRequestAftermarketDevice != nil:
		for _, l1 := range data.Transaction.Logs {
			if l1.Topics[0] == devicePairedEvent.ID {
				out := new(contracts.RegistryAftermarketDevicePaired)
				err := p.parseLog(out, devicePairedEvent, l1)
				if err != nil {
					return err
				}

				mtr.R.PairRequestAftermarketDevice.VehicleTokenID = types.NewNullDecimal(new(decimal.Big).SetBigMantScale(out.VehicleNode, 0))
				_, err = mtr.R.PairRequestAftermarketDevice.Update(ctx, p.DB().Writer, boil.Infer())
				if err != nil {
					return err
				}

				return p.ap.Pair(ctx, out.AftermarketDeviceNode, out.VehicleNode)
			}
		}
	case mtr.R.UnpairRequestAftermarketDevice != nil:
		for _, l1 := range data.Transaction.Logs {
			if l1.Topics[0] == deviceUnpairedEvent.ID {
				out := new(contracts.RegistryAftermarketDeviceUnpaired)
				err := p.parseLog(out, deviceUnpairedEvent, l1)
				if err != nil {
					return err
				}

				mtr.R.UnpairRequestAftermarketDevice.VehicleTokenID = types.NullDecimal{}
				mtr.R.UnpairRequestAftermarketDevice.PairRequestID = null.String{}
				_, err = mtr.R.UnpairRequestAftermarketDevice.Update(ctx, p.DB().Writer, boil.Infer())
				if err != nil {
					return err
				}

				return p.ap.Unpair(ctx, out.AftermarketDeviceNode, out.VehicleNode)
			}
		}
	case mtr.R.MintRequestSyntheticDevice != nil:
		for _, l1 := range data.Transaction.Logs {
			if l1.Topics[0] == syntheticDeviceMintedEvent.ID {
				out := new(contracts.RegistrySyntheticDeviceNodeMinted)
				err := p.parseLog(out, syntheticDeviceMintedEvent, l1)

				if err != nil {
					return err
				}

				sd := mtr.R.MintRequestSyntheticDevice

				tkID := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(out.SyntheticDeviceNode, 0))
				sd.TokenID = tkID

				if _, err := sd.Update(ctx, p.DB().Writer, boil.Infer()); err != nil {
					return err
				}

				intToken, intTkErr := sd.IntegrationTokenID.Uint64()
				if !intTkErr {
					return errors.New("error occurred parsing integration tokenID")
				}
				integration, err := p.deviceDefSvc.GetIntegrationByTokenID(ctx, intToken)
				if err != nil {
					return err
				}

				ud, err := models.UserDeviceAPIIntegrations(
					models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(sd.R.VehicleToken.UserDeviceID.String),
					models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusPending),
					models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integration.Id),
					qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
				).One(ctx, p.DB().Reader)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						p.Logger.Debug().Err(err).Str("userDeviceID", sd.R.VehicleToken.UserDeviceID.String).Msg("Device has been deleted")
						return nil
					}
					return err
				}

				ud.Status = models.UserDeviceAPIIntegrationStatusPendingFirstData
				if _, err := ud.Update(ctx, p.DB().Writer, boil.Infer()); err != nil {
					return err
				}

				switch integration.Vendor {
				case constants.SmartCarVendor:
					if err := p.smartcarTaskSvc.StartPoll(ud); err != nil {
						logger.Err(err).Msg("Couldn't start Smartcar polling.")
						return err
					}
				case constants.TeslaVendor:
					extID, err := strconv.Atoi(ud.ExternalID.String)
					if err != nil {
						logger.Err(err).Msg("cannot convert tesla external id to int")
						return err
					}

					var metadata services.UserDeviceAPIIntegrationsMetadata
					err = json.Unmarshal(ud.Metadata.JSON, &metadata)
					if err != nil {
						logger.Err(err).Msg("unable to parse metadata")
						return err
					}

					v := &services.TeslaVehicle{
						ID:        extID,
						VIN:       ud.R.UserDevice.VinIdentifier.String,
						VehicleID: metadata.TeslaVehicleID,
					}

					if err := p.teslaTaskSvc.StartPoll(v, ud); err != nil {
						logger.Err(err).Msg("Couldn't start Tesla polling.")
						return err
					}
				}
			}
		}
	case mtr.R.BurnRequestSyntheticDevice != nil:
		for _, l1 := range data.Transaction.Logs {
			if l1.Topics[0] == sdBurnEvent.ID {
				out := new(contracts.RegistrySyntheticDeviceNodeBurned)
				err := p.parseLog(out, sdBurnEvent, l1)
				if err != nil {
					return err
				}

				if _, err := mtr.R.BurnRequestSyntheticDevice.Delete(ctx, p.DB().Writer); err != nil {
					return err
				}

				v, err := models.VehicleNFTS(models.VehicleNFTWhere.TokenID.EQ(types.NewNullDecimal(mtr.R.BurnRequestSyntheticDevice.VehicleTokenID.Big))).One(ctx, p.DB().Reader)
				if err != nil {
					return err
				}

				udai, err := models.FindUserDeviceAPIIntegration(ctx, p.DB().Reader, v.UserDeviceID.String, "22N2xaPOq2WW2gAHBHd0Ikn4Zob")
				if err != nil {
					if err == sql.ErrNoRows {
						return nil
					}
					return err
				}

				_, err = udai.Delete(ctx, p.DB().Writer)
				return err
			}
		}
	}

	return nil
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
	ap *autopi.Integration,
	settings *config.Settings,
	smartcarTaskSvc services.SmartcarTaskService,
	teslaTaskService services.TeslaTaskService,
	deviceDefSvc services.DeviceDefinitionService,
) (StatusProcessor, error) {
	abi, err := contracts.RegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	return &proc{
		ABI:             abi,
		DB:              db,
		Logger:          logger,
		ap:              ap,
		settings:        settings,
		smartcarTaskSvc: smartcarTaskSvc,
		teslaTaskSvc:    teslaTaskService,
		deviceDefSvc:    deviceDefSvc,
	}, nil
}
