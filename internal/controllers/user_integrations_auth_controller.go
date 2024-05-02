package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/shared"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/redis"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

const teslaFleetAuthCacheKey = "integration_credentials_%s"

type UserIntegrationAuthController struct {
	Settings         *config.Settings
	DBS              func() *db.ReaderWriter
	DeviceDefSvc     services.DeviceDefinitionService
	log              *zerolog.Logger
	teslaFleetAPISvc services.TeslaFleetAPIService
	cache            redis.CacheService
	cipher           shared.Cipher
	usersClient      pb.UserServiceClient
}

func NewUserIntegrationAuthController(
	settings *config.Settings,
	dbs func() *db.ReaderWriter,
	logger *zerolog.Logger,
	ddSvc services.DeviceDefinitionService,
	teslaFleetAPISvc services.TeslaFleetAPIService,
	cache redis.CacheService,
	cipher shared.Cipher,
	usersClient pb.UserServiceClient,
) UserIntegrationAuthController {
	return UserIntegrationAuthController{
		Settings:         settings,
		DBS:              dbs,
		DeviceDefSvc:     ddSvc,
		log:              logger,
		teslaFleetAPISvc: teslaFleetAPISvc,
		cache:            cache,
		cipher:           cipher,
		usersClient:      usersClient,
	}
}

// CompleteOAuthExchangeResponseWrapper response wrapper for list of user vehicles
type CompleteOAuthExchangeResponseWrapper struct {
	Vehicles []CompleteOAuthExchangeResponse `json:"vehicles"`
}

// CompleteOAuthExchangeRequest request object for completing tesla OAuth
type CompleteOAuthExchangeRequest struct {
	AuthorizationCode string `json:"authorizationCode"`
	RedirectURI       string `json:"redirectUri"`
	Region            string `json:"region"`
}

// CompleteOAuthExchangeResponse response object for tesla vehicles attached to user account
type CompleteOAuthExchangeResponse struct {
	ExternalID string           `json:"externalId"`
	VIN        string           `json:"vin"`
	Definition DeviceDefinition `json:"definition"`
}

// DeviceDefinition inner definition object containing meta data for each tesla vehicle
type DeviceDefinition struct {
	Make               string `json:"make"`
	Model              string `json:"model"`
	Year               int    `json:"year"`
	DeviceDefinitionID string `json:"id"`
}

// CompleteOAuthExchange godoc
// @Description Complete Tesla auth and get devices for authenticated user
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       tokenID path string                   true "token id for integration"
// @Param       user_device body controllers.CompleteOAuthExchangeRequest true "all fields are required"
// @Security    ApiKeyAuth
// @Success     200 {object} controllers.CompleteOAuthExchangeResponseWrapper
// @Security    BearerAuth
// @Router      /integration/{tokenID}/credentials [post]
func (u *UserIntegrationAuthController) CompleteOAuthExchange(c *fiber.Ctx) error {
	// Get the current user
	userID := helpers.GetUserID(c)
	user, err := u.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		return helpers.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "missing eth address for user!")
	}

	tokenID := c.Params("tokenID")
	tkID, err := strconv.ParseUint(tokenID, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "could not process the provided tokenId!")
	}

	intd, err := u.DeviceDefSvc.GetIntegrationByTokenID(c.Context(), tkID)
	if err != nil {
		u.log.Err(err).Str("Calling Function", "GetIntegrationByTokenID").Uint64("tokenID", tkID).Msg("Error occurred trying to get integration using tokenID")
		return fiber.NewError(fiber.StatusInternalServerError, "an error occurred completing authorization")
	}
	if intd.Vendor != constants.TeslaVendor {
		return fiber.NewError(fiber.StatusBadRequest, "invalid value provided for tokenId!")
	}

	reqBody := new(CompleteOAuthExchangeRequest)
	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request JSON body.")
	}

	if reqBody.Region != "na" && reqBody.Region != "eu" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid value provided for region, only na and eu are allowed")
	}

	logger := u.log.With().
		Str("region", reqBody.Region).
		Str("redirectUri", reqBody.RedirectURI).
		Str("route", c.Route().Path).
		Logger()
	logger.Info().Msg("Attempting to complete Tesla authorization")

	teslaAuth, err := u.teslaFleetAPISvc.CompleteTeslaAuthCodeExchange(c.Context(), reqBody.AuthorizationCode, reqBody.RedirectURI, reqBody.Region)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get tesla authCode:"+err.Error())
	}
	teslaAuth.Region = reqBody.Region
	// Save tesla oauth credentials in cache
	err = u.persistOauthCredentials(c.Context(), *teslaAuth, *user.EthereumAddress)
	if err != nil {
		u.log.Err(err).Msg("an error occurred while trying to persist user auth credentials to cache")
		return fiber.NewError(fiber.StatusInternalServerError, "an error occurred during tesla authorization")
	}

	vehicles, err := u.teslaFleetAPISvc.GetVehicles(c.Context(), teslaAuth.AccessToken, reqBody.Region)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error occurred fetching vehicles:"+err.Error())
	}

	response := make([]CompleteOAuthExchangeResponse, 0, len(vehicles))
	for _, v := range vehicles {
		decodeVIN, err := u.DeviceDefSvc.DecodeVIN(c.Context(), v.VIN, "", 0, "")
		if err != nil || len(decodeVIN.DeviceDefinitionId) == 0 {
			u.log.Err(err).Msg("An error occurred decoding vin for tesla vehicle")
			return fiber.NewError(fiber.StatusFailedDependency, "An error occurred completing tesla authorization")
		}

		dd, err := u.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), decodeVIN.DeviceDefinitionId)
		if err != nil || len(decodeVIN.DeviceDefinitionId) == 0 {
			u.log.Err(err).Str("deviceDefinitionID", decodeVIN.DeviceDefinitionId).Msg("An error occurred fetching device definition using ID")
			return fiber.NewError(fiber.StatusFailedDependency, "An error occurred completing tesla authorization")
		}

		response = append(response, CompleteOAuthExchangeResponse{
			ExternalID: strconv.Itoa(v.ID),
			VIN:        v.VIN,
			Definition: DeviceDefinition{
				Make:               dd.Type.Make,
				Model:              dd.Type.Model,
				Year:               int(dd.Type.Year),
				DeviceDefinitionID: decodeVIN.DeviceDefinitionId,
			},
		})
	}

	vehicleResp := &CompleteOAuthExchangeResponseWrapper{
		Vehicles: response,
	}

	return c.JSON(vehicleResp)
}

func (u *UserIntegrationAuthController) persistOauthCredentials(ctx context.Context, teslaAuth services.TeslaAuthCodeResponse, userEthAddr string) error {
	tokenStr, err := json.Marshal(teslaAuth)
	if err != nil {
		return fmt.Errorf("an error occurred json encoding auth credentials: %w", err)
	}

	encToken, err := u.cipher.Encrypt(string(tokenStr))
	if err != nil {
		return fmt.Errorf("an error occurred encrypting auth credentials: %w", err)
	}

	cacheKey := fmt.Sprintf(teslaFleetAuthCacheKey, userEthAddr)
	status := u.cache.Set(ctx, cacheKey, encToken, 5*time.Minute)
	if status.Err() != nil {
		return fmt.Errorf("an error occurred saving auth credentials to cache: %w", status.Err())
	}

	return nil
}
