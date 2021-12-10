package main

import (
	"context"
	"os"
	"time"

	_ "github.com/DIMO-INC/devices-api/docs"
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/controllers"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v3"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

// @title     DIMO Devices API
// @version   2.0
// @BasePath  /v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
func main() {
	gitSha1 := os.Getenv("GIT_SHA1")
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Str("git-sha1", gitSha1).
		Logger()

	settings, err := config.LoadConfig("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("could not load settings")
	}
	pdb := database.NewDbConnectionFromSettings(ctx, settings)

	arg := ""
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}
	switch arg {
	case "migrate":
		migrateDatabase(logger, settings)
	case "seed-smartcar":
		loadSmartCarData(ctx, logger, settings, pdb)
	default:
		startWebAPI(logger, settings, pdb)
	}
}

func startWebAPI(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return ErrorHandler(c, err, logger)
		},
		DisableStartupMessage: true,
	})
	nhtsaSvc := services.NewNHTSAService()
	deviceControllers := controllers.NewDevicesController(settings, pdb.DBS, &logger, nhtsaSvc)
	userDeviceControllers := controllers.NewUserDevicesController(settings, pdb.DBS, &logger)

	app.Use(recover.New(recover.Config{
		Next:              nil,
		EnableStackTrace:  true,
		StackTraceHandler: nil,
	}))
	app.Use(cors.New())
	cacheHandler := cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Query("refresh") == "true"
		},
		Expiration:   30 * time.Minute,
		CacheControl: true,
	})
	app.Get("/", HealthCheck)
	v1 := app.Group("/v1")

	v1.Get("/device-definitions/vin/:vin", deviceControllers.LookupDeviceDefinitionByVIN) // generic response, specific for vehicle lookup
	v1.Get("/device-definitions/all", cacheHandler, deviceControllers.GetAllDeviceMakeModelYears)
	v1.Get("/device-definitions/:id", deviceControllers.GetDeviceDefinitionByID)
	v1.Get("/device-definitions/:id/integrations", deviceControllers.GetIntegrationsByID)
	// secured paths
	jwtAuth := jwtware.New(jwtware.Config{KeySetURL: settings.JwtKeySetURL})
	v1.Get("/user/devices/me", jwtAuth, userDeviceControllers.GetUserDevices)
	v1.Post("/user/devices", jwtAuth, userDeviceControllers.RegisterDeviceForUser)
	v1.Post("/user/devices/:id/integrations/smartcar", jwtAuth, userDeviceControllers.RegisterSmartCarIntegration)

	// swagger - note could add auth middleware so it is not open
	sc := swagger.Config{ // custom
		// Expand ("list") or Collapse ("none") tag groups by default
		DocExpansion: "list",
	}
	v1.Get("/swagger/*", swagger.New(sc))

	logger.Info().Msg("Server started on port " + settings.Port)
	// Start Server
	if err := app.Listen(":" + settings.Port); err != nil {
		logger.Fatal().Err(err)
	}
}

// HealthCheck godoc
// @Summary Show the status of server.
// @Description get the status of server.
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func HealthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	if err := c.JSON(res); err != nil {
		return err
	}

	return nil
}

// ErrorHandler custom handler to log recovered errors using our logger and return json instead of string
func ErrorHandler(c *fiber.Ctx, err error, logger zerolog.Logger) error {
	code := fiber.StatusInternalServerError // Default 500 statuscode

	if e, ok := err.(*fiber.Error); ok {
		// Override status code if fiber.Error type
		code = e.Code
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	logger.Err(err).Msg("caught a panic")

	return c.Status(code).JSON(fiber.Map{
		"error": true,
		"msg":   err.Error(),
	})
}
