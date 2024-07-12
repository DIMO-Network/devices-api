package address

import (
	"errors"
	"fmt"

	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	addrKey  = "ethereumAddress"
	tokenKey = "user"
)

const (
	subClaim  = "sub"
	addrClaim = "ethereum_address"
)

func New() fiber.Handler {
	var client pb.UserServiceClient

	return func(c *fiber.Ctx) error {
		token := c.Locals(tokenKey).(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)

		addrAny, ok := claims[addrClaim]
		if ok {
			addrStr, ok := addrAny.(string)
			if !ok {
				return fmt.Errorf("%s claim had unexpected type %T", addrClaim, addrAny)
			}

			if !common.IsHexAddress(addrStr) {
				return fmt.Errorf("couldn't parse %s claim %q as a hex address", addrClaim, addrStr)
			}

			c.Locals(addrKey, common.HexToAddress(addrStr))
			return c.Next()
		}

		subAny, ok := claims[subClaim]
		if ok {
			subStr, ok := subAny.(string)
			if !ok {
				return fmt.Errorf("%s claim had unexpected type %T", subClaim, addrAny)
			}

			user, err := client.GetUser(c.Context(), &pb.GetUserRequest{Id: subStr})
			if err != nil {
				return err
			}

			if user.EthereumAddress == nil {
				return errors.New("user does not have an Ethereum address on file")
			}

			if !common.IsHexAddress(*user.EthereumAddress) {
				return errors.New("address on file is invalid")
			}

			c.Locals(addrKey, common.HexToAddress(*user.EthereumAddress))
			return c.Next()
		}

		return errors.New("no sub claim in token")
	}
}

func Get(c *fiber.Ctx) common.Address {
	return c.Locals(addrKey).(common.Address)
}
