package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services/issuer"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/lovoo/goka"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	deviceStatusEventType = "zone.dimo.device.status.update"
	odometerCooldown      = time.Hour
)

type DeviceStatusIngestService struct {
	db           func() *db.ReaderWriter
	log          *zerolog.Logger
	eventService EventService
	deviceDefSvc DeviceDefinitionService
	integrations []*grpc.Integration
	autoPiSvc    AutoPiAPIService
	memoryCache  *gocache.Cache
	issuer       *issuer.Issuer
}

func NewDeviceStatusIngestService(db func() *db.ReaderWriter, log *zerolog.Logger, eventService EventService, ddSvc DeviceDefinitionService, autoPiSvc AutoPiAPIService, settings *config.Settings) *DeviceStatusIngestService {
	// Cache the list of integrations.
	integrations, err := ddSvc.GetIntegrations(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't retrieve global integration list.")
	}
	c := gocache.New(30*time.Minute, 60*time.Minute) // band-aid on top of band-aids

	pk, err := base64.RawURLEncoding.DecodeString(settings.IssuerPrivateKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't parse issuer private key.")
	}

	issuer, err := issuer.New(
		issuer.Config{
			PrivateKey:        pk,
			ChainID:           big.NewInt(settings.DIMORegistryChainID),
			VehicleNFTAddress: common.HexToAddress(settings.VehicleNFTAddress),
			DBS:               db,
		},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create issuer.")
	}

	return &DeviceStatusIngestService{
		db:           db,
		log:          log,
		deviceDefSvc: ddSvc,
		eventService: eventService,
		integrations: integrations,
		autoPiSvc:    autoPiSvc,
		memoryCache:  c,
		issuer:       issuer,
	}
}

// ProcessDeviceStatusMessages works on channel stream of messages from watermill kafka consumer
func (i *DeviceStatusIngestService) ProcessDeviceStatusMessages(ctx goka.Context, msg interface{}) {
	if err := i.processMessage(ctx, msg.(*DeviceStatusEvent)); err != nil {
		i.log.Err(err).Msg("Error processing device status message.")
	}
}

func (i *DeviceStatusIngestService) processMessage(ctx goka.Context, event *DeviceStatusEvent) error {
	if event.Type != deviceStatusEventType {
		return fmt.Errorf("received vehicle status event with unexpected type %s", event.Type)
	}

	integration, err := i.getIntegrationFromEvent(event)
	if err != nil {
		return err
	}

	switch integration.Vendor {
	case constants.SmartCarVendor:
		defer appmetrics.SmartcarIngestTotalOps.Inc()
	case constants.AutoPiVendor:
		defer appmetrics.AutoPiIngestTotalOps.Inc()
	}

	return i.processEvent(ctx, event)
}

var userDeviceDataPrimaryKeyColumns = []string{models.UserDeviceDatumColumns.UserDeviceID, models.UserDeviceDatumColumns.IntegrationID}

