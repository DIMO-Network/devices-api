package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/segmentio/ksuid"

	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
)

type QueryDeviceErrorCodesReq struct {
	ErrorCodes []string `json:"errorCodes"`
}

type QueryDeviceErrorCodesResponse struct {
	Message []services.ErrorCodesResponse `json:"message"`
}

type GetUserDeviceErrorCodeQueriesResponse struct {
	Queries []GetUserDeviceErrorCodeQueriesResponseItem `json:"queries"`
}

type GetUserDeviceErrorCodeQueriesResponseItem struct {
	Codes       []string  `json:"errorCodes"`
	Description string    `json:"description"`
	RequestedAt time.Time `json:"requestedAt"`
}

func PrepareDeviceStatusInformation(ctx context.Context, ddSvc services.DeviceDefinitionService, deviceData models.UserDeviceDatumSlice, deviceDefinitionID string, deviceStyleID null.String, privilegeIDs []int64) DeviceSnapshot {
	ds := DeviceSnapshot{}

	// set the record created date to most recent one
	for _, datum := range deviceData {
		if ds.RecordCreatedAt == nil || ds.RecordCreatedAt.Unix() < datum.CreatedAt.Unix() {
			ds.RecordCreatedAt = &datum.CreatedAt
		}
	}
	// future: if time btw UpdateAt and timestamp > 7 days, ignore property

	// todo further refactor by passing in type for each option, then have switch in function below, can also refactor timestamp thing
	if slices.Contains(privilegeIDs, NonLocationData) {
		charging := findMostRecentSignal(deviceData, "charging", false)
		if charging.Exists() {
			c := charging.Get("value").Bool()
			ds.Charging = &c
		}
		fuelPercentRemaining := findMostRecentSignal(deviceData, "fuelPercentRemaining", false)
		if fuelPercentRemaining.Exists() {
			ts := fuelPercentRemaining.Get("timestamp").Time()
			if ds.RecordUpdatedAt == nil || ds.RecordUpdatedAt.Unix() < ts.Unix() {
				ds.RecordUpdatedAt = &ts
			}
			f := fuelPercentRemaining.Get("value").Float()
			if f >= 0.01 {
				ds.FuelPercentRemaining = &f
			}
		}
		batteryCapacity := findMostRecentSignal(deviceData, "batteryCapacity", false)
		if batteryCapacity.Exists() {
			b := batteryCapacity.Get("value").Int()
			ds.BatteryCapacity = &b
		}
		oilLevel := findMostRecentSignal(deviceData, "oil", false)
		if oilLevel.Exists() {
			o := oilLevel.Get("value").Float()
			ds.OilLevel = &o
		}
		stateOfCharge := findMostRecentSignal(deviceData, "soc", false)
		if stateOfCharge.Exists() {
			o := stateOfCharge.Get("value").Float()
			ds.StateOfCharge = &o
		}
		chargeLimit := findMostRecentSignal(deviceData, "chargeLimit", false)
		if chargeLimit.Exists() {
			o := chargeLimit.Get("value").Float()
			ds.ChargeLimit = &o
		}
		odometer := findMostRecentSignal(deviceData, "odometer", true)
		if odometer.Exists() {
			ts := odometer.Get("timestamp").Time()
			if ds.RecordUpdatedAt == nil || ds.RecordUpdatedAt.Unix() < ts.Unix() {
				ds.RecordUpdatedAt = &ts
			}
			o := odometer.Get("value").Float()
			ds.Odometer = &o
		}
		rangeG := findMostRecentSignal(deviceData, "range", false)
		if rangeG.Exists() {
			r := rangeG.Get("value").Float()
			ds.Range = &r
		}
		batteryVoltage := findMostRecentSignal(deviceData, "batteryVoltage", false)
		if batteryVoltage.Exists() {
			ts := batteryVoltage.Get("timestamp").Time()
			if ds.RecordUpdatedAt == nil || ds.RecordUpdatedAt.Unix() < ts.Unix() {
				ds.RecordUpdatedAt = &ts
			}
			bv := batteryVoltage.Get("value").Float()
			ds.BatteryVoltage = &bv
		}
		ambientTemp := findMostRecentSignal(deviceData, "ambientTemp", false)
		if ambientTemp.Exists() {
			at := ambientTemp.Get("value").Float()
			ds.AmbientTemp = &at
		}
		// TirePressure
		tires := findMostRecentSignal(deviceData, "tires", false)
		if tires.Exists() {
			// weird thing here is in example payloads these are all ints, but the smartcar lib has as floats
			ds.TirePressure = &smartcar.TirePressure{
				FrontLeft:  tires.Get("value").Get("frontLeft").Float(),
				FrontRight: tires.Get("value").Get("frontRight").Float(),
				BackLeft:   tires.Get("value").Get("backLeft").Float(),
				BackRight:  tires.Get("value").Get("backRight").Float(),
			}
		}
	}

	if slices.Contains(privilegeIDs, CurrentLocation) || slices.Contains(privilegeIDs, AllTimeLocation) {
		latitude := findMostRecentSignal(deviceData, "latitude", false)
		if latitude.Exists() {
			ts := latitude.Get("timestamp").Time()
			if ds.RecordUpdatedAt == nil || ds.RecordUpdatedAt.Unix() < ts.Unix() {
				ds.RecordUpdatedAt = &ts
			}
			l := latitude.Get("value").Float()
			ds.Latitude = &l
		}
		longitude := findMostRecentSignal(deviceData, "longitude", false)
		if longitude.Exists() {
			l := longitude.Get("value").Float()
			ds.Longitude = &l
		}
	}

	if ds.Range == nil && ds.FuelPercentRemaining != nil {
		calcRange, err := calculateRange(ctx, ddSvc, deviceDefinitionID, deviceStyleID, *ds.FuelPercentRemaining)
		if err == nil {
			ds.Range = calcRange
		}
	}

	return ds
}

