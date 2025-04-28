package address

import (
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

const (
	addrKey   = "ethereumAddress"
	loggerKey = "logger"
	tokenKey  = "user"
)

func New(logger *zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		addr, err := helpers.GetJWTEthAddr(c)
		if err != nil {
			return err
		}

		c.Locals(addrKey, addr)

		logger := logger.With().Str("user", addr.Hex()).Logger()
		c.Locals(loggerKey, &logger)

		return c.Next()
	}
}

func Get(c *fiber.Ctx) common.Address {
	return c.Locals(addrKey).(common.Address)
}