// processEvent handles the device data status update so we have a latest snapshot. This should all be refactored to device data api.
func (i *DeviceStatusIngestService) processEvent(ctxGk goka.Context, event *DeviceStatusEvent) error {
	ctx := context.Background()
	userDeviceID := event.Subject

	integration, err := i.getIntegrationFromEvent(event)
	if err != nil {
		return err
	}

	device := &models.UserDevice{}
	get, found := i.memoryCache.Get(userDeviceID + "_" + integration.Id)

	if found {
		device = get.(*models.UserDevice)
	} else {
		device, err = models.UserDevices(
			models.UserDeviceWhere.ID.EQ(userDeviceID),
			qm.Load(
				models.UserDeviceRels.UserDeviceAPIIntegrations,
				models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integration.Id),
			),
			qm.Load(
				models.UserDeviceRels.UserDeviceData,
				models.UserDeviceDatumWhere.IntegrationID.EQ(integration.Id),
			),
		).One(ctx, i.db().Reader)

		if err != nil {
			return fmt.Errorf("failed to find device: %w", err)
		}
		i.memoryCache.Set(userDeviceID+"_"+integration.Id, device, 30*time.Minute)
	}

	i.vinCredentialer(ctxGk, event, device)

	if len(device.R.UserDeviceAPIIntegrations) == 0 {
		return fmt.Errorf("can't find API integration for device %s and integration %s", userDeviceID, integration.Id)
	}

	deviceDefinitionResponse, err := i.deviceDefSvc.GetDeviceDefinitionsByIDs(ctx, []string{device.DeviceDefinitionID})
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+device.DeviceDefinitionID)
	}

	if len(deviceDefinitionResponse) == 0 {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("device definition with id %s not found", device.DeviceDefinitionID))
	}

	dd := deviceDefinitionResponse[0]

	// update status to Active if not already set
	apiIntegration := device.R.UserDeviceAPIIntegrations[0]
	if apiIntegration.Status != models.UserDeviceAPIIntegrationStatusActive {
		apiIntegration.Status = models.UserDeviceAPIIntegrationStatusActive
		if _, err := apiIntegration.Update(ctx, i.db().Writer, boil.Infer()); err != nil {
			return fmt.Errorf("failed to update API integration: %w", err)
		}

		if integration.Vendor == constants.AutoPiVendor {
			err := i.autoPiSvc.UpdateState(apiIntegration.ExternalID.String, apiIntegration.Status)
			if err != nil {
				return fmt.Errorf("failed to update status when calling autopi api for deviceId: %s", apiIntegration.ExternalID.String)
			}
		}
		i.memoryCache.Delete(userDeviceID + "_" + integration.Id)
	}

	// techdebt: could likely get rid of this with tweak in app so that people just see that data came through - not specific to odometer
	// Null for most AutoPis.
	var newOdometer null.Float64
	if o, err := extractOdometer(event.Data); err == nil {
		newOdometer = null.Float64From(o)
	} else if integration.Vendor == constants.AutoPiVendor {
		// For AutoPis, for the purpose of odometer events we are pretending to always have
		// an odometer reading. Users became accustomed to seeing the associated events, even
		// though we mostly don't have odometer readings for AutoPis. For now, we fake it.

		// Update PLA-934:  Now that we are starting to receive real odometer values from
		//             		the AutoPi, we need the real odometer timestamp. To avoid alarming
		//					users as mentioned above, we resolved to create another column
		//					called "real_last_odometer_event_at" to store this value
		newOdometer = null.Float64From(0.0)
	}

	var datum *models.UserDeviceDatum
	if len(device.R.UserDeviceData) > 0 {
		// Update the existing record.
		datum = device.R.UserDeviceData[0]
	} else {
		// Insert a new record.
		datum = &models.UserDeviceDatum{UserDeviceID: userDeviceID, IntegrationID: integration.Id}
		i.memoryCache.Delete(userDeviceID + "_" + integration.Id)
	}

	i.processOdometer(datum, newOdometer, device, dd, integration.Id)

	// Not every update has every signal. Merge the new into the old.
	compositeData := make(map[string]any)
	if err := datum.Data.Unmarshal(&compositeData); err != nil {
		return err
	}

	// This will preserve any mappings with keys present in datum.Data but not in
	// event.Data. If a key is present in both maps then the value from event.Data
	// takes precedence.
	//
	// For example, if in the database we have {A: 1, B: 2} and the new event has
	// {B: 4, C: 9} then the result should be {A: 1, B: 4, C: 9}.
	if err := json.Unmarshal(event.Data, &compositeData); err != nil {
		return err
	}

	if err := datum.Data.Marshal(compositeData); err != nil {
		return err
	}
	datum.ErrorData = null.JSON{}

	// extract signals with timestamps and persist to signals
	existingSignalData := make(map[string]any)
	if err := datum.Signals.Unmarshal(&existingSignalData); err != nil {
		return err
	}
	// unmarshall only the event data
	eventData := make(map[string]any)
	err = json.Unmarshal(event.Data, &eventData)
	if err != nil {
		return errors.Wrap(err, "could not unmarshall event data")
	}
	newSignals, err := mergeSignals(existingSignalData, eventData, event.Time)
	if err != nil {
		return err
	}
	if err := datum.Signals.Marshal(newSignals); err != nil {
		return err
	}

	if err := datum.Upsert(ctx, i.db().Writer, true, userDeviceDataPrimaryKeyColumns, boil.Infer(), boil.Infer()); err != nil {
		return fmt.Errorf("error upserting datum: %w", err)
	}

	switch integration.Vendor {
	case constants.SmartCarVendor:
		appmetrics.SmartcarIngestSuccessOps.Inc()
	case constants.AutoPiVendor:
		appmetrics.AutoPiIngestSuccessOps.Inc()
	}

	return nil
}

