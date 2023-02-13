package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type WebhooksController struct {
	dbs             func() *db.ReaderWriter
	settings        *config.Settings
	log             *zerolog.Logger
	autoPiSvc       services.AutoPiAPIService
	deviceDefIntSvc services.DeviceDefinitionIntegrationService
}

func NewWebhooksController(settings *config.Settings, dbs func() *db.ReaderWriter, log *zerolog.Logger, autoPiSvc services.AutoPiAPIService, deviceDefIntSvc services.DeviceDefinitionIntegrationService) WebhooksController {
	return WebhooksController{
		dbs:             dbs,
		settings:        settings,
		log:             log,
		autoPiSvc:       autoPiSvc,
		deviceDefIntSvc: deviceDefIntSvc,
	}
}

// ProcessCommand handles the command webhook request. even when there is an error we'll return 204 to avoid autopi retrying against us
func (wc *WebhooksController) ProcessCommand(c *fiber.Ctx) error {
	logger := wc.log.With().
		Str("integration", "autopi").
		Str("handler", "webhooks.ProcessCommand").
		Logger()
	logger.Info().Msg("attempting to process webhook request")
	// grab values from webhook using gjson instead of parsing since more error-prone as they make changes
	apwJID := gjson.GetBytes(c.Body(), "jid")
	apwDeviceID := gjson.GetBytes(c.Body(), "device_id")
	apwState := gjson.GetBytes(c.Body(), "state")
	// grab value, like for pulling VIN or other obd query responses
	cmdValue := gjson.GetBytes(c.Body(), "response.data.return.value")

	if !apwJID.Exists() || !apwDeviceID.Exists() || !apwState.Exists() {
		logger.Error().Str("payload", string(c.Body())).Msg("no jobId or deviceId found in payload")
		return fiber.NewError(fiber.StatusBadRequest, "invalid autopi webhook request payload")
	}
	logger = logger.With().Str("autopi deviceID", apwDeviceID.String()).Str("state", apwState.String()).Str("jobID", apwJID.String()).Logger()

	// hmac signature validation
	reqSig := c.Get("X-Request-Signature")
	if !validateSignature(wc.settings.AutoPiAPIToken, string(c.Body()), reqSig) {
		logger.Error().Str("payload", string(c.Body())).Msg("invalid webhook signature")
		return fiber.NewError(fiber.StatusUnauthorized, "invalid autopi webhook signature")
	}
	var cmdResult *services.AutoPiCommandResult
	if cmdValue.Exists() {
		cmdResult = &services.AutoPiCommandResult{
			Value: cmdValue.String(),
			Tag:   gjson.GetBytes(c.Body(), "response.tag").String(),
			Type:  gjson.GetBytes(c.Body(), "response.data.return._type").String(),
		}
	}

	autopiJob, err := wc.autoPiSvc.UpdateJob(c.Context(), apwJID.String(), apwState.String(), cmdResult)
	if err != nil {
		logger.Err(err).Msg("error updating autopi job")
		return c.SendStatus(fiber.StatusNoContent)
	}
	// if we can link the autopi job to a device, it could be a job related to an integration registration sync command
	if !autopiJob.UserDeviceID.IsZero() && autopiJob.Command == "state.sls pending" && strings.EqualFold(apwState.String(), "COMMAND_EXECUTED") {
		autoPiInteg, err := wc.deviceDefIntSvc.GetAutoPiIntegration(c.Context())
		if err != nil {
			logger.Err(err).Msg("could not create or get autopi integration record")
			return c.SendStatus(fiber.StatusNoContent)
		}
		// we could have situation where there are multiple results, eg. if the AutoPi was moved from one car to another, so order by updated_at desc and grab first
		apiIntegration, err := models.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(autoPiInteg.Id),
			models.UserDeviceAPIIntegrationWhere.ExternalID.EQ(null.StringFrom(apwDeviceID.String())),
			models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(autopiJob.UserDeviceID.String),
			qm.OrderBy("updated_at desc"), qm.Limit(1)).One(c.Context(), wc.dbs().Reader)
		if err != nil {
			logger.Err(err).Msg("could not get user device api integrations")
			return c.SendStatus(fiber.StatusNoContent)
		}
		// get the metadata so we can update it
		udMetadata := new(services.UserDeviceAPIIntegrationsMetadata)
		err = apiIntegration.Metadata.Unmarshal(udMetadata)
		if err != nil {
			logger.Err(err).Msg("failed to unmarshal user device api integrations metadata column into struct")
			return c.SendStatus(fiber.StatusNoContent)
		}
		// update the integration state, Pending first data means we are succesfully paired and template applied, just waiting for data to stream
		apiIntegration.Status = models.UserDeviceAPIIntegrationStatusPendingFirstData
		ss := constants.TemplateConfirmed.String()
		udMetadata.AutoPiSubStatus = &ss
		// update database
		err = apiIntegration.Metadata.Marshal(udMetadata)
		if err != nil {
			logger.Err(err).Msg("failed to marshal user device api integration metadata json from autopi webhook")
			return c.SendStatus(fiber.StatusNoContent)
		}
		_, err = apiIntegration.Update(c.Context(), wc.dbs().Writer, boil.Whitelist(
			models.UserDeviceAPIIntegrationColumns.Metadata, models.UserDeviceAPIIntegrationColumns.Status,
			models.UserDeviceAPIIntegrationColumns.UpdatedAt))
		if err != nil {
			logger.Err(err).Msg("failed to save user device integration changes")
			return c.SendStatus(fiber.StatusNoContent)
		}
	}
	logger.Info().Msg("processed webhook successfully")

	return c.SendStatus(fiber.StatusNoContent)
}

func validateSignature(secret, data, expectedSignature string) bool {
	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secret))
	// Write Data to it
	h.Write([]byte(data))
	// Get result and encode as hexadecimal string
	sha := hex.EncodeToString(h.Sum(nil))

	return sha == expectedSignature
}