// findMostRecentSignal finds the highest value float instead of most recent, eg. for odometer
func findMostRecentSignal(udd models.UserDeviceDatumSlice, path string, highestFloat bool) gjson.Result {
	// todo write test
	if len(udd) == 0 {
		return gjson.Result{}
	}
	if len(udd) > 1 {
		if highestFloat {
			sortBySignalValueDesc(udd, path)
		} else {
			sortBySignalTimestampDesc(udd, path)
		}
	}
	return gjson.GetBytes(udd[0].Signals.JSON, path)
}

// calculateRange returns the current estimated range based on fuel tank capacity, mpg, and fuelPercentRemaining and returns it in Kilometers
func calculateRange(ctx context.Context, ddSvc services.DeviceDefinitionService, deviceDefinitionID string, deviceStyleID null.String, fuelPercentRemaining float64) (*float64, error) {
	if fuelPercentRemaining <= 0.01 {
		return nil, errors.New("fuelPercentRemaining lt 0.01 so cannot calculate range")
	}

	dd, err := ddSvc.GetDeviceDefinitionByID(ctx, deviceDefinitionID)

	if err != nil {
		return nil, helpers.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+deviceDefinitionID)
	}

	rangeData := helpers.GetActualDeviceDefinitionMetadataValues(dd, deviceStyleID)

	// calculate, convert to Km
	if rangeData.FuelTankCapGal > 0 && rangeData.Mpg > 0 {
		fuelTankAtGal := rangeData.FuelTankCapGal * fuelPercentRemaining
		rangeMiles := rangeData.Mpg * fuelTankAtGal
		rangeKm := 1.60934 * rangeMiles
		return &rangeKm, nil
	}

	return nil, nil
}

// GetUserDeviceStatus godoc
// @Description Returns the latest status update for the device. May return 404 if the
// @Description user does not have a device with the ID, or if no status updates have come. Note this endpoint also exists under nft_controllers
// @Tags        user-devices
// @Produce     json
// @Param       user_device_id path     string true "user device ID"
// @Success     200            {object} controllers.DeviceSnapshot
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/status [get]
func (udc *UserDevicesController) GetUserDeviceStatus(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")

	userDevice, err := models.FindUserDevice(c.Context(), udc.DBS().Reader, userDeviceID)
	if err != nil {
		return err
	}

	deviceData, err := models.UserDeviceData(
		models.UserDeviceDatumWhere.UserDeviceID.EQ(userDevice.ID),
		models.UserDeviceDatumWhere.Signals.IsNotNull(),
		models.UserDeviceDatumWhere.UpdatedAt.GT(time.Now().Add(-14*24*time.Hour)),
	).All(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}

	if len(deviceData) == 0 {
		return fiber.NewError(fiber.StatusNotFound, "No status updates yet.")
	}

	ds := PrepareDeviceStatusInformation(c.Context(), udc.DeviceDefSvc, deviceData, userDevice.DeviceDefinitionID,
		userDevice.DeviceStyleID, []int64{NonLocationData, CurrentLocation, AllTimeLocation})

	return c.JSON(ds)
}

