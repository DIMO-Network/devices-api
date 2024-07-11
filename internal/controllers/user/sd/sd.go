package sd

import (
	"fmt"

	"github.com/DIMO-Network/devices-api/internal/middleware/address"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type Controller struct {
	dbs db.Store
}

type IntegrationIDMapper interface {
	IDFor(tokenID int) (string, error)
}

func (co *Controller) PostReauthenticate(c *fiber.Ctx) error {
	var mapper IntegrationIDMapper

	userAddr := address.Get(c)

	tokenID, err := c.ParamsInt("tokenId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", c.Params("tokenId")))
	}

	sd, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(int64(tokenID), 0))),
		qm.Load(models.SyntheticDeviceRels.VehicleToken),
	).One(c.Context(), co.dbs.DBS().Reader)
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

	intID, _ := mapper.IDFor(int(i))

	udai, err := models.FindUserDeviceAPIIntegration(c.Context(), co.dbs.DBS().Reader, ud.ID, intID)
	if err != nil {
		return err
	}

	if udai.Status != models.UserDeviceAPIIntegrationStatusAuthenticationFailure {
		// TODO(elffjs): Can probably still "succeed" in this case.
		return fiber.NewError(fiber.StatusBadRequest, "Device is not in authentication failure.")
	}

	var scTask services.SmartcarTaskService

	scTask.StopPoll(udai)
	scTask.StartPoll(udai, sd)

	return nil
}
