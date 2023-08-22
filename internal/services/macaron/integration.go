package autopi

import (
	"context"
	"database/sql"
	"math/big"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
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

type Integration struct {
	db          func() *db.ReaderWriter
	defs        services.DeviceDefinitionService
	apReg       services.IngestRegistrar
	eventer     services.EventService
	ddRegistrar services.DeviceDefinitionRegistrar
	logger      *zerolog.Logger
}

func NewIntegration(
	db func() *db.ReaderWriter,
	defs services.DeviceDefinitionService,
	ap services.AutoPiAPIService,
	apTask services.AutoPiTaskService,
	apReg services.IngestRegistrar,
	eventer services.EventService,
	ddRegistrar services.DeviceDefinitionRegistrar,
	logger *zerolog.Logger,
) *Integration {
	return &Integration{
		db:          db,
		defs:        defs,
		apReg:       apReg,
		eventer:     eventer,
		ddRegistrar: ddRegistrar,
		logger:      logger,
	}
}

func intToDec(x *big.Int) types.NullDecimal {
	return types.NewNullDecimal(new(decimal.Big).SetBigMantScale(x, 0))
}

func powertrainToTemplate(pt *services.PowertrainType, integ *ddgrpc.Integration) int32 {
	out := integ.AutoPiDefaultTemplateId
	if pt != nil {
		switch *pt {
		case services.ICE:
			out = integ.AutoPiPowertrainTemplate.ICE
		case services.HEV:
			out = integ.AutoPiPowertrainTemplate.HEV
		case services.PHEV:
			out = integ.AutoPiPowertrainTemplate.PHEV
		case services.BEV:
			out = integ.AutoPiPowertrainTemplate.BEV
		}
	}
	return out
}

func (i *Integration) Pair(ctx context.Context, amTokenID, vehicleTokenID *big.Int) error {
	tx, err := i.db().Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	integ, err := i.defs.GetIntegrationByVendor(ctx, "Macaron")
	if err != nil {
		return err
	}

	amDev, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.TokenID.EQ(intToDec(amTokenID)),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(intToDec(vehicleTokenID)),
		qm.Load(models.VehicleNFTRels.UserDevice),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	if nft.R.UserDevice == nil {
		return errors.New("vehicle deleted")
	}

	ud := nft.R.UserDevice

	oldInt, err := models.FindUserDeviceAPIIntegration(ctx, tx, ud.ID, integ.Id)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else {
		// We might be using this function to "repair" a connection. Just kill the old row.
		_, err = oldInt.Delete(ctx, tx)
		if err != nil {
			return err
		}
	}

	def, err := i.defs.GetDeviceDefinitionByID(ctx, ud.DeviceDefinitionID)
	if err != nil {
		return err
	}

	udai := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: integ.Id,
		ExternalID:    null.StringFrom(amDev.Serial),
		Status:        models.UserDeviceAPIIntegrationStatusPending,
		Serial:        null.StringFrom(amDev.Serial),
	}
	if err = udai.Insert(ctx, tx, boil.Infer()); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit new autopi integration")
	}

	err = i.apReg.Register2(&services.AftermarketDeviceVehicleMapping{
		AftermarketDevice: services.AftermarketDeviceVehicleMappingAftermarketDevice{
			Address:       common.BytesToAddress(amDev.EthereumAddress),
			Token:         amTokenID,
			Serial:        amDev.Serial,
			IntegrationID: integ.Id,
		},
		Vehicle: services.AftermarketDeviceVehicleMappingVehicle{
			Token:        vehicleTokenID,
			UserDeviceID: ud.ID,
		},
	})
	if err != nil {
		return err
	}

	_ = i.eventer.Emit(
		&services.Event{
			Type:    "com.dimo.zone.device.integration.create",
			Source:  "devices-api",
			Subject: ud.ID,
			Data: services.UserDeviceIntegrationEvent{
				Timestamp: time.Now(),
				UserID:    ud.UserID,
				Device: services.UserDeviceEventDevice{
					ID:                 ud.ID,
					DeviceDefinitionID: def.DeviceDefinitionId,
					Make:               def.Type.Make,
					Model:              def.Type.Model,
					Year:               int(def.Type.Year),
					VIN:                ud.VinIdentifier.String,
				},
				Integration: services.UserDeviceEventIntegration{
					ID:     integ.Id,
					Type:   integ.Type,
					Style:  integ.Style,
					Vendor: integ.Vendor,
				},
			},
		},
	)

	region := ""
	if ud.CountryCode.Valid {
		countryRecord := constants.FindCountry(ud.CountryCode.String)
		if countryRecord != nil {
			region = countryRecord.Region
		}
	}
	_ = i.ddRegistrar.Register(services.DeviceDefinitionDTO{
		IntegrationID:      integ.Id,
		UserDeviceID:       ud.ID,
		DeviceDefinitionID: ud.DeviceDefinitionID,
		Make:               def.Type.Make,
		Model:              def.Type.Model,
		Year:               int(def.Type.Year),
		Region:             region,
		MakeSlug:           def.Type.MakeSlug,
		ModelSlug:          def.Type.ModelSlug,
	})

	return nil
}

func (i *Integration) Unpair(ctx context.Context, autoPiTokenID, vehicleTokenID *big.Int) error {
	tx, err := i.db().Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(intToDec(vehicleTokenID)),
		qm.Load(models.VehicleNFTRels.UserDevice),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	if nft.R.UserDevice == nil {
		return errors.New("vehicle deleted")
	}

	ud := nft.R.UserDevice

	autoPiModel, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.TokenID.EQ(intToDec(autoPiTokenID)),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	integ, err := i.defs.GetIntegrationByVendor(ctx, "Macaron")
	if err != nil {
		return err
	}

	udai, err := models.FindUserDeviceAPIIntegration(ctx, i.db().Writer, ud.ID, integ.Id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	} else {
		_, err = udai.Delete(ctx, i.db().Writer)
		if err != nil {
			return err
		}
	}

	err = i.apReg.Deregister2(common.BytesToAddress(autoPiModel.EthereumAddress))
	if err != nil {
		return err
	}

	def, err := i.defs.GetDeviceDefinitionByID(ctx, ud.DeviceDefinitionID)
	if err != nil {
		return err
	}

	_ = i.eventer.Emit(&services.Event{
		Type:    "com.dimo.zone.device.integration.delete",
		Source:  "devices-api",
		Subject: ud.ID,
		Data: services.UserDeviceIntegrationEvent{
			Timestamp: time.Now(),
			UserID:    ud.UserID,
			Device: services.UserDeviceEventDevice{
				ID:    ud.ID,
				Make:  def.Make.Name,
				Model: def.Type.Model,
				Year:  int(def.Type.Year),
			},
			Integration: services.UserDeviceEventIntegration{
				ID:     integ.Id,
				Type:   integ.Type,
				Style:  integ.Style,
				Vendor: integ.Vendor,
			},
		},
	})

	return nil
}
