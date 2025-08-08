package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/shared/pkg/payloads"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

//go:generate mockgen -source user_device_service.go -destination mocks/user_device_service_mock.go -package mock_services

type UserDeviceService interface {
	CreateUserDevice(ctx context.Context, definitionID, styleID, countryCode, userID string, vin, canProtocol *string, vinConfirmed bool) (*models.UserDevice, *ddgrpc.GetDeviceDefinitionItemResponse, error)
	CreateUserDeviceByOwner(ctx context.Context, definitionID, styleID, countryCode, vin string, ownerAddress []byte) (*models.UserDevice, *ddgrpc.GetDeviceDefinitionItemResponse, error)
	CreateIntegration(ctx context.Context, tx *sql.Tx, userDeviceID string, integrationID string, externalID string, encryptedAccessToken string, accessExpiry time.Time, encryptedRefreshToken string, metadata []byte) error
}

type userDeviceService struct {
	deviceDefSvc DeviceDefinitionService
	log          zerolog.Logger
	dbs          func() *db.ReaderWriter
	eventService EventService
}

func (uds *userDeviceService) CreateIntegration(ctx context.Context, tx *sql.Tx, userDeviceID, integrationID, externalID, encryptedAccessToken string,
	accessExpiry time.Time, encryptedRefreshToken string, metadata []byte) error {

	taskID := ksuid.New().String()
	integration := &models.UserDeviceAPIIntegration{
		TaskID:          null.StringFrom(taskID),
		ExternalID:      null.StringFrom(externalID),
		UserDeviceID:    userDeviceID,
		IntegrationID:   integrationID,
		Status:          models.UserDeviceAPIIntegrationStatusPendingFirstData,
		AccessToken:     null.StringFrom(encryptedAccessToken),
		AccessExpiresAt: null.TimeFrom(accessExpiry),
		RefreshToken:    null.StringFrom(encryptedRefreshToken),
		Metadata:        null.JSONFrom(metadata),
	}

	if err := integration.Insert(ctx, tx, boil.Infer()); err != nil {
		return errors.Wrap(err, "unexpected database error inserting new integration")
	}
	return nil
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

// CreateUserDeviceByOwner same as below but uses owner wallet address. Currently only being used by admin.
func (uds *userDeviceService) CreateUserDeviceByOwner(ctx context.Context, definitionID, styleID, countryCode, vin string, ownerAddress []byte) (*models.UserDevice, *ddgrpc.GetDeviceDefinitionItemResponse, error) {
	if len(definitionID) == 0 {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "definitionID is empty")
	}

	// attach device def to user
	dd, err2 := uds.deviceDefSvc.GetDeviceDefinitionBySlug(ctx, definitionID)
	if err2 != nil {
		return nil, nil, errors.Wrap(err2, fmt.Sprintf("error querying for device definition id: %s ", definitionID))
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

	tx, err := uds.dbs().Writer.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback() //nolint

	userDeviceID := ksuid.New().String()
	// validate country_code
	if constants.FindCountry(countryCode) == nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "CountryCode does not exist: "+countryCode)
	}
	// register device for the user
	ud := models.UserDevice{
		ID:            userDeviceID,
		OwnerAddress:  null.BytesFrom(ownerAddress),
		UserID:        "", // normally the dex user id but since lookup is by owner address hoping this can be blank
		DefinitionID:  dd.Id,
		CountryCode:   null.StringFrom(countryCode),
		VinIdentifier: null.StringFrom(vin),
		VinConfirmed:  true,
	}
	// always instantiate metadata with powerTrain and CANProtocol
	udMD := &UserDeviceMetadata{
		PowertrainType: &powertrainType,
	}
	err = ud.Metadata.Marshal(udMD)
	if err != nil {
		uds.log.Warn().Str("func", "createUserDevice").Msg("failed to marshal user device metadata on create")
	}

	err = ud.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "could not create user device for definition_id: "+dd.Id)
	}

	err = tx.Commit() // commmit the transaction
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error commiting transaction")
	}

	// todo call devide definitions to check and pull image for this device in case don't have one

	err = uds.eventService.Emit(&payloads.CloudEvent[any]{
		Type:    constants.UserDeviceCreationEventType,
		Subject: common.BytesToAddress(ownerAddress).Hex(),
		Source:  "devices-api",
		Data: UserDeviceEvent{
			Timestamp: time.Now(),
			UserID:    common.BytesToAddress(ownerAddress).Hex(),
			Device: UserDeviceEventDevice{
				ID:           userDeviceID,
				Make:         dd.Make.Name,
				Model:        dd.Model,
				Year:         int(dd.Year), // Odd.
				DefinitionID: dd.Id,
			},
		},
	})
	if err != nil {
		uds.log.Err(err).Msg("Failed emitting device creation event")
	}
	return &ud, dd, nil
}

// CreateUserDevice creates the user_device record with all the logic we manage, including setting the countryCode, setting the powertrain based on the def or style, and setting the protocol
func (uds *userDeviceService) CreateUserDevice(ctx context.Context, definitionID, styleID, countryCode, userID string, vin, canProtocol *string, vinConfirmed bool) (*models.UserDevice, *ddgrpc.GetDeviceDefinitionItemResponse, error) {
	if len(definitionID) == 0 {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "definitionID is empty")
	}

	// attach device def to user
	dd, err2 := uds.deviceDefSvc.GetDeviceDefinitionBySlug(ctx, definitionID)
	if err2 != nil {
		return nil, nil, errors.Wrap(err2, fmt.Sprintf("error querying for device definition id: %s ", definitionID))
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

	tx, err := uds.dbs().Writer.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback() //nolint

	userDeviceID := ksuid.New().String()
	// validate country_code
	if constants.FindCountry(countryCode) == nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "CountryCode does not exist: "+countryCode)
	}
	// register device for the user
	ud := models.UserDevice{
		ID:            userDeviceID,
		UserID:        userID,
		CountryCode:   null.StringFrom(countryCode),
		VinIdentifier: null.StringFromPtr(vin),
		VinConfirmed:  vinConfirmed,
		DefinitionID:  dd.Id,
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
		return nil, nil, errors.Wrapf(err, "error commiting transaction")
	}

	// todo call devide definitions to check and pull image for this device in case don't have one

	err = uds.eventService.Emit(&payloads.CloudEvent[any]{
		Type:    constants.UserDeviceCreationEventType,
		Subject: userID,
		Source:  "devices-api",
		Data: UserDeviceEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: UserDeviceEventDevice{
				ID:           userDeviceID,
				Make:         dd.Make.Name,
				Model:        dd.Model,
				Year:         int(dd.Year), // Odd.
				DefinitionID: dd.Id,
			},
		},
	})
	if err != nil {
		uds.log.Err(err).Msg("Failed emitting device creation event")
	}
	return &ud, dd, nil
}
