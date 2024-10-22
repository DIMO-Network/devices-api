package services

import (
	"context"
	"fmt"
	"time"

	"github.com/DIMO-Network/shared"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

//go:generate mockgen -source user_device_service.go -destination mocks/user_device_service_mock.go -package mock_services

type UserDeviceService interface {
	CreateUserDevice(ctx context.Context, deviceDefID, styleID, countryCode, userID string, vin, canProtocol *string, vinConfirmed bool) (*models.UserDevice, *ddgrpc.GetDeviceDefinitionItemResponse, error)
}

type userDeviceService struct {
	deviceDefSvc DeviceDefinitionService
	log          zerolog.Logger
	dbs          func() *db.ReaderWriter
	eventService EventService
}

func NewUserDeviceService(deviceDefSvc DeviceDefinitionService, log zerolog.Logger, dbs func() *db.ReaderWriter, eventService EventService) UserDeviceService {
	return &userDeviceService{
		deviceDefSvc: deviceDefSvc,
		log:          log,
		dbs:          dbs,
		eventService: eventService,
	}
}

var ErrEmailUnverified = fmt.Errorf("email not verified")

// CreateUserDevice creates the user_device record with all the logic we manage, including setting the countryCode, setting the powertrain based on the def or style, and setting the protocol
func (uds *userDeviceService) CreateUserDevice(ctx context.Context, deviceDefID, styleID, countryCode, userID string, vin, canProtocol *string, vinConfirmed bool) (*models.UserDevice, *ddgrpc.GetDeviceDefinitionItemResponse, error) {
	// attach device def to user
	dd, err2 := uds.deviceDefSvc.GetDeviceDefinitionByID(ctx, deviceDefID)
	if err2 != nil {
		return nil, nil, errors.Wrap(err2, fmt.Sprintf("error querying for device definition id: %s ", deviceDefID))
	}
	powertrainType := ICE // default
	for _, attr := range dd.DeviceAttributes {
		if attr.Name == constants.PowerTrainTypeKey {
			powertrainType = PowertrainType(attr.Value) // todo does this work? validat with test
			break
		}
	}
	// check if style exists to get powertrain
	if len(styleID) > 0 {
		ds, err := uds.deviceDefSvc.GetDeviceStyleByID(ctx, styleID)
		if err != nil {
			// just log warn
			uds.log.Warn().Err(err).Msgf("failed to get device style %s - continuing", styleID)
		}

		if ds != nil && len(ds.DeviceAttributes) > 0 {
			// Find device attribute (powertrain_type)
			for _, item := range ds.DeviceAttributes {
				if item.Name == constants.PowerTrainTypeKey {
					powertrainType = ConvertPowerTrainStringToPowertrain(item.Value)
					break
				}
			}
		}
	}

	tx, err := uds.dbs().Writer.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback() //nolint

	userDeviceID := ksuid.New().String()
	// register device for the user
	ud := models.UserDevice{
		ID:                 userDeviceID,
		UserID:             userID,
		DeviceDefinitionID: dd.DeviceDefinitionId,
		CountryCode:        null.StringFrom(countryCode),
		VinIdentifier:      null.StringFromPtr(vin),
		VinConfirmed:       vinConfirmed,
	}
	// always instantiate metadata with powerTrain and CANProtocol
	udMD := &UserDeviceMetadata{
		PowertrainType: &powertrainType,
	}
	if canProtocol != nil && len(*canProtocol) > 0 {
		udMD.CANProtocol = canProtocol
	}
	err = ud.Metadata.Marshal(udMD)
	if err != nil {
		uds.log.Warn().Str("func", "createUserDevice").Msg("failed to marshal user device metadata on create")
	}

	err = ud.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "could not create user device for def_id: "+dd.DeviceDefinitionId)
	}

	err = tx.Commit() // commmit the transaction
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error commiting transaction to create geofence")
	}

	// todo call devide definitions to check and pull image for this device in case don't have one

	err = uds.eventService.Emit(&shared.CloudEvent[any]{
		Type:    constants.UserDeviceCreationEventType,
		Subject: userID,
		Source:  "devices-api",
		Data: UserDeviceEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: UserDeviceEventDevice{
				ID:    userDeviceID,
				Make:  dd.Make.Name,
				Model: dd.Type.Model,
				Year:  int(dd.Type.Year), // Odd.
			},
		},
	})
	if err != nil {
		uds.log.Err(err).Msg("Failed emitting device creation event")
	}
	return &ud, dd, nil
}
