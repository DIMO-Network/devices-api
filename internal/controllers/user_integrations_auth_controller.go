package controllers

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/middleware/address"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/tmpcred"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserIntegrationAuthController struct {
	Settings         *config.Settings
	DBS              func() *db.ReaderWriter
	DeviceDefSvc     services.DeviceDefinitionService
	log              *zerolog.Logger
	teslaFleetAPISvc services.TeslaFleetAPIService
	store            CredStore
}

//go:generate mockgen -destination=cred_store_mock_test.go -package controllers . CredStore
type CredStore interface {
	Store(ctx context.Context, user common.Address, cred *tmpcred.Credential) error
}

func NewUserIntegrationAuthController(
	settings *config.Settings,
	dbs func() *db.ReaderWriter,
	logger *zerolog.Logger,
	ddSvc services.DeviceDefinitionService,
	teslaFleetAPISvc services.TeslaFleetAPIService,
	credStore CredStore,
) UserIntegrationAuthController {

	return UserIntegrationAuthController{
		Settings:         settings,
		DBS:              dbs,
		DeviceDefSvc:     ddSvc,
		log:              logger,
		teslaFleetAPISvc: teslaFleetAPISvc,
		store:            credStore,
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

type partialTeslaClaims struct {
	jwt.RegisteredClaims
	Scopes []string `json:"scp"`

	// For debugging.
	OUCode string `json:"ou_code"`
}

var teslaDataScope = "vehicle_device_data"

var teslaCodeFailureCount = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "devices_api",
		Subsystem: "tesla",
		Name:      "code_exchange_failures_total",
		Help:      "Known strains of failure during Tesla authorization code exchange and ensuing vehicle display.",
	},
	[]string{"type"},
)

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
	userAddr := address.Get(c)
	logger := helpers.GetLogger(c, u.log)

	tokenIDRaw := c.Params("tokenID")
	tokenID, err := strconv.ParseUint(tokenIDRaw, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse integration token id %q.", tokenIDRaw))
	}

	intd, err := u.DeviceDefSvc.GetIntegrationByTokenID(c.Context(), tokenID)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("No integration with token id %d known.", tokenID))
		}
		return fmt.Errorf("error looking up integration %d: %w", tokenID, err)
	}

	if intd.Vendor != constants.TeslaVendor {
		return fiber.NewError(fiber.StatusBadRequest, "Endpoint only valid for the Tesla integration.")
	}

	var reqBody CompleteOAuthExchangeRequest
	if err := c.BodyParser(&reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse JSON request body.")
	}

	logger.Info().Msg("Attempting to complete Tesla authorization")

	teslaAuth, err := u.teslaFleetAPISvc.CompleteTeslaAuthCodeExchange(c.Context(), reqBody.AuthorizationCode, reqBody.RedirectURI)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get tesla authCode:"+err.Error())
	}

	if teslaAuth.RefreshToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Code exchange did not return a refresh token. Make sure you've granted offline_access.")
	}

	var claims partialTeslaClaims
	_, _, err = jwt.NewParser().ParseUnverified(teslaAuth.AccessToken, &claims)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Code exchange returned an unparseable access token.")
	}

	if !slices.Contains(claims.Scopes, teslaDataScope) {
		teslaCodeFailureCount.WithLabelValues("missing_scope").Inc()
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Missing scope %s.", teslaDataScope))
	}

	// Save tesla oauth credentials in cache
	if err := u.store.Store(c.Context(), userAddr, &tmpcred.Credential{
		IntegrationID: int(tokenID),
		AccessToken:   teslaAuth.AccessToken,
		RefreshToken:  teslaAuth.RefreshToken,
		Expiry:        teslaAuth.Expiry,
	}); err != nil {
		return fmt.Errorf("error persisting credentials: %w", err)
	}

	vehicles, err := u.teslaFleetAPISvc.GetVehicles(c.Context(), teslaAuth.AccessToken)
	if err != nil {
		logger.Err(err).Str("subject", claims.Subject).Str("ouCode", claims.OUCode).Interface("audience", claims.Audience).Msg("Error retrieving vehicles.")
		return fiber.NewError(fiber.StatusInternalServerError, "Couldn't fetch vehicles from Tesla.")
	}

	response := make([]CompleteOAuthExchangeResponse, 0, len(vehicles))
	for _, v := range vehicles {
		queryVIN := shared.VIN(v.VIN) // Try to help decoding out with model and year hints.
		decodeVIN, err := u.DeviceDefSvc.DecodeVIN(c.Context(), v.VIN, queryVIN.TeslaModel(), queryVIN.Year(), "")
		if err != nil {
			teslaCodeFailureCount.WithLabelValues("vin_decode").Inc()
			logger.Err(err).Str("vin", v.VIN).Msg("Failed to decode Tesla VIN.")
			return fiber.NewError(fiber.StatusFailedDependency, "An error occurred completing tesla authorization")
		}

		dd, err := u.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), decodeVIN.DeviceDefinitionId)
		if err != nil {
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
