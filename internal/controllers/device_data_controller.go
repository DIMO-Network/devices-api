package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/DIMO-Network/devices-api/internal/services"

	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
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
	Message string `json:"message"`
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
	// order the records by odometer asc so that if they both have it, the latter one replaces with more recent values.
	sortByJSONOdometerAsc(deviceData)

	// merging data: foreach order by updatedAt desc, only set property if it exists in json data
	for _, datum := range deviceData {
		if datum.Data.Valid {
			// The response has the creation and update times of the most recently updated integration.
			// For users with, e.g., both Smartcar and AutoPi this may produce confusing results.
			ds.RecordCreatedAt = &datum.CreatedAt
			ds.RecordUpdatedAt = &datum.UpdatedAt

			// note we are assuming json property names are same accross smartcar, tesla, autopi, AND same types eg. int / float / string
			// we could use reflection and just have single line assuming json name in struct matches what is in data
			if slices.Contains(privilegeIDs, NonLocationData) {
				charging := gjson.GetBytes(datum.Data.JSON, "charging")
				if charging.Exists() {
					c := charging.Bool()
					ds.Charging = &c
				}
				fuelPercentRemaining := gjson.GetBytes(datum.Data.JSON, "fuelPercentRemaining")
				if fuelPercentRemaining.Exists() {
					f := fuelPercentRemaining.Float()
					if f >= 0.01 {
						ds.FuelPercentRemaining = &f
					}
				}
				batteryCapacity := gjson.GetBytes(datum.Data.JSON, "batteryCapacity")
				if batteryCapacity.Exists() {
					b := batteryCapacity.Int()
					ds.BatteryCapacity = &b
				}
				oilLevel := gjson.GetBytes(datum.Data.JSON, "oil")
				if oilLevel.Exists() {
					o := oilLevel.Float()
					ds.OilLevel = &o
				}
				stateOfCharge := gjson.GetBytes(datum.Data.JSON, "soc")
				if stateOfCharge.Exists() {
					o := stateOfCharge.Float()
					ds.StateOfCharge = &o
				}
				chargeLimit := gjson.GetBytes(datum.Data.JSON, "chargeLimit")
				if chargeLimit.Exists() {
					o := chargeLimit.Float()
					ds.ChargeLimit = &o
				}
				odometer := gjson.GetBytes(datum.Data.JSON, "odometer")
				if odometer.Exists() {
					o := odometer.Float()
					ds.Odometer = &o
				}
				rangeG := gjson.GetBytes(datum.Data.JSON, "range")
				if rangeG.Exists() {
					r := rangeG.Float()
					ds.Range = &r
				}
				batteryVoltage := gjson.GetBytes(datum.Data.JSON, "batteryVoltage")
				if batteryVoltage.Exists() {
					bv := batteryVoltage.Float()
					ds.BatteryVoltage = &bv
				}
				ambientTemp := gjson.GetBytes(datum.Data.JSON, "ambientTemp")
				if ambientTemp.Exists() {
					at := ambientTemp.Float()
					ds.AmbientTemp = &at
				}
				// TirePressure
				tires := gjson.GetBytes(datum.Data.JSON, "tires")
				if tires.Exists() {
					// weird thing here is in example payloads these are all ints, but the smartcar lib has as floats
					ds.TirePressure = &smartcar.TirePressure{
						FrontLeft:  tires.Get("frontLeft").Float(),
						FrontRight: tires.Get("frontRight").Float(),
						BackLeft:   tires.Get("backLeft").Float(),
						BackRight:  tires.Get("backRight").Float(),
					}
				}

			}
			if slices.Contains(privilegeIDs, CurrentLocation) || slices.Contains(privilegeIDs, AllTimeLocation) {
				latitude := gjson.GetBytes(datum.Data.JSON, "latitude")
				if latitude.Exists() {
					l := latitude.Float()
					ds.Latitude = &l
				}
				longitude := gjson.GetBytes(datum.Data.JSON, "longitude")
				if longitude.Exists() {
					l := longitude.Float()
					ds.Longitude = &l
				}
			}

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
// @Description user does not have a device with the ID, or if no status updates have come
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
		qm.OrderBy(models.UserDeviceDatumColumns.UpdatedAt),
	).All(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}

	if len(deviceData) == 0 || !deviceData[0].Data.Valid {
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

	q := &models.ErrorCodeQuery{ID: ksuid.New().String(), UserDeviceID: udi, ErrorCodes: req.ErrorCodes, QueryResponse: chtResp}
	err = q.Insert(c.Context(), udc.DBS().Writer, boil.Infer())

	if err != nil {
		// TODO - should we return an error for this or just log it
		logger.Err(err).Msg("Could not save user query response")
	}

	return c.JSON(&QueryDeviceErrorCodesResponse{
		Message: chtResp,
	})
}

// GetUserDevicesErrorCodeQueries godoc
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
	StateOfCharge        *float64               `json:"soc,omitempty"` // todo: change json to match after update frontend
	ChargeLimit          *float64               `json:"chargeLimit,omitempty"`
	RecordUpdatedAt      *time.Time             `json:"recordUpdatedAt,omitempty"`
	RecordCreatedAt      *time.Time             `json:"recordCreatedAt,omitempty"`
	TirePressure         *smartcar.TirePressure `json:"tirePressure,omitempty"`
	BatteryVoltage       *float64               `json:"batteryVoltage,omitempty"`
	AmbientTemp          *float64               `json:"ambientTemp,omitempty"`
}

// sortByJSONOdometerAsc Sort user device data so the highest odometer is last
func sortByJSONOdometerAsc(udd models.UserDeviceDatumSlice) {
	sort.Slice(udd, func(i, j int) bool {
		fpri := gjson.GetBytes(udd[i].Data.JSON, "odometer")
		fprj := gjson.GetBytes(udd[j].Data.JSON, "odometer")
		// if one has it and the other does not, makes no difference
		if fpri.Exists() && !fprj.Exists() {
			return true
		} else if !fpri.Exists() && fprj.Exists() {
			return false
		}
		return fprj.Float() > fpri.Float()
	})
}
