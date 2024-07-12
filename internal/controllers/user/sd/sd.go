package sd

import (
	"fmt"

	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/middleware/address"
	"github.com/DIMO-Network/devices-api/internal/services/integration"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type Controller struct {
	DBS         db.Store
	Smartcar    SyntheticTaskManager
	IntegClient *integration.Client
}

type SyntheticTaskManager interface {
	StartPoll(udai *models.UserDeviceAPIIntegration, sd *models.SyntheticDevice) error
}

// PostReauthenticate godoc
// @Description Restarts a synthetic device polling job with a new set of credentials.
// @Produce json
// @Param tokenID path int true "Synthetic device token id"
// @Success 200 {object} sd.Message
// @Router /user/synthetic/device/{tokenID}/reauthenticate [post]
func (co *Controller) PostReauthenticate(c *fiber.Ctx) error {
	userAddr := address.Get(c)
	logger := helpers.GetLogger(c, nil)

	tokenID, err := c.ParamsInt("tokenID")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", c.Params("tokenId")))
	}

	sd, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(int64(tokenID), 0))),
		qm.Load(models.SyntheticDeviceRels.VehicleToken),
	).One(c.Context(), co.DBS.DBS().Reader)
	if err != nil {
		return err
	}

	ud := sd.R.VehicleToken

	if !ud.OwnerAddress.Valid {
		return fmt.Errorf("no owner for minted vehicle?")
	}

	if common.BytesToAddress(ud.OwnerAddress.Bytes) != userAddr {
		return fiber.NewError(fiber.StatusForbidden, "Caller is not the owner of this synthetic device.")
	}

	i, _ := sd.IntegrationTokenID.Int64()

	integ, _ := co.IntegClient.ByTokenID(c.Context(), int(i))
	if integ.Vendor != constants.SmartCarVendor {
		// TODO(elffjs): This is super dumb.
		return fiber.NewError(fiber.StatusBadRequest, "Reauthentication only supported for Smartcar.")
	}

	udai, err := models.FindUserDeviceAPIIntegration(c.Context(), co.DBS.DBS().Reader, ud.ID, integ.ID)
	if err != nil {
		return err
	}

	if udai.Status != models.UserDeviceAPIIntegrationStatusAuthenticationFailure {
		// TODO(elffjs): Can probably still "succeed" in this case.
		return fiber.NewError(fiber.StatusBadRequest, "Device is not in authentication failure.")
	}

	udai.Status = models.UserDeviceAPIIntegrationStatusPendingFirstData
	udai.TaskID = null.StringFrom(ksuid.New().String())

	cols := models.UserDeviceAPIIntegrationColumns
	_, err = udai.Update(c.Context(), co.DBS.DBS().Writer, boil.Whitelist(cols.Status, cols.TaskID, cols.UpdatedAt))
	if err != nil {
		return err
	}

	err = co.Smartcar.StartPoll(udai, sd)
	if err != nil {
		return err
	}

	logger.Info().Int("sdId", tokenID).Msg("Restarted polling job.")

	return c.JSON(Message{Message: "Restarted polling job."})
}

type Message struct {
	Message string `json:"message"`
}