// RefreshUserDeviceStatus godoc
// @Description Starts the process of refreshing device status from Smartcar
// @Tags        user-devices
// @Param       user_device_id path string true "user device ID"
// @Success     204
// @Failure     429 "rate limit hit for integration"
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/commands/refresh [post]
func (udc *UserDevicesController) RefreshUserDeviceStatus(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)
	// We could probably do a smarter join here, but it's unclear to me how to handle that
	// in SQLBoiler.
	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceData),
		qm.Load(qm.Rels(models.UserDeviceRels.UserDeviceData)),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return err
	}
	smartCarInteg, err := udc.DeviceDefSvc.GetIntegrationByVendor(c.Context(), constants.SmartCarVendor)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get smartcar integration")
	}

	for _, deviceDatum := range ud.R.UserDeviceData {
		if deviceDatum.IntegrationID == smartCarInteg.Id {
			nextAvailableTime := deviceDatum.UpdatedAt.Add(time.Second * time.Duration(smartCarInteg.RefreshLimitSecs))
			if time.Now().Before(nextAvailableTime) {
				return fiber.NewError(fiber.StatusTooManyRequests, "rate limit for integration refresh hit")
			}

			udai, err := models.FindUserDeviceAPIIntegration(c.Context(), udc.DBS().Reader, deviceDatum.UserDeviceID, deviceDatum.IntegrationID)
			if err != nil {
				return err
			}
			if udai.Status == models.UserDeviceAPIIntegrationStatusActive && udai.TaskID.Valid {
				err = udc.smartcarTaskSvc.Refresh(udai)
				if err != nil {
					return err
				}
				return c.SendStatus(204)
			}

			return fiber.NewError(fiber.StatusConflict, "Integration not active.")
		}
	}
	return fiber.NewError(fiber.StatusBadRequest, "no active Smartcar integration found for this device")
}

var errorCodeRegex = regexp.MustCompile(`^.{5,8}$`)

// QueryDeviceErrorCodes godoc
// @Description Queries chatgpt for user device error codes
// @Tags        user-devices
// @Param       user_device_id path string true "user device ID"
// @Param       queryDeviceErrorCodes body controllers.QueryDeviceErrorCodesReq true "error codes"
// @Success     200 {object} controllers.QueryDeviceErrorCodesResponse
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/error-codes [post]
func (udc *UserDevicesController) QueryDeviceErrorCodes(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")

	logger := helpers.GetLogger(c, udc.log)
	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id found.")
		}
		return err
	}

	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), ud.DeviceDefinitionID)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+ud.DeviceDefinitionID)
	}

	req := &QueryDeviceErrorCodesReq{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request.")
	}

	errorCodesLimit := 100
	if len(req.ErrorCodes) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "No error codes provided")
	}
	if len(req.ErrorCodes) > errorCodesLimit {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Too many error codes. Error codes list must be %d or below in length.", errorCodesLimit))
	}

	for _, v := range req.ErrorCodes {
		if !errorCodeRegex.MatchString(v) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Invalid error code %s", v))
		}
	}

	appmetrics.OpenAITotalCallsOps.Inc() // record new total call to chatgpt
	chtResp, err := udc.openAI.GetErrorCodesDescription(dd.Type.Make, dd.Type.Model, req.ErrorCodes)
	if err != nil {
		appmetrics.OpenAITotalFailedCallsOps.Inc()
		logger.Err(err).Interface("requestBody", req).Msg("Error occurred fetching description for error codes")
		return err
	}

	chtJSON, err := json.Marshal(chtResp)
	if err != nil {
		logger.Err(err).Interface("requestBody", req).Msg("Error occurred fetching description for error codes")
		return fiber.NewError(fiber.StatusInternalServerError, "Error occurred fetching description for error codes")
	}

	q := &models.ErrorCodeQuery{ID: ksuid.New().String(), UserDeviceID: udi, ErrorCodes: req.ErrorCodes, CodesQueryResponse: null.JSONFrom(chtJSON)}
	err = q.Insert(c.Context(), udc.DBS().Writer, boil.Infer())

	if err != nil {
		// TODO - should we return an error for this or just log it
		logger.Err(err).Msg("Could not save user query response")
	}

	return c.JSON(&QueryDeviceErrorCodesResponse{
		Message: chtResp,
	})
}