// processOdometer emits an odometer event and updates the last_odometer_event timestamp on the
// data record if the following conditions are met:
//   - there is no existing timestamp, or an hour has passed since that timestamp,
//   - the incoming status update has an odometer value, and
//   - the old status update lacks an odometer value, or has an odometer value that differs from
//     the new odometer reading
func (i *DeviceStatusIngestService) processOdometer(datum *models.UserDeviceDatum, newOdometer null.Float64, device *models.UserDevice, dd *grpc.GetDeviceDefinitionItemResponse, integrationID string) {
	if !newOdometer.Valid {
		return
	}

	var oldOdometer null.Float64
	if datum.Data.Valid {
		if o, err := extractOdometer(datum.Data.JSON); err == nil {
			oldOdometer = null.Float64From(o)
		}
	}

	now := time.Now()
	odometerOffCooldown := !datum.LastOdometerEventAt.Valid || now.Sub(datum.LastOdometerEventAt.Time) >= odometerCooldown
	odometerChanged := !oldOdometer.Valid || newOdometer.Float64 > oldOdometer.Float64

	if odometerOffCooldown && odometerChanged {
		datum.LastOdometerEventAt = null.TimeFrom(now)
		if newOdometer.Float64 > 0.01 {
			// Since this function will always receive 0.0 for odo if not present
			// if odometer value is 0 then it must have been fake
			datum.RealLastOdometerEventAt = null.TimeFrom(now)
		}
		i.emitOdometerEvent(device, dd, integrationID, newOdometer.Float64)
	}

}

func (i *DeviceStatusIngestService) emitOdometerEvent(device *models.UserDevice, dd *grpc.GetDeviceDefinitionItemResponse, integrationID string, odometer float64) {
	event := &Event{
		Type:    "com.dimo.zone.device.odometer.update",
		Subject: device.ID,
		Source:  "dimo/integration/" + integrationID,
		Data: OdometerEvent{
			Timestamp: time.Now(),
			UserID:    device.UserID,
			Device: odometerEventDevice{
				ID:    device.ID,
				Make:  dd.Make.Name,
				Model: dd.Type.Model,
				Year:  int(dd.Type.Year),
			},
			Odometer: odometer,
		},
	}
	if err := i.eventService.Emit(event); err != nil {
		i.log.Err(err).Msgf("Failed to emit odometer event for device %s", device.ID)
	}
}

func extractOdometer(data []byte) (float64, error) {
	partialData := new(struct {
		Odometer *float64 `json:"odometer"`
	})
	if err := json.Unmarshal(data, partialData); err != nil {
		return 0, fmt.Errorf("failed parsing data field: %w", err)
	}
	if partialData.Odometer == nil {
		return 0, errors.New("data payload did not have an odometer reading")
	}

	return *partialData.Odometer, nil
}

func mergeSignals(currentData map[string]interface{}, newData map[string]interface{}, t time.Time) (map[string]interface{}, error) {

	merged := make(map[string]interface{})
	for k, v := range currentData {
		merged[k] = v
	}
	// now iterate over new data and update any keys present in the new data with the events timestamp
	for k, v := range newData {
		merged[k] = map[string]interface{}{
			"timestamp": t.Format("2006-01-02T15:04:05Z"), // utc tz RFC3339
			"value":     v,
		}
	}
	return merged, nil
}

var basicVINExp = regexp.MustCompile(`^[A-Z0-9]{17}$`)

// extractVIN extracts the vin field from a status update's data object.
// If this field is not present or fails basic validation, an error is returned.
// The function does clean up the input slightly.
func extractVIN(data []byte) (string, error) {
	partialData := new(struct {
		VIN *string `json:"vin"`
	})

	if err := json.Unmarshal(data, partialData); err != nil {
		return "", fmt.Errorf("failed parsing data field: %w", err)
	}

	if partialData.VIN == nil {
		return "", errors.New("data payload did not have a VIN reading")
	}

	// Minor cleaning.
	vin := strings.ToUpper(strings.ReplaceAll(*partialData.VIN, " ", ""))

	// We have seen crazy VINs like "\u000" before.
	if !basicVINExp.MatchString(vin) {
		return "", errors.New("invalid VIN")
	}

	return vin, nil
}

