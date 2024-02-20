package controllers

import (
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/redis"
	"github.com/Shopify/sarama"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"strconv"
)

type UserIntegrationAuthController struct {
	Settings         *config.Settings
	DBS              func() *db.ReaderWriter
	DeviceDefSvc     services.DeviceDefinitionService
	log              *zerolog.Logger
	cipher           shared.Cipher
	producer         sarama.SyncProducer
	redisCache       redis.CacheService
	teslaFleetAPISvc services.TeslaFleetAPIService
}

func NewUserIntegrationAuthController(
	settings *config.Settings,
	dbs func() *db.ReaderWriter,
	logger *zerolog.Logger,
	ddSvc services.DeviceDefinitionService,
	cipher shared.Cipher,
	producer sarama.SyncProducer,
	cache redis.CacheService,
	teslaFleetAPISvc services.TeslaFleetAPIService,
) UserIntegrationAuthController {
	return UserIntegrationAuthController{
		Settings:         settings,
		DBS:              dbs,
		DeviceDefSvc:     ddSvc,
		log:              logger,
		cipher:           cipher,
		producer:         producer,
		redisCache:       cache,
		teslaFleetAPISvc: teslaFleetAPISvc,
	}
}

type CompleteOAuthExchangeResponseWrapper struct {
	Vehicles []CompleteOAuthExchangeResponse `json:"vehicles"`
}

type DeviceDefinition struct {
	Make               string `json:"make"`
	Model              string `json:"model"`
	Year               int    `json:"year"`
	DeviceDefinitionID string `json:"deviceDefinitionId"`
}

type CompleteOAuthExchangeResponse struct {
	ExternalID string           `json:"externalId"`
	VehicleID  int              `json:"vehicleId"`
	VIN        string           `json:"vin"`
	Definition DeviceDefinition `json:"definition"`
}

// CompleteOAuthExchange godoc
// @Description Complete Tesla auth and get devices for authenticated user
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       user_device body controllers.CompleteOAuthExchangeRequest true "all fields are required"
// @Security    ApiKeyAuth
// @Success     200 {object} controllers.CompleteOAuthExchangeResponse
// @Security    BearerAuth
// @Router      /integration/:tokenID/credentials [post]
func (u *UserIntegrationAuthController) CompleteOAuthExchange(c *fiber.Ctx) error {
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

	teslaAuth, err := u.teslaFleetAPISvc.CompleteTeslaAuthCodeExchange(reqBody.AuthorizationCode, reqBody.RedirectURI, reqBody.Region)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get tesla authCode:"+err.Error())
	}

	vehicles, err := u.teslaFleetAPISvc.GetVehicles(teslaAuth.AccessToken, reqBody.Region)
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
			VehicleID:  v.VehicleID,
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
