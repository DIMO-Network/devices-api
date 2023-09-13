package owner

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errNotFound = fiber.NewError(fiber.StatusNotFound, "Device not found.")

// UserDevice creates a new middleware handler that checks whether a user is authorized to access
// a user device. For the middleware to allow the request to proceed:
//
//   - The request must have a valid JWT, identifying a user.
//   - There must be a userDeviceID path parameter, and that device must exist.
//   - Either the user owns the device, or the user's account has an Ethereum address that
//     owns the corresponding NFT.
func UserDevice(dbs db.Store, usersClient pb.UserServiceClient, logger *zerolog.Logger) fiber.Handler {
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

		user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
				return errNotFound
			}
			return err
		}

		if user.EthereumAddress == nil {
			return errNotFound
		}

		if userAddrOwns, err := models.VehicleNFTS(
			models.VehicleNFTWhere.UserDeviceID.EQ(null.StringFrom(udi)),
			models.VehicleNFTWhere.OwnerAddress.EQ(null.BytesFrom(common.FromHex(*user.EthereumAddress))),
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
func AftermarketDevice(dbs db.Store, usersClient pb.UserServiceClient, logger *zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := helpers.GetUserID(c)

		serial := c.Params("serial")
		serial = strings.TrimSpace(strings.ToLower(serial)) // A UUID for AutoPi. An 11-digit number for Macaron.

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

		user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
		if err != nil {
			return helpers.GrpcErrorToFiber(err, "Error retrieving user")
		}

		if user.EthereumAddress == nil {
			return fiber.NewError(fiber.StatusForbidden, "user does not have a valid ethereum address")
		}

		userAddr := common.HexToAddress(*user.EthereumAddress)
		apOwner := common.BytesToAddress(aftermarketDevice.OwnerAddress.Bytes)

		if userAddr == apOwner {
			return c.Next()
		}

		ownsVehicle, err := models.VehicleNFTS(
			models.VehicleNFTWhere.TokenID.EQ(types.NewNullDecimal(aftermarketDevice.VehicleTokenID.Big)),
			models.VehicleNFTWhere.OwnerAddress.EQ(null.BytesFrom(userAddr.Bytes())),
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
