package address

import (
	"errors"
	"fmt"

	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const tokenKey = "user"

const subClaim = "sub"
const addrClaim = "ethereum_address"

const addrKey = "ethereumAddress"

func New() fiber.Handler {
	var client pb.UserServiceClient

	return func(c *fiber.Ctx) error {
		token, ok := c.Locals(tokenKey).(*jwt.Token)
		if !ok {
			return fmt.Errorf("token has unexpected type %T", token)
		}
		if token == nil {
			return errors.New("token is nil")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return fmt.Errorf("claims object has unexpected type %T", token)
		}
		if claims == nil {
			return errors.New("claims object is nil")
		}

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
