package controllers

import (
	"database/sql"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
)

type UserDevicesControllerV2 struct {
	Settings     *config.Settings
	DBS          func() *db.ReaderWriter
	log          *zerolog.Logger
	DeviceDefSvc services.DeviceDefinitionService
}

func NewUserDevicesControllerV2(settings *config.Settings, dbs func() *db.ReaderWriter, logger *zerolog.Logger, deviceDefSvc services.DeviceDefinitionService) UserDevicesControllerV2 {
	return UserDevicesControllerV2{
		Settings:     settings,
		DBS:          dbs,
		log:          logger,
		DeviceDefSvc: deviceDefSvc,
	}
}

type DeviceRange struct {
	// Contains a list of range sets, one for each range basis (may be empty)
	RangeSets []RangeSet `json:"rangeSets"`
}

type RangeSet struct {
	// The time the data was collected
	Updated string `json:"updated"`
	// The basis for the range calculation (eg. "MPG" or "MPG Highway")
	RangeBasis string `json:"rangeBasis"`
	// The estimated range distance
	RangeDistance int `json:"rangeDistance"`
	// The unit used for the rangeDistance (eg. "miles" or "kilometers")
	RangeUnit string `json:"rangeUnit"`
}

// GetRange godoc
// @Description gets the estimated range for a particular user device
// @Tags        user-devices
// @Produce     json
// @Success     200 {object} controllers.DeviceRange
// @Security    BearerAuth
// @Param       tokenId path int true "tokenId"
// @Router      /v2/vehicles/{tokenId}/analytics/range [get]
func (udc *UserDevicesControllerV2) GetRange(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}

	tkID := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(tkID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that token ID found.")
		}
		udc.log.Err(err).Str("TokenID", tkID.String()).Msg("could not get device")
		return fiber.NewError(fiber.StatusInternalServerError, "error occurred getting device with token id")
	}

	if !nft.UserDeviceID.Valid {
		return fiber.NewError(fiber.StatusInternalServerError, "error occurred getting device with token id")
	}

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(nft.UserDeviceID.String),
		qm.Load(models.UserDeviceRels.UserDeviceData),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}

	dds, err := udc.DeviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), []string{userDevice.DeviceDefinitionID})
	if err != nil {
		return shared.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+userDevice.DeviceDefinitionID)
	}

	deviceRange := DeviceRange{
		RangeSets: []RangeSet{},
	}
	udd := userDevice.R.UserDeviceData
	if len(dds) > 0 && dds[0] != nil && len(udd) > 0 {

		rangeData := helpers.GetActualDeviceDefinitionMetadataValues(dds[0], userDevice.DeviceStyleID)

		sortByJSONFieldMostRecent(udd, "fuelPercentRemaining")
		fuelPercentRemaining := gjson.GetBytes(udd[0].Signals.JSON, "fuelPercentRemaining.value")
		dataUpdatedOn := gjson.GetBytes(udd[0].Signals.JSON, "fuelPercentRemaining.timestamp").Time()
		if fuelPercentRemaining.Exists() && rangeData.FuelTankCapGal > 0 && rangeData.Mpg > 0 {
			fuelTankAtGal := rangeData.FuelTankCapGal * fuelPercentRemaining.Float()
			rangeSet := RangeSet{
				Updated:       dataUpdatedOn.Format(time.RFC3339),
				RangeBasis:    "MPG",
				RangeDistance: int(rangeData.Mpg * fuelTankAtGal),
				RangeUnit:     "miles",
			}
			deviceRange.RangeSets = append(deviceRange.RangeSets, rangeSet)
			if rangeData.MpgHwy > 0 {
				rangeSet.RangeBasis = "MPG Highway"
				rangeSet.RangeDistance = int(rangeData.MpgHwy * fuelTankAtGal)
				deviceRange.RangeSets = append(deviceRange.RangeSets, rangeSet)
			}
		}
		sortByJSONFieldMostRecent(udd, "range")
		reportedRange := gjson.GetBytes(udd[0].Signals.JSON, "range.value")
		dataUpdatedOn = gjson.GetBytes(udd[0].Signals.JSON, "range.timestamp").Time()
		if reportedRange.Exists() {
			reportedRangeMiles := int(reportedRange.Float() / services.MilesToKmFactor)
			rangeSet := RangeSet{
				Updated:       dataUpdatedOn.Format(time.RFC3339),
				RangeBasis:    "Vehicle Reported",
				RangeDistance: reportedRangeMiles,
				RangeUnit:     "miles",
			}
			deviceRange.RangeSets = append(deviceRange.RangeSets, rangeSet)
		}
	}

	return c.JSON(deviceRange)
}

// sortByJSONFieldMostRecent Sort user device data so the latest that has the specified field is first
// only pass in field name, as this will append "timestamp" to look compare signals.field.timestamp
func sortByJSONFieldMostRecent(udd models.UserDeviceDatumSlice, field string) {
	sort.Slice(udd, func(i, j int) bool {
		fpri := gjson.GetBytes(udd[i].Signals.JSON, field+".timestamp")
		fprj := gjson.GetBytes(udd[j].Signals.JSON, field+".timestamp")
		if fpri.Exists() && !fprj.Exists() {
			return true
		} else if !fpri.Exists() && fprj.Exists() {
			return false
		}
		return fpri.Time().After(fprj.Time())
	})
}
