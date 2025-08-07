package owner

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
)

var errNotFound = fiber.NewError(fiber.StatusNotFound, "Device not found.")

// UserDevice creates a new middleware handler that checks whether a user is authorized to access
// a user device. For the middleware to allow the request to proceed:
//
//   - The request must have a valid JWT, identifying a user.
//   - There must be a userDeviceID path parameter, and that device must exist.
//   - Either the user owns the device, or the user's account has an Ethereum address that
//     owns the corresponding NFT.
func UserDevice(dbs db.Store, logger *zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := helpers.GetUserID(c)
		udi := c.Params("userDeviceID")

		c.Locals("userID", userID)
		c.Locals("userDeviceID", udi)

		logger := logger.With().Str("userId", userID).Str("userDeviceId", udi).Logger()

		c.Locals("logger", &logger)

		if userIDOwns, err := models.UserDevices(
			models.UserDeviceWhere.ID.EQ(udi),
			models.UserDeviceWhere.UserID.EQ(userID),
		).Exists(c.Context(), dbs.DBS().Reader); err != nil {
			return err
		} else if userIDOwns {
			return c.Next()
		}

		userAddr, err := helpers.GetJWTEthAddr(c)
		if err != nil {
			return err
		}

		if userAddrOwns, err := models.UserDevices(
			models.UserDeviceWhere.ID.EQ(udi),
			models.UserDeviceWhere.OwnerAddress.EQ(null.BytesFrom(userAddr.Bytes())),
		).Exists(c.Context(), dbs.DBS().Reader); err != nil {
			return err
		} else if userAddrOwns {
			return c.Next()
		}

		return errNotFound
	}
}

// AftermarketDevice creates a new middleware handler that checks whether an autopi is paired.
// For the middleware to allow the request to proceed:
//
//   - The request must have a valid JWT, identifying a user.
//   - There must be a serial path parameter, and that autopi must exist (serial = unitID).
//   - Either the device has not been paired on chain (anyone can access the endpoint) or
//     the user has an address on file that is either the owner of the AftermarketDevice or the owner
//     of the paired vehicle.
func AftermarketDevice(dbs db.Store, logger *zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := helpers.GetUserID(c)

		serial := c.Params("serial")
		serial = strings.TrimSpace(serial)
		if len(serial) == 36 {
			// The lowercasing here is really just for AutoPi's UUIDs.
			serial = strings.ToLower(serial)
		}

		logger := logger.With().Str("userId", userID).Str("serial", serial).Logger()
		c.Locals("userID", userID)
		c.Locals("serial", serial)
		c.Locals("logger", &logger)

		aftermarketDevice, err := models.AftermarketDevices(models.AftermarketDeviceWhere.Serial.EQ(serial)).One(c.Context(), dbs.DBS().Reader)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fiber.NewError(fiber.StatusNotFound, "AftermarketDevice not minted, or serial is invalid.")
			}
			return err
		}

		// If token_id is null, device is not paired.
		// Also short-circuit the address checks if user is the "web2 owner".
		if aftermarketDevice.VehicleTokenID.IsZero() || aftermarketDevice.UserID.Valid && aftermarketDevice.UserID.String == userID {
			return c.Next()
		}

		userAddr, err := helpers.GetJWTEthAddr(c)
		if err != nil {
			return err
		}

		apOwner := common.BytesToAddress(aftermarketDevice.OwnerAddress.Bytes)

		if userAddr == apOwner {
			return c.Next()
		}

		ownsVehicle, err := models.UserDevices(
			models.UserDeviceWhere.TokenID.EQ(types.NewNullDecimal(aftermarketDevice.VehicleTokenID.Big)),
			models.UserDeviceWhere.OwnerAddress.EQ(null.BytesFrom(userAddr.Bytes())),
		).Exists(c.Context(), dbs.DBS().Reader)
		if err != nil {
			return err
		}
		if !ownsVehicle {
			return fiber.NewError(fiber.StatusForbidden, "user is not owner of paired vehicle or AftermarketDevice")
		}

		return c.Next()
	}
}
