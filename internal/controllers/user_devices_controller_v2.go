package controllers

import (
	"log"

	"github.com/DIMO-Network/shared/db"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/DIMO-Network/devices-api/internal/config"
)

type UserDevicesControllerV2 struct {
	Settings *config.Settings
	DBS      func() *db.ReaderWriter
	log      *zerolog.Logger
}

func NewUserDevicesControllerV2(settings *config.Settings, dbs func() *db.ReaderWriter, logger *zerolog.Logger) UserDevicesControllerV2 {
	return UserDevicesControllerV2{
		Settings: settings,
		DBS:      dbs,
		log:      logger,
	}
}

// GetRange godoc
// @Description gets the estimated range for a particular user device
// @Tags        user-devices
// @Produce     json
// @Success     200 {object} controllers.DeviceRange
// @Security    BearerAuth
// @Param       userDeviceID path string true "user device id"
// @Router      /v2/vehicles/{tokenId}/analytics/range [get]
func (udc *UserDevicesControllerV2) GetRange(c *fiber.Ctx) error {
	tokenID := c.Params("tokenId")
	log.Println(tokenID)
	return nil
}