func (i *DeviceStatusIngestService) getIntegrationFromEvent(event *DeviceStatusEvent) (*grpc.Integration, error) {
	for _, integration := range i.integrations {
		if strings.HasSuffix(event.Source, integration.Id) {
			return integration, nil
		}
	}
	return nil, fmt.Errorf("no matching integration found in DB for event source: %s", event.Source)
}

// here ae here
func (i *DeviceStatusIngestService) vinCredentialer(ctx goka.Context, event *DeviceStatusEvent, device *models.UserDevice) {
	if !device.VinConfirmed {
		// Nothing to compare with.
		return
	}

	storedVIN := device.VinIdentifier.String
	observedVIN, err := extractVIN(event.Data)
	if err != nil {
		// This could get noisy. Even for vehicles that do transmit VIN, it may not be in every record.
		i.log.Debug().Err(err).Str("userDeviceId", event.Subject).Msg("Couldn't extract a valid VIN from the status update.")
		return
	}

	tkn, ok := device.R.VehicleNFT.TokenID.Big.Int64()
	if !ok {
		i.log.Err(errors.New("invalid tokenID")).Str("userDeviceId", event.Subject).Msg("credential not issued")
		return
	}

	issuanceWk := currentIssuanceweek()
	issueCred := observedVIN == storedVIN

	val := ctx.Value()
	if val != nil {
		record, ok := val.(*shared.CloudEvent[RegisteredVIN])
		if !ok {
			i.log.Err(errors.New("unrecognized record format")).Str("userDeviceId", event.Subject).Msg("unable to parse record from table")
			return
		}

		if record.Data.IssuanceWeek < issuanceWk {
			if !issueCred {
				// eventually may want to revoke here
				return
			}

			_, err = i.issuer.VIN(observedVIN, big.NewInt(tkn))
			if err != nil {
				i.log.Err(err).Str("userDeviceId", event.Subject).Msg("error issuing vin credential")
			}
			record.Data.CredentialIssued = true
			record.Data.IssuanceWeek = issuanceWk
			ctx.SetValue(record)

		}

		return
	}

	record := &shared.CloudEvent[RegisteredVIN]{
		ID:      ksuid.New().String(),
		Time:    time.Now(),
		Subject: event.Subject,
		Type:    "zone.dimo.device.vin.validation",
		Data: RegisteredVIN{
			IssuanceWeek:     issuanceWk,
			CredentialIssued: issueCred,
		},
	}

	if !issueCred {
		i.log.Info().Str("userDeviceId", event.Subject).Str("observedVin", observedVIN).Str("storedVin", storedVIN).Msg("Detected potential VIN fraud.")
		return
	}

	_, err = i.issuer.VIN(observedVIN, big.NewInt(tkn))
	if err != nil {
		i.log.Err(err).Str("userDeviceId", event.Subject).Msg("error issuing vin credential")
	}

	ctx.SetValue(record)
}

func currentIssuanceweek() int {
	sinceStart := time.Now().Sub(time.Date(2022, time.January, 31, 5, 0, 0, 0, time.UTC))
	weekNum := int(sinceStart.Truncate(weekDuration) / weekDuration)
	return weekNum
}

var weekDuration = 7 * 24 * time.Hour

type odometerEventDevice struct {
	ID    string `json:"id"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
}

type OdometerEvent struct {
	Timestamp time.Time           `json:"timestamp"`
	UserID    string              `json:"userId"`
	Device    odometerEventDevice `json:"device"`
	Odometer  float64             `json:"odometer"`
}

type DeviceStatusEvent struct {
	ID          string          `json:"id"`
	Source      string          `json:"source"`
	Specversion string          `json:"specversion"`
	Subject     string          `json:"subject"`
	Time        time.Time       `json:"time"`
	Type        string          `json:"type"`
	Data        json.RawMessage `json:"data"`
}

type RegisteredVIN struct {
	IssuanceWeek     int
	CredentialIssued bool
}
