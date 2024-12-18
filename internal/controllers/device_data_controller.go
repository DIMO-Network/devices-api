package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/DIMO-Network/shared"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/segmentio/ksuid"

	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type QueryDeviceErrorCodesReq struct {
	ErrorCodes []string `json:"errorCodes" example:"P0106,P0279"`
}

type QueryDeviceErrorCodesResponse struct {
	ErrorCodes []services.ErrorCodesResponse `json:"errorCodes"`
	ClearedAt  *time.Time                    `json:"clearedAt"`
}

type GetUserDeviceErrorCodeQueriesResponse struct {
	Queries []GetUserDeviceErrorCodeQueriesResponseItem `json:"queries"`
}

type GetUserDeviceErrorCodeQueriesResponseItem struct {
	ErrorCodes  []services.ErrorCodesResponse `json:"errorCodes"`
	RequestedAt time.Time                     `json:"requestedAt" example:"2023-05-23T12:56:36Z"`
	// ClearedAt is the time at which the user cleared the codes from this query.
	// May be null.
	ClearedAt *time.Time `json:"clearedAt" example:"2023-05-23T12:57:05Z"`
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

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return err
	}

	smartCarInteg, err := udc.DeviceDefSvc.GetIntegrationByVendor(c.Context(), constants.SmartCarVendor)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "failed to get smartcar integration")
	}

	deviceData, err := udc.deviceDataSvc.GetRawDeviceData(c.Context(), ud.ID, smartCarInteg.Id)
	if err != nil {
		return errors.Wrapf(err, "failed to get device data to do smartcar refresh")
	}

	for _, deviceDatum := range deviceData.Items {
		if deviceDatum.IntegrationId == smartCarInteg.Id {
			nextAvailableTime := deviceDatum.RecordUpdatedAt.AsTime().UTC().Add(time.Second * time.Duration(smartCarInteg.RefreshLimitSecs))
			if time.Now().UTC().Before(nextAvailableTime) {
				helpers.SkipErrorLog(c)
				return fiber.NewError(fiber.StatusTooManyRequests,
					fmt.Sprintf("rate limit for integration refresh hit, next available: %s", nextAvailableTime.Format(time.RFC3339)))
			}

			udai, err := models.FindUserDeviceAPIIntegration(c.Context(), udc.DBS().Reader, ud.ID, smartCarInteg.Id)
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
			helpers.SkipErrorLog(c)
			return fiber.NewError(fiber.StatusConflict, "Integration not active.")
		}
	}
	return fiber.NewError(fiber.StatusBadRequest, "no active Smartcar integration found for this device")
}

var errorCodeRegex = regexp.MustCompile(`^.{5,8}$`)

// QueryDeviceErrorCodes godoc
// @Summary     Obtain, store, and return descriptions for a list of error codes from this vehicle.
// @Description Deprecated. Use `/user/vehicle/{tokenID}/error-codes` instead
// @Tags        error-codes
// @Param       userDeviceID path string true "user device id"
// @Param       queryDeviceErrorCodes body controllers.QueryDeviceErrorCodesReq true "error codes"
// @Success     200 {object} controllers.QueryDeviceErrorCodesResponse
// @Failure     404 {object} helpers.ErrorRes "Vehicle not found"
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/error-codes [post]
// Deprecated
func (udc *UserDevicesController) QueryDeviceErrorCodes(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")

	logger := helpers.GetLogger(c, udc.log)
	ud, err := models.UserDevices(models.UserDeviceWhere.ID.EQ(udi)).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id found.")
		}
		return err
	}

	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionBySlug(c.Context(), ud.DefinitionID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+ud.DefinitionID)
	}

	req := &QueryDeviceErrorCodesReq{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request.")
	}

	errorCodesLimit := 100
	if len(req.ErrorCodes) > errorCodesLimit {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Too many error codes. Error codes list must be %d or below in length.", errorCodesLimit))
	}

	errorCodesCleaned := make([]string, 0, len(req.ErrorCodes))

	for _, v := range req.ErrorCodes {
		if v == "" {
			logger.Warn().Msg("Client sent an empty error code.")
			// The app is sending in a lot of these.
			continue
		}
		if !errorCodeRegex.MatchString(v) {
			logger.Error().Msgf("Got a weird error code list %v.", req.ErrorCodes)
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Invalid error code %q.", v))
		}
		errorCodesCleaned = append(errorCodesCleaned, v)
	}

	if len(errorCodesCleaned) == 0 {
		logger.Warn().Msg("Client sent an empty list.")
		return c.JSON(&QueryDeviceErrorCodesResponse{
			ErrorCodes: []services.ErrorCodesResponse{},
		})
	}

	appmetrics.OpenAITotalCallsOps.Inc() // record new total call to chatgpt
	chtResp, err := udc.openAI.GetErrorCodesDescription(dd.Make.Name, dd.Model, errorCodesCleaned)
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

	q := &models.ErrorCodeQuery{ID: ksuid.New().String(), UserDeviceID: udi, CodesQueryResponse: null.JSONFrom(chtJSON)}
	err = q.Insert(c.Context(), udc.DBS().Writer, boil.Infer())

	if err != nil {
		// TODO - should we return an error for this or just log it
		logger.Err(err).Msg("Could not save user query response")
	}

	return c.JSON(&QueryDeviceErrorCodesResponse{
		ErrorCodes: chtResp,
	})
}

