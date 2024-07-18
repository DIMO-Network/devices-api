package address

import (
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func New(client pb.UserServiceClient, logger *zerolog.Logger) fiber.Handler {
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

		subAny, ok := claims[subClaim]
		if ok {
			subStr, ok := subAny.(string)
			if !ok {
				return zeroAddr, helpers.APIError(fiber.StatusUnauthorized, "The %s claim has type %T instead of string.", subClaim, addrAny)
			}

			user, err := client.GetUser(c.Context(), &pb.GetUserRequest{Id: subStr})
			if err != nil {
				if err, ok := status.FromError(err); ok && err.Code() == codes.NotFound {
					return zeroAddr, helpers.APIError(fiber.StatusForbidden, "No record of user %s.", subStr)
				}
				return zeroAddr, err
			}

			if user.EthereumAddress == nil {
				return zeroAddr, helpers.APIError(fiber.StatusUnauthorized, "User %s has no Ethereum address on file.", subStr)
			}

			if !common.IsHexAddress(*user.EthereumAddress) {
				return zeroAddr, helpers.APIError(fiber.StatusUnauthorized, "Ethereum address %q on file is invalid.", *user.EthereumAddress)
			}

			return common.HexToAddress(*user.EthereumAddress), nil
		}

		return zeroAddr, helpers.APIError(fiber.StatusUnauthorized, "No %s or %s claim found.", addrClaim, subClaim)
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