// GetUserDeviceErrorCodeQueries godoc
// @Description Returns all error codes queries for user devices
// @Tags        user-devices
// @Success     200 {object} controllers.GetUserDeviceErrorCodeQueriesResponse
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/error-codes [get]
func (udc *UserDevicesController) GetUserDeviceErrorCodeQueries(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	logger := helpers.GetLogger(c, udc.log)

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.ErrorCodeQueries, qm.OrderBy(models.ErrorCodeQueryColumns.CreatedAt+" DESC")),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Could not find user device")
		}
		logger.Err(err).Msg("error occurred when fetching error codes for device")
		return fiber.NewError(fiber.StatusInternalServerError, "error occurred fetching device error queries")
	}

	queries := []GetUserDeviceErrorCodeQueriesResponseItem{}

	for _, erc := range userDevice.R.ErrorCodeQueries {
		queries = append(queries, GetUserDeviceErrorCodeQueriesResponseItem{
			Codes:       erc.ErrorCodes,
			Description: erc.QueryResponse,
			RequestedAt: erc.CreatedAt,
		})
	}

	return c.JSON(GetUserDeviceErrorCodeQueriesResponse{Queries: queries})
}

// DeviceSnapshot is the response object for device status endpoint
// https://docs.google.com/document/d/1DYzzTOR9WA6WJNoBnwpKOoxfmrVwPWNLv0x0MkjIAqY/edit#heading=h.dnp7xngl47bw
type DeviceSnapshot struct {
	Charging             *bool                  `json:"charging,omitempty"`
	FuelPercentRemaining *float64               `json:"fuelPercentRemaining,omitempty"`
	BatteryCapacity      *int64                 `json:"batteryCapacity,omitempty"`
	OilLevel             *float64               `json:"oil,omitempty"`
	Odometer             *float64               `json:"odometer,omitempty"`
	Latitude             *float64               `json:"latitude,omitempty"`
	Longitude            *float64               `json:"longitude,omitempty"`
	Range                *float64               `json:"range,omitempty"`
	StateOfCharge        *float64               `json:"soc,omitempty"`
	ChargeLimit          *float64               `json:"chargeLimit,omitempty"`
	RecordUpdatedAt      *time.Time             `json:"recordUpdatedAt,omitempty"`
	RecordCreatedAt      *time.Time             `json:"recordCreatedAt,omitempty"`
	TirePressure         *smartcar.TirePressure `json:"tirePressure,omitempty"`
	BatteryVoltage       *float64               `json:"batteryVoltage,omitempty"`
	AmbientTemp          *float64               `json:"ambientTemp,omitempty"`
}

// sortBySignalValueDesc Sort user device data so the highest value is first
func sortBySignalValueDesc(udd models.UserDeviceDatumSlice, path string) {
	sort.Slice(udd, func(i, j int) bool {
		fpri := gjson.GetBytes(udd[i].Signals.JSON, path+".value")
		fprj := gjson.GetBytes(udd[j].Signals.JSON, path+".value")
		// if one has it and the other does not, makes no difference
		if fpri.Exists() && !fprj.Exists() {
			return true
		} else if !fpri.Exists() && fprj.Exists() {
			return false
		}
		return fprj.Float() < fpri.Float()
	})
}

// sortBySignalTimestampDesc Sort user device data so the most recent timestamp is first
func sortBySignalTimestampDesc(udd models.UserDeviceDatumSlice, path string) {
	sort.Slice(udd, func(i, j int) bool {
		fpri := gjson.GetBytes(udd[i].Signals.JSON, path+".timestamp")
		fprj := gjson.GetBytes(udd[j].Signals.JSON, path+".timestamp")
		// if one has it and the other does not, makes no difference
		if fpri.Exists() && !fprj.Exists() {
			return true
		} else if !fpri.Exists() && fprj.Exists() {
			return false
		}
		return fprj.Time().Unix() < fpri.Time().Unix()
	})
}
