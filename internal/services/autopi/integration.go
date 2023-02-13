package autopi

import (
	"context"
	"database/sql"
	"math/big"
	"strconv"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type Integration struct {
	db                      func() *db.ReaderWriter
	defs                    services.DeviceDefinitionService
	ap                      services.AutoPiAPIService
	apTask                  services.AutoPiTaskService
	apReg                   services.IngestRegistrar
	eventer                 services.EventService
	ddRegistrar             services.DeviceDefinitionRegistrar
	hardwareTemplateService HardwareTemplateService
	logger                  *zerolog.Logger
}

func NewIntegration(
	db func() *db.ReaderWriter,
	defs services.DeviceDefinitionService,
	ap services.AutoPiAPIService,
	apTask services.AutoPiTaskService,
	apReg services.IngestRegistrar,
	eventer services.EventService,
	ddRegistrar services.DeviceDefinitionRegistrar,
	hardwareTemplateService HardwareTemplateService,
	logger *zerolog.Logger,
) *Integration {
	return &Integration{
		db:                      db,
		defs:                    defs,
		ap:                      ap,
		apTask:                  apTask,
		apReg:                   apReg,
		eventer:                 eventer,
		ddRegistrar:             ddRegistrar,
		hardwareTemplateService: hardwareTemplateService,
		logger:                  logger,
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

func (i *Integration) Pair(ctx context.Context, autoPiTokenID, vehicleTokenID *big.Int) error {
	tx, err := i.db().Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	integ, err := i.defs.GetIntegrationByVendor(ctx, "AutoPi")
	if err != nil {
		return err
	}

	var autoPiUnitID string

	autoPiModel, err := models.AutopiUnits(
		models.AutopiUnitWhere.TokenID.EQ(intToDec(autoPiTokenID)),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	autoPiUnitID = autoPiModel.AutopiUnitID

	autoPi, err := i.ap.GetDeviceByUnitID(autoPiUnitID)
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

	hardwareTemplate, err := i.hardwareTemplateService.GetTemplateID(ud, def, integ)

	if err != nil {
		return err
	}

	templateID, _ := strconv.Atoi(hardwareTemplate)

	udMetadata := services.UserDeviceAPIIntegrationsMetadata{
		AutoPiUnitID:          &autoPi.UnitID,
		AutoPiIMEI:            &autoPi.IMEI,
		AutoPiTemplateApplied: &templateID,
	}

	apiInt := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: integ.Id,
		ExternalID:    null.StringFrom(autoPi.ID),
		Status:        models.UserDeviceAPIIntegrationStatusPending,
		AutopiUnitID:  null.StringFrom(autoPi.UnitID),
	}

	err = apiInt.Metadata.Marshal(udMetadata)
	if err != nil {
		return err
	}

	if err = apiInt.Insert(ctx, tx, boil.Infer()); err != nil {
		return err
	}

	substatus := constants.QueriedDeviceOk
	// update integration record as failed if errors after this
	defer func() {
		if err != nil {
			apiInt.Status = models.UserDeviceAPIIntegrationStatusFailed
			msg := err.Error()
			udMetadata.AutoPiRegistrationError = &msg
			ss := substatus.String()
			udMetadata.AutoPiSubStatus = &ss
			_ = apiInt.Metadata.Marshal(udMetadata)
			_, _ = apiInt.Update(ctx, tx,
				boil.Whitelist(models.UserDeviceAPIIntegrationColumns.Status, models.UserDeviceAPIIntegrationColumns.UpdatedAt))
			err = tx.Commit()
		}
	}()
	// update the profile on autopi
	profile := services.PatchVehicleProfile{
		Year: int(def.Type.Year),
	}

	if ud.Name.Valid {
		profile.CallName = ud.Name.String
	}
	err = i.ap.PatchVehicleProfile(autoPi.Vehicle.ID, profile)
	if err != nil {
		return errors.Wrap(err, "failed to patch autopi vehicle profile")
	}

	substatus = constants.PatchedVehicleProfile
	// update autopi to unassociate from current base template
	if autoPi.Template > 0 {
		err = i.ap.UnassociateDeviceTemplate(autoPi.ID, autoPi.Template)
		if err != nil {
			return errors.Wrapf(err, "failed to unassociate template %d", autoPi.Template)
		}
	}
	// set our template on the autoPiDevice
	err = i.ap.AssociateDeviceToTemplate(autoPi.ID, templateID)
	if err != nil {
		return errors.Wrapf(err, "failed to associate autoPiDevice %s to template %d", autoPi.ID, templateID)
	}
	substatus = constants.AssociatedDeviceToTemplate
	// apply for next reboot
	err = i.ap.ApplyTemplate(autoPi.ID, templateID)
	if err != nil {
		return errors.Wrapf(err, "failed to apply autoPiDevice %s with template %d", autoPi.ID, templateID)
	}

	substatus = constants.AppliedTemplate

	// send sync command in case autoPiDevice is on at this moment (should be during initial setup)
	_, err = i.ap.CommandSyncDevice(ctx, autoPi.UnitID, autoPi.ID, ud.ID)
	if err != nil {
		i.logger.Warn().Err(err).Msg("Failed to send sync command to AutoPi.")
	} else {
		substatus = constants.PendingTemplateConfirm
	}

	ss := substatus.String() // This is either going to be AppliedTemplate or PendingTemplateConfirm, at this point.
	udMetadata.AutoPiSubStatus = &ss
	err = apiInt.Metadata.Marshal(udMetadata)
	if err != nil {
		return errors.Wrap(err, "failed to marshall user device integration metadata")
	}

	_, err = apiInt.Update(ctx, tx, boil.Whitelist(models.UserDeviceAPIIntegrationColumns.Metadata,
		models.UserDeviceAPIIntegrationColumns.UpdatedAt))
	if err != nil {
		return errors.Wrap(err, "failed to update integration status to Pending")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit new autopi integration")
	}

	// send kafka message to autopi ingest registrar. Note we're using the UnitID for the data stream join.
	err = i.apReg.Register(autoPi.UnitID, ud.ID, integ.Id)
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

	autoPiModel, err := models.AutopiUnits(
		models.AutopiUnitWhere.TokenID.EQ(intToDec(autoPiTokenID)),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	integ, err := i.defs.GetIntegrationByVendor(ctx, "AutoPi")
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

	err = i.apReg.Deregister(autoPiModel.AutopiDeviceID.String, ud.ID, integ.Id)
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
