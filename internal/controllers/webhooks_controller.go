package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
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

func validateSignature(secret, data, expectedSignature string) bool {
	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secret))
	// Write Data to it
	h.Write([]byte(data))
	// Get result and encode as hexadecimal string
	sha := hex.EncodeToString(h.Sum(nil))

	return sha == expectedSignature
}
