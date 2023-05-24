package owner

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errNotFound = fiber.NewError(fiber.StatusNotFound, "Device not found.")

// UserDeviceOwner creates a new middleware handler that checks whether a user is authorized to access
// a user device. For the middleware to allow the request to proceed:
//
//   - The request must have a valid JWT, identifying a user.
//   - There must be a userDeviceID path parameter, and that device must exist.
//   - Either the user owns the device, or the user's account has an Ethereum address that
//     owns the corresponding NFT.
func UserDeviceOwner(dbs db.Store, usersClient pb.UserServiceClient, logger *zerolog.Logger) fiber.Handler {
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

// AutoPiOwner creates a new middleware handler that checks whether an autopi is paired.
// For the middleware to allow the request to proceed:
//
//   - The request must have a valid JWT, identifying a user.
//   - There must be a unitID path parameter, and that autopi must exist.
//   - Either the device has not been paired on chain (anyone can access the endpoint) or
//     the user has an address on file that is either the owner of the AutoPi or the owner
//     of the paired vehicle.
func AutoPiOwner(dbs db.Store, usersClient pb.UserServiceClient, logger *zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := helpers.GetUserID(c)
		unitID := c.Params("unitID")
		valid, unitID := services.ValidateAndCleanUUID(unitID)
		if !valid {
			return fiber.NewError(fiber.StatusBadRequest, "Unit id is not a valid UUID.")
		}

		log := logger.With().Str("userId", userID).Str("autoPiUnitId", unitID).Logger()
		c.Locals("userID", userID)
		c.Locals("unitID", unitID)
		c.Locals("logger", &logger)

		autopiUnit, err := models.FindAutopiUnit(c.Context(), dbs.DBS().Reader, unitID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			log.Err(err).Msg("autopi not found")
			return errNotFound
		case err != nil:
			log.Err(err).Msg("unable to query database")
			return err
		}

		if !autopiUnit.PairRequestID.IsZero() && autopiUnit.UnpairRequestID.IsZero() {
			usr, err := usersClient.GetUser(context.Background(), &pb.GetUserRequest{Id: userID})
			if err != nil {
				return err
			}

			userAddr := common.HexToAddress(usr.GetEthereumAddress())
			if (userAddr != common.BytesToAddress(autopiUnit.EthereumAddress.Bytes)) &&
				(userAddr != common.BytesToAddress(autopiUnit.OwnerAddress.Bytes)) {
				log.Info().Msg("user is not owner of paired vehicle or autopi")
				return errors.New("user is not owner of paired vehicle or autopi")
			}
		}

		return nil
	}
}
