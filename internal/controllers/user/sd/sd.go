package sd

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/middleware/address"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/integration"
	"github.com/DIMO-Network/devices-api/internal/services/tmpcred"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
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
	Tesla       SyntheticTaskManager
	IntegClient *integration.Client
	Store       *tmpcred.Store
	TeslaAPI    services.TeslaFleetAPIService
	Cipher      shared.Cipher
}

type SyntheticTaskManager interface {
	StartPoll(udai *models.UserDeviceAPIIntegration, sd *models.SyntheticDevice) error
	StopPoll(udai *models.UserDeviceAPIIntegration) error
}

// PostReauthenticate godoc
// @Description Restarts a synthetic device polling job with a new set of credentials.
// @Produce json
// @Param tokenID path int true "Synthetic device token id"
// @Success 200 {object} sd.Message
// @Router /user/synthetic/device/{tokenID}/commands/reauthenticate [post]
func (co *Controller) PostReauthenticate(c *fiber.Ctx) error {
	userAddr := address.Get(c)

	tokenID, err := c.ParamsInt("tokenID")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", c.Params("tokenId")))
	}

	tx, err := co.DBS.DBS().Writer.BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	sd, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(int64(tokenID), 0))),
		qm.Load(models.SyntheticDeviceRels.VehicleToken),
		qm.Load(models.SyntheticDeviceRels.BurnRequest),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("No synthetic device with token id %d known.", tokenID))
		}
		return err
	}

	ud := sd.R.VehicleToken

	if !ud.OwnerAddress.Valid {
		return fmt.Errorf("no owner for minted vehicle?")
	}

	if common.BytesToAddress(ud.OwnerAddress.Bytes) != userAddr {
		return fiber.NewError(fiber.StatusForbidden, "Caller is not the owner of this synthetic device.")
	}

	if sd.R.BurnRequest != nil && sd.R.BurnRequest.Status != models.MetaTransactionRequestStatusFailed {
		return fiber.NewError(fiber.StatusBadRequest, "Synthetic device is in the process of being burned.")
	}

	integTokenID, _ := sd.IntegrationTokenID.Int64()

	integ, _ := co.IntegClient.ByTokenID(c.Context(), int(integTokenID))

	udai, err := models.FindUserDeviceAPIIntegration(c.Context(), tx, ud.ID, integ.ID)
	if err != nil {
		return err
	}

	// TODO(elffjs): Yeah, yeah, this is bad.
	switch integ.Vendor {
	case constants.SmartCarVendor:
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
	case constants.TeslaVendor:
		cred, err := co.Store.Retrieve(c.Context(), userAddr)
		if err != nil {
			return err
		}

		if cred.IntegrationID != int(integTokenID) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Stored credentials are for integration %d, not Tesla.", cred.IntegrationID))
		}

		teslaID, err := strconv.Atoi(udai.ExternalID.String)
		if err != nil {
			return err
		}

		// Make sure that these credentials have access to this particular vehicle.
		_, err = co.TeslaAPI.GetVehicle(c.Context(), cred.AccessToken, "na", teslaID)
		if err != nil {
			return err
		}

		encAccess, err := co.Cipher.Encrypt(cred.AccessToken)
		if err != nil {
			return err
		}
		encRefresh, err := co.Cipher.Encrypt(cred.RefreshToken)
		if err != nil {
			return err
		}
		// TODO(elffjs): Really need to clear these so that they can't be used again.
		// Refreshes will clash.

		if udai.TaskID.Valid {
			if err := co.Tesla.StopPoll(udai); err != nil {
				return err
			}
		}

		udai.AccessToken = null.StringFrom(encAccess)
		udai.RefreshToken = null.StringFrom(encRefresh)
		udai.AccessExpiresAt = null.TimeFrom(cred.Expiry)

		udai.Status = models.UserDeviceAPIIntegrationStatusPendingFirstData
		udai.TaskID = null.StringFrom(ksuid.New().String())

		cols := models.UserDeviceAPIIntegrationColumns
		_, err = udai.Update(c.Context(), co.DBS.DBS().Writer, boil.Whitelist(cols.Status, cols.TaskID, cols.AccessToken, cols.RefreshToken, cols.AccessExpiresAt, cols.UpdatedAt))
		if err != nil {
			return err
		}

		// TODO(elffjs): Stop the old one, regenerate the id. Some races here.
		if err := co.Tesla.StartPoll(udai, sd); err != nil {
			return err
		}
	default:
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Integration %d does not support reauthentication.", integTokenID))
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return c.JSON(Message{Message: "Restarted polling job."})
}

type Message struct {
	Message string `json:"message"`
}