// GetUserDeviceErrorCodeQueries godoc
// @Summary List all error code queries made for this vehicle.
// @Description Deprecated. Use `/user/vehicle/{tokenID}/error-codes` instead
// @Tags        error-codes
// @Param       userDeviceID path string true "user device id"
// @Success     200 {object} controllers.GetUserDeviceErrorCodeQueriesResponse
// @Failure     404 {object} helpers.ErrorRes "Vehicle not found"
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/error-codes [get]
// Deprecated
func (udc *UserDevicesController) GetUserDeviceErrorCodeQueries(c *fiber.Ctx) error {
	logger := helpers.GetLogger(c, udc.log)

	userDeviceID := c.Params("userDeviceID")

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
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
		ercJSON := []services.ErrorCodesResponse{}
		if err := erc.CodesQueryResponse.Unmarshal(&ercJSON); err != nil {
			return err
		}

		userDeviceresp := GetUserDeviceErrorCodeQueriesResponseItem{
			ErrorCodes:  ercJSON,
			RequestedAt: erc.CreatedAt,
			ClearedAt:   erc.ClearedAt.Ptr(),
		}

		queries = append(queries, userDeviceresp)
	}

	return c.JSON(GetUserDeviceErrorCodeQueriesResponse{Queries: queries})
}

// ClearUserDeviceErrorCodeQuery godoc
// @Summary     Mark the most recent set of error codes as having been cleared.
// @Description Deprecated. Use `/user/vehicle/{tokenID}/error-codes/clear` instead
// @Tags        error-codes
// @Success     200 {object} controllers.QueryDeviceErrorCodesResponse
// @Failure     429 {object} helpers.ErrorRes "Last query already cleared"
// @Failure     404 {object} helpers.ErrorRes "Vehicle not found"
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/error-codes/clear [post]
// Deprecated
func (udc *UserDevicesController) ClearUserDeviceErrorCodeQuery(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")

	logger := helpers.GetLogger(c, udc.log)

	errCodeQuery, err := models.ErrorCodeQueries(
		models.ErrorCodeQueryWhere.UserDeviceID.EQ(udi),
		qm.OrderBy(models.ErrorCodeQueryColumns.CreatedAt+" DESC"),
		qm.Limit(1),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		logger.Err(err).Msg("error occurred when fetching error codes for device")
		return fiber.NewError(fiber.StatusBadRequest, "error occurred fetching device error queries")
	}

	if errCodeQuery.ClearedAt.Valid {
		return fiber.NewError(fiber.StatusBadRequest, "all error codes already cleared")
	}

	errCodeQuery.ClearedAt = null.TimeFrom(time.Now().UTC().Truncate(time.Microsecond))
	if _, err = errCodeQuery.Update(c.Context(), udc.DBS().Writer, boil.Whitelist(models.ErrorCodeQueryColumns.ClearedAt)); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error occurred updating device error queries")
	}

	errorCodeResp := []services.ErrorCodesResponse{}
	if err := errCodeQuery.CodesQueryResponse.Unmarshal(&errorCodeResp); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error occurred updating device error queries")
	}

	return c.JSON(&QueryDeviceErrorCodesResponse{
		ErrorCodes: errorCodeResp,
		ClearedAt:  &errCodeQuery.ClearedAt.Time,
	})
}

// GetUserDeviceErrorCodeQueriesByTokenID godoc
// @Summary List all error code queries made for this vehicle.
// @Tags        error-codes
// @Param       tokenID path int true "vehicle token id"
// @Success     200 {object} controllers.GetUserDeviceErrorCodeQueriesResponse
// @Failure     404 {object} helpers.ErrorRes "Vehicle not found"
// @Security    BearerAuth
// @Router      /user/vehicle/{tokenID}/error-codes [get]
func (udc *UserDevicesController) GetUserDeviceErrorCodeQueriesByTokenID(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))
	logger := helpers.GetLogger(c, udc.log)

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(tid),
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
		var ercJSON []services.ErrorCodesResponse
		if err := erc.CodesQueryResponse.Unmarshal(&ercJSON); err != nil {
			return err
		}

		userDeviceresp := GetUserDeviceErrorCodeQueriesResponseItem{
			ErrorCodes:  ercJSON,
			RequestedAt: erc.CreatedAt,
			ClearedAt:   erc.ClearedAt.Ptr(),
		}

		queries = append(queries, userDeviceresp)
	}

	return c.JSON(GetUserDeviceErrorCodeQueriesResponse{Queries: queries})
}

