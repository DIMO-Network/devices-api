package controllers

import (
	"database/sql"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

type middleware struct {
	DBS         func() *db.ReaderWriter
	UsersClient pb.UserServiceClient
	log         *zerolog.Logger
	Settings    *config.Settings
}

// NewMiddleware if user has valid JWT, confirm that they own the userDeviceID in path or have a confirmed eth address that owns the corresponding nft
func NewMiddleware(settings *config.Settings, dbs func() *db.ReaderWriter, usersClient pb.UserServiceClient, logger *zerolog.Logger) *middleware {
	return &middleware{
		DBS:         dbs,
		UsersClient: usersClient,
		log:         logger,
		Settings:    settings,
	}
}

// DeviceOwnershipMiddleware check that authenticated user owns the userDeviceID in the path or has a confirmed Eth address that owns the corresponding NFT
func (m *middleware) DeviceOwnershipMiddleware(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	l := m.log.With().
		Str("userDeviceId", udi).Str("userId", userID).
		Logger()

	c.Locals("logger", l)

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
	).Exists(c.Context(), m.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fiber.NewError(fiber.StatusNotFound, "no device associated with user id and userDeviceID")
		}
		m.log.Err(err).Msg("error while checking if authenticated user owns device or corresponding nft")
		return err
	}

	if userDevice {
		return c.Next()
	}

	user, err := m.UsersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		m.log.Err(err).Msg("Failed to retrieve user information.")
		return err
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusNotFound, "User does not have an Ethereum address and does not own device associated with userDeviceID.")
	}

	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.OwnerAddress.EQ(null.BytesFrom(common.FromHex(*user.EthereumAddress))),
		models.VehicleNFTWhere.UserDeviceID.EQ(null.StringFrom(udi)),
	).Exists(c.Context(), m.DBS().Reader)
	if err != nil {
		return err
	}

	if !nft {
		return fiber.NewError(fiber.StatusNotFound, "User does not own device or nft associated with userDeviceID.")
	}

	return c.Next()
}
