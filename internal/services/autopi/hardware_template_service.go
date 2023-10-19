package autopi

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rs/zerolog"

	"github.com/pkg/errors"

	"github.com/volatiletech/null/v8"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type HardwareTemplateService interface {
	GetTemplateID(ud *models.UserDevice, dd *ddgrpc.GetDeviceDefinitionItemResponse, integ *ddgrpc.Integration) (string, error)
	ApplyHardwareTemplate(ctx context.Context, req *pb.ApplyHardwareTemplateRequest) (*pb.ApplyHardwareTemplateResponse, error)
	CreateTemplate(req *pb.CreateTemplateRequest) (*pb.CreateTemplateResponse, error)
}

type hardwareTemplateService struct {
	dbs    func() *db.ReaderWriter
	ap     services.AutoPiAPIService
	logger *zerolog.Logger
}

func NewHardwareTemplateService(ap services.AutoPiAPIService, dbs func() *db.ReaderWriter, logger *zerolog.Logger) HardwareTemplateService {
	return &hardwareTemplateService{
		ap:     ap,
		dbs:    dbs,
		logger: logger,
	}
}

func (a *hardwareTemplateService) GetTemplateID(ud *models.UserDevice, dd *ddgrpc.GetDeviceDefinitionItemResponse, integ *ddgrpc.Integration) (string, error) {
	const defaultTemplate = "10" // if for some reason get an empty or 0 template value, always return this.
	// get template from device style, only if UD has a DS set and the DS has a templateID set
	if ud.DeviceStyleID.Valid {
		if len(dd.DeviceStyles) > 0 {
			for _, ds := range dd.DeviceStyles {
				if ds.Id == ud.DeviceStyleID.String {
					if isTemplateIDValid(ds.HardwareTemplateId) {
						return ds.HardwareTemplateId, nil
					}
				}
			}
		}
	}

	// get template from Device Definition
	if isTemplateIDValid(dd.HardwareTemplateId) {
		return dd.HardwareTemplateId, nil
	}

	// get template from Make
	if isTemplateIDValid(dd.Make.HardwareTemplateId) {
		return dd.Make.HardwareTemplateId, nil
	}

	// get template from powertrain based on map in integration metadata
	if integ.AutoPiPowertrainTemplate != nil {
		udMd := services.UserDeviceMetadata{}
		err := ud.Metadata.Unmarshal(&udMd)
		if err != nil {
			return defaultTemplate, err
		}

		tIDFromPowerTrain := powertrainToTemplate(udMd.PowertrainType, integ)
		if tIDFromPowerTrain > 0 {
			return strconv.Itoa(int(tIDFromPowerTrain)), nil
		}
	}

	// get template from autopi integration default
	if integ.AutoPiDefaultTemplateId > 0 {
		return strconv.Itoa(int(integ.AutoPiDefaultTemplateId)), nil
	}
	a.logger.Warn().Str("user_device_id", ud.ID).Str("device_definition_id", dd.DeviceDefinitionId).
		Msgf("could not find a templateID for this user_device")

	return defaultTemplate, nil
}

// isTemplateIDValid returns true if not empty and can be converted to a number, otherwise returns false
func isTemplateIDValid(templateID string) bool {
	if len(templateID) > 0 {
		// currently assume template must be numeric
		t, err := strconv.Atoi(templateID)
		if err == nil && t > 0 {
			return true
		}
	}
	return false
}

func (a *hardwareTemplateService) CreateTemplate(req *pb.CreateTemplateRequest) (*pb.CreateTemplateResponse, error) {
	newTemplateID, err := a.ap.CreateNewTemplate(req.Name, int(req.Parent), req.Name)
	if err == nil && newTemplateID > 0 {
		err := a.ap.SetTemplateICEPowerSettings(newTemplateID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		err = a.ap.AddDefaultPIDsToTemplate(newTemplateID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateTemplateResponse{Id: int64(newTemplateID)}, err
}

func (a *hardwareTemplateService) ApplyHardwareTemplate(ctx context.Context, req *pb.ApplyHardwareTemplateRequest) (*pb.ApplyHardwareTemplateResponse, error) {
	tx, err := a.dbs().Writer.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	udapi, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(req.UserDeviceId),
		models.UserDeviceAPIIntegrationWhere.Serial.EQ(null.StringFrom(req.AutoApiUnitId)),
	).One(ctx, tx)
	if err != nil {
		return nil, err
	}

	autoPiModel, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.Serial.EQ(req.AutoApiUnitId),
	).One(ctx, tx)
	if err != nil {
		return nil, err
	}

	autoPi, err := a.ap.GetDeviceByUnitID(autoPiModel.Serial)
	if err != nil {
		return nil, err
	}

	if autoPi.Template > 0 {
		err = a.ap.UnassociateDeviceTemplate(autoPi.ID, autoPi.Template)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to unassociate template %d on device %s", autoPi.Template, autoPi.ID))
		}
		vehicleLoggers, err := a.ap.GetVehicleLoggers(autoPi.Vehicle.ID)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to get vehicle loggers for vehicleID %d", autoPi.Vehicle.ID))
		} else {
			for _, logger := range *vehicleLoggers {
				if logger.Parent == nil { // IF (logger was not assigned via template)
					// Autopi's API returns a NULL value for "AutoPiVehicleLogger.Parent" when the logger was not assigned via template
					err := a.ap.DeleteVehicleLogger(logger.LoggerType, logger.ID)
					if err != nil {
						return nil, errors.Wrap(err, fmt.Sprintf("failed to delete vehicle logger %d", logger.ID))
					}
				}
			}
		}

	}

	hardwareTemplateID, err := strconv.Atoi(req.HardwareTemplateId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert template id %s to integer", req.HardwareTemplateId)
	}

	// set our template on the autoPiDevice
	err = a.ap.AssociateDeviceToTemplate(autoPi.ID, hardwareTemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to associate autoPiDevice %s to template %d", autoPi.ID, hardwareTemplateID)
	}

	// apply for next reboot
	err = a.ap.ApplyTemplate(autoPi.ID, hardwareTemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to apply autoPiDevice %s with template %d", autoPi.ID, hardwareTemplateID)
	}

	// send sync command in case autoPiDevice is on at this moment (should be during initial setup)
	_, err = a.ap.CommandSyncDevice(ctx, autoPi.UnitID, autoPi.ID, req.UserDeviceId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to sync changes to autoPiDevice %s", autoPi.ID)
	}

	udMetadata := services.UserDeviceAPIIntegrationsMetadata{
		AutoPiUnitID:          &autoPi.UnitID,
		AutoPiIMEI:            &autoPi.IMEI,
		AutoPiTemplateApplied: &hardwareTemplateID,
	}

	err = udapi.Metadata.Marshal(udMetadata)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshall user device integration metadata")
	}

	_, err = udapi.Update(ctx, tx, boil.Whitelist(models.UserDeviceColumns.Metadata, models.UserDeviceColumns.UpdatedAt))
	if err != nil {
		return nil, errors.Wrap(err, "failed to update user device status to Pending")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit new hardware template to user device")
	}

	return &pb.ApplyHardwareTemplateResponse{Applied: true}, nil
}