// QueryDeviceErrorCodesByTokenID godoc
// @Summary     Obtain, store, and return descriptions for a list of error codes from this vehicle.
// @Tags        error-codes
// @Param       tokenID path int true "vehicle token id"
// @Param       queryDeviceErrorCodes body controllers.QueryDeviceErrorCodesReq true "error codes"
// @Success     200 {object} controllers.QueryDeviceErrorCodesResponse
// @Failure     404 {object} helpers.ErrorRes "Vehicle not found"
// @Security    BearerAuth
// @Router      /user/vehicle/{tokenID}/error-codes [post]
func (udc *UserDevicesController) QueryDeviceErrorCodesByTokenID(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	logger := helpers.GetLogger(c, udc.log)
	ud, err := models.UserDevices(models.UserDeviceWhere.TokenID.EQ(tid)).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id found.")
		}
		return err
	}

	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionBySlug(c.Context(), ud.DefinitionID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+ud.DefinitionID)
	}

	req := &QueryDeviceErrorCodesReq{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request.")
	}

	errorCodesLimit := 100
	if len(req.ErrorCodes) > errorCodesLimit {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Too many error codes. Error codes list must be %d or below in length.", errorCodesLimit))
	}

	errorCodesCleaned := make([]string, 0, len(req.ErrorCodes))

	for _, v := range req.ErrorCodes {
		if v == "" {
			logger.Warn().Msg("Client sent an empty error code.")
			// The app is sending in a lot of these.
			continue
		}
		if !errorCodeRegex.MatchString(v) {
			logger.Error().Msgf("Got a weird error code list %v.", req.ErrorCodes)
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Invalid error code %q.", v))
		}
		errorCodesCleaned = append(errorCodesCleaned, v)
	}

	if len(errorCodesCleaned) == 0 {
		logger.Warn().Msg("Client sent an empty list.")
		return c.JSON(&QueryDeviceErrorCodesResponse{
			ErrorCodes: []services.ErrorCodesResponse{},
		})
	}

	appmetrics.OpenAITotalCallsOps.Inc() // record new total call to chatgpt
	chtResp, err := udc.openAI.GetErrorCodesDescription(dd.Make.Name, dd.Model, errorCodesCleaned)
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

	q := &models.ErrorCodeQuery{ID: ksuid.New().String(), UserDeviceID: ud.ID, VehicleTokenID: ud.TokenID, CodesQueryResponse: null.JSONFrom(chtJSON)}
	err = q.Insert(c.Context(), udc.DBS().Writer, boil.Infer())

	if err != nil {
		// TODO - should we return an error for this or just log it
		logger.Err(err).Msg("Could not save user query response")
	}

	return c.JSON(&QueryDeviceErrorCodesResponse{
		ErrorCodes: chtResp,
	})
}

// ClearUserDeviceErrorCodeQueryByTokenID godoc
// @Summary     Mark the most recent set of error codes as having been cleared.
// @Tags        error-codes
// @Param       tokenID path int true "vehicle token id"
// @Success     200 {object} controllers.QueryDeviceErrorCodesResponse
// @Failure     429 {object} helpers.ErrorRes "Last query already cleared"
// @Failure     404 {object} helpers.ErrorRes "Vehicle not found"
// @Security    BearerAuth
// @Router      /user/vehicle/{tokenID}/error-codes/clear [post]
func (udc *UserDevicesController) ClearUserDeviceErrorCodeQueryByTokenID(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	logger := helpers.GetLogger(c, udc.log)

	errCodeQuery, err := models.ErrorCodeQueries(
		models.ErrorCodeQueryWhere.VehicleTokenID.EQ(tid),
		qm.OrderBy(models.ErrorCodeQueryColumns.CreatedAt+" DESC"),
		qm.Limit(1),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Could not find user device")
		}
		logger.Err(err).Msg("error occurred when fetching error codes for device")
		return fiber.NewError(fiber.StatusBadRequest, "error occurred fetching device error queries")
	}

	if errCodeQuery.ClearedAt.Valid {
		return fiber.NewError(fiber.StatusTooManyRequests, "all error codes already cleared")
	}

	errCodeQuery.ClearedAt = null.TimeFrom(time.Now().UTC().Truncate(time.Microsecond))
	if _, err = errCodeQuery.Update(c.Context(), udc.DBS().Writer, boil.Whitelist(models.ErrorCodeQueryColumns.ClearedAt)); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error occurred updating device error queries")
	}

	var errorCodeResp []services.ErrorCodesResponse
	if err := errCodeQuery.CodesQueryResponse.Unmarshal(&errorCodeResp); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error occurred updating device error queries")
	}

	return c.JSON(&QueryDeviceErrorCodesResponse{
		ErrorCodes: errorCodeResp,
		ClearedAt:  &errCodeQuery.ClearedAt.Time,
	})
}
