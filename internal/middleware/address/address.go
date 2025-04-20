package address

import (
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
)

const (
	addrKey   = "ethereumAddress"
	loggerKey = "logger"
	tokenKey  = "user"
)

const (
	subClaim  = "sub"
	addrClaim = "ethereum_address"
)

var zeroAddr common.Address

func New(logger *zerolog.Logger) fiber.Handler {
	getAddr := func(c *fiber.Ctx) (common.Address, error) {
		token := c.Locals(tokenKey).(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)

		addrAny, ok := claims[addrClaim]
		if ok {
			addrStr, ok := addrAny.(string)
			if !ok {
				return zeroAddr, helpers.APIError(fiber.StatusUnauthorized, "The %s claim has type %T instead of string.", addrClaim, addrAny)
			}

			if !common.IsHexAddress(addrStr) {
				return zeroAddr, helpers.APIError(fiber.StatusUnauthorized, "Couldn't parse %s claim %q as a hex address.", addrClaim, addrStr)
			}

			return common.HexToAddress(addrStr), nil
		}

		return zeroAddr, helpers.APIError(fiber.StatusUnauthorized, "No %s claim found.", addrClaim)
	}

	return func(c *fiber.Ctx) error {
		addr, err := getAddr(c)
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
