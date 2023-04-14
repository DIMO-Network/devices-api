package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/DIMO-Network/shared/redis"

	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"

	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"

	"github.com/DIMO-Network/devices-api/internal/api"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/autopi"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	pbuser "github.com/DIMO-Network/shared/api/users"
	pr "github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/DIMO-Network/zflogger"
	"github.com/Shopify/sarama"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func startWebAPI(logger zerolog.Logger, settings *config.Settings, pdb db.Store, eventService services.EventService, producer sarama.SyncProducer, s3ServiceClient *s3.Client, s3NFTServiceClient *s3.Client) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return helpers.ErrorHandler(c, err, logger, settings.IsProduction())
		},
		DisableStartupMessage: true,
		ReadBufferSize:        16000,
		BodyLimit:             10 * 1024 * 1024,
	})

	var cipher shared.Cipher
	if settings.Environment == "dev" || settings.IsProduction() {
		cipher = createKMS(settings, &logger)
	} else {
		logger.Warn().Msg("Using ROT13 encrypter. Only use this for testing!")
		cipher = new(shared.ROT13Cipher)
	}

	gcon, err := grpc.Dial(settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed dialing users-api.")
	}
	usersClient := pbuser.NewUserServiceClient(gcon)

	// services
	nhtsaSvc := services.NewNHTSAService()
	ddIntSvc := services.NewDeviceDefinitionIntegrationService(pdb.DBS, settings)
	ddSvc := services.NewDeviceDefinitionService(pdb.DBS, &logger, nhtsaSvc, settings)
	scTaskSvc := services.NewSmartcarTaskService(settings, producer)
	smartcarClient := services.NewSmartcarClient(settings)
	teslaTaskService := services.NewTeslaTaskService(settings, producer)
	teslaSvc := services.NewTeslaService(settings)
	autoPiSvc := services.NewAutoPiAPIService(settings, pdb.DBS)
	autoPiIngest := services.NewIngestRegistrar(services.AutoPi, producer)
	deviceDefinitionRegistrar := services.NewDeviceDefinitionRegistrar(producer, settings)
	autoPiTaskService := services.NewAutoPiTaskService(settings, autoPiSvc, pdb.DBS, logger)
	drivlyTaskService := services.NewDrivlyTaskService(settings, ddSvc, logger)
	hardwareTemplateService := autopi.NewHardwareTemplateService(autoPiSvc, pdb.DBS, &logger)
	autoPi := autopi.NewIntegration(pdb.DBS, ddSvc, autoPiSvc, autoPiTaskService, autoPiIngest, eventService, deviceDefinitionRegistrar, hardwareTemplateService, &logger)
	openAI := services.NewOpenAI(&logger, *settings)

	redisCache := redis.NewRedisCacheService(settings.IsProduction(), redis.Settings{
		URL:       settings.RedisURL,
		Password:  settings.RedisPassword,
		TLS:       settings.RedisTLS,
		KeyPrefix: "devices-api",
	})

	// controllers
	deviceControllers := controllers.NewDevicesController(settings, pdb.DBS, &logger, nhtsaSvc, ddSvc, ddIntSvc)
	userDeviceController := controllers.NewUserDevicesController(settings, pdb.DBS, &logger, ddSvc, ddIntSvc, eventService, smartcarClient, scTaskSvc, teslaSvc, teslaTaskService, cipher, autoPiSvc, services.NewNHTSAService(), autoPiIngest, deviceDefinitionRegistrar, autoPiTaskService, producer, s3NFTServiceClient, drivlyTaskService, autoPi, redisCache, openAI, usersClient)
	geofenceController := controllers.NewGeofencesController(settings, pdb.DBS, &logger, producer, ddSvc)
	webhooksController := controllers.NewWebhooksController(settings, pdb.DBS, &logger, autoPiSvc, ddIntSvc)
	documentsController := controllers.NewDocumentsController(settings, &logger, s3ServiceClient, pdb.DBS)

	// commenting this out b/c the library includes the path in the metrics which saturates prometheus queries - need to fork / make our own
	//prometheus := fiberprometheus.New("devices-api")
	//app.Use(prometheus.Middleware)

	app.Use(fiberrecover.New(fiberrecover.Config{
		Next:              nil,
		EnableStackTrace:  true,
		StackTraceHandler: nil,
	}))
	//cors
	app.Use(cors.New())
	// request logging
	app.Use(zflogger.New(logger, nil))
	//cache
	cacheHandler := cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Query("refresh") == "true"
		},
		Expiration:   1 * time.Minute,
		CacheControl: true,
	})

	// application routes
	app.Get("/", healthCheck)

	v1 := app.Group("/v1")

	v1.Get("/swagger/*", swagger.HandlerDefault)
	// Device Definitions
	v1.Get("/device-definitions/:id", cacheHandler, deviceControllers.GetDeviceDefinitionByID)
	v1.Get("/device-definitions/:id/integrations", cacheHandler, deviceControllers.GetDeviceIntegrationsByID)
	v1.Get("/device-definitions", deviceControllers.GetDeviceDefinitionByMMY)

	nftController := controllers.NewNFTController(settings, pdb.DBS, &logger, s3NFTServiceClient, ddSvc, scTaskSvc, teslaTaskService, ddIntSvc)
	v1.Get("/vehicle/:tokenID", nftController.GetNFTMetadata)
	v1.Get("/vehicle/:tokenID/image", nftController.GetNFTImage)

	v1.Get("/aftermarket/device/:tokenID", nftController.GetAftermarketDeviceNFTMetadata)
	v1.Get("/aftermarket/device/:tokenID/image", nftController.GetAftermarketDeviceNFTImage)
	v1.Get("/manufacturer/:tokenID", nftController.GetManufacturerNFTMetadata)

	// webhooks, performs signature validation
	v1.Post(constants.AutoPiWebhookPath, webhooksController.ProcessCommand)

	// secured paths
	keyRefreshInterval := time.Hour
	keyRefreshUnknownKID := true
	jwtAuth := jwtware.New(jwtware.Config{
		KeySetURL:            settings.JwtKeySetURL,
		KeyRefreshInterval:   &keyRefreshInterval,
		KeyRefreshUnknownKID: &keyRefreshUnknownKID,
	})

	v1Auth := app.Group("/v1")

	if settings.EnablePrivileges {
		privilegeAuth := jwtware.New(jwtware.Config{
			KeySetURL:            settings.TokenExchangeJWTKeySetURL,
			KeyRefreshInterval:   &keyRefreshInterval,
			KeyRefreshUnknownKID: &keyRefreshUnknownKID,
		})

		vPriv := app.Group("/v1/vehicle/:tokenID", privilegeAuth)

		tk := pr.New(pr.Config{
			Log: &logger,
		})

		vehicleAddr := common.HexToAddress(settings.VehicleNFTAddress)

		// vehicle command privileges
		vPriv.Get("/status", tk.OneOf(vehicleAddr, []int64{controllers.NonLocationData, controllers.CurrentLocation, controllers.AllTimeLocation}), nftController.GetVehicleStatus)
		vPriv.Post("/commands/doors/unlock", tk.OneOf(vehicleAddr, []int64{controllers.Commands}), nftController.UnlockDoors)
		vPriv.Post("/commands/doors/lock", tk.OneOf(vehicleAddr, []int64{controllers.Commands}), nftController.LockDoors)
		vPriv.Post("/commands/trunk/open", tk.OneOf(vehicleAddr, []int64{controllers.Commands}), nftController.OpenTrunk)
		vPriv.Post("/commands/frunk/open", tk.OneOf(vehicleAddr, []int64{controllers.Commands}), nftController.OpenFrunk)
	}

	v1Auth.Use(jwtAuth)
	// user's devices
	v1Auth.Get("/user/devices/me", userDeviceController.GetUserDevices)

	if settings.EnablePrivileges {
		v1Auth.Get("/user/devices/shared", userDeviceController.GetSharedDevices)
	}

	v1Auth.Post("/user/devices/fromvin", userDeviceController.RegisterDeviceForUserFromVIN)
	v1Auth.Post("/user/devices/fromsmartcar", userDeviceController.RegisterDeviceForUserFromSmartcar)
	v1Auth.Post("/user/devices", userDeviceController.RegisterDeviceForUser)

	v1Auth.Delete("/user/devices/:userDeviceID", userDeviceController.DeleteUserDevice)
	v1Auth.Patch("/user/devices/:userDeviceID/vin", userDeviceController.UpdateVIN).Name("UpdateVIN")
	v1Auth.Patch("/user/devices/:userDeviceID/name", userDeviceController.UpdateName)
	v1Auth.Patch("/user/devices/:userDeviceID/country-code", userDeviceController.UpdateCountryCode)
	v1Auth.Patch("/user/devices/:userDeviceID/image", userDeviceController.UpdateImage)
	v1Auth.Get("/user/devices/:userDeviceID/valuations", userDeviceController.GetValuations)
	v1Auth.Get("/user/devices/:userDeviceID/offers", userDeviceController.GetOffers)
	v1Auth.Get("/user/devices/:userDeviceID/range", userDeviceController.GetRange)
	v1Auth.Get("/user/devices/:userDeviceID/status", userDeviceController.GetUserDeviceStatus)
	v1Auth.Post("/user/devices/:userDeviceID/error-codes", userDeviceController.QueryDeviceErrorCodes)
	v1Auth.Get("/user/devices/:userDeviceID/error-codes", userDeviceController.GetUserDeviceErrorCodeQueries)

	// device integrations
	v1Auth.Get("/user/devices/:userDeviceID/integrations/:integrationID", userDeviceController.GetUserDeviceIntegration)
	v1Auth.Delete("/user/devices/:userDeviceID/integrations/:integrationID", userDeviceController.DeleteUserDeviceIntegration)
	v1Auth.Post("/user/devices/:userDeviceID/integrations/:integrationID", userDeviceController.RegisterDeviceIntegration)
	v1Auth.Post("/user/devices/:userDeviceID/commands/refresh", userDeviceController.RefreshUserDeviceStatus)

	// Device commands.
	v1Auth.Get("/user/devices/:userDeviceID/integrations/:integrationID/commands/:requestID", userDeviceController.GetCommandRequestStatus)
	v1Auth.Post("/user/devices/:userDeviceID/integrations/:integrationID/commands/doors/unlock", userDeviceController.UnlockDoors)
	v1Auth.Post("/user/devices/:userDeviceID/integrations/:integrationID/commands/doors/lock", userDeviceController.LockDoors)
	v1Auth.Post("/user/devices/:userDeviceID/integrations/:integrationID/commands/trunk/open", userDeviceController.OpenTrunk)
	v1Auth.Post("/user/devices/:userDeviceID/integrations/:integrationID/commands/frunk/open", userDeviceController.OpenFrunk)

	// Data sharing opt-in.
	// TODO(elffjs): Opt out.
	v1Auth.Post("/user/devices/:userDeviceID/commands/opt-in", userDeviceController.DeviceOptIn).Name("DeviceOptIn")

	v1Auth.Get("/integrations", userDeviceController.GetIntegrations)
	// autopi specific
	v1Auth.Post("/user/devices/:userDeviceID/autopi/command", userDeviceController.SendAutoPiCommand)
	v1Auth.Get("/user/devices/:userDeviceID/autopi/command/:jobID", userDeviceController.GetAutoPiCommandStatus)
	v1Auth.Get("/autopi/unit/:unitID", userDeviceController.GetAutoPiUnitInfo)
	v1Auth.Get("/autopi/unit/:unitID/is-online", userDeviceController.GetIsAutoPiOnline)
	// delete below line once confirmed no active apps using it.
	v1Auth.Get("/autopi/unit/is-online/:unitID", userDeviceController.GetIsAutoPiOnline) // this one is deprecated
	v1Auth.Post("/autopi/unit/:unitID/update", userDeviceController.StartAutoPiUpdateTask)
	v1Auth.Get("/autopi/task/:taskID", userDeviceController.GetAutoPiTask)

	// New-style NFT mint, claim, pair.
	v1Auth.Get("/user/devices/:userDeviceID/commands/mint", userDeviceController.GetMintDevice)
	v1Auth.Post("/user/devices/:userDeviceID/commands/mint", userDeviceController.PostMintDevice).Name("PostMintDevice")
	v1Auth.Post("/user/devices/:userDeviceID/commands/update-nft-image", userDeviceController.UpdateNFTImage)

	v1Auth.Get("/autopi/unit/:unitID/commands/claim", userDeviceController.GetAutoPiClaimMessage)
	v1Auth.Post("/autopi/unit/:unitID/commands/claim", userDeviceController.PostClaimAutoPi).Name("PostClaimAutoPi")
	if !settings.IsProduction() {
		v1Auth.Post("/autopi/unit/:unitID/commands/unclaim", userDeviceController.PostUnclaimAutoPi)
	}

	v1Auth.Get("/user/devices/:userDeviceID/autopi/commands/pair", userDeviceController.GetAutoPiPairMessage)
	v1Auth.Post("/user/devices/:userDeviceID/autopi/commands/pair", userDeviceController.PostPairAutoPi).Name("PostPairAutoPi")

	v1Auth.Get("/user/devices/:userDeviceID/autopi/commands/unpair", userDeviceController.GetAutoPiUnpairMessage)
	v1Auth.Post("/user/devices/:userDeviceID/autopi/commands/unpair", userDeviceController.UnpairAutoPi)

	v1Auth.Post("/user/devices/:userDeviceID/autopi/commands/cloud-repair", userDeviceController.CloudRepairAutoPi)

	// geofence
	v1Auth.Post("/user/geofences", geofenceController.Create)
	v1Auth.Get("/user/geofences", geofenceController.GetAll)
	v1Auth.Delete("/user/geofences/:geofenceID", geofenceController.Delete)
	v1Auth.Put("/user/geofences/:geofenceID", geofenceController.Update)

	// documents
	v1Auth.Get("/documents", documentsController.GetDocuments)
	v1Auth.Get("/documents/:id", documentsController.GetDocumentByID)
	v1Auth.Post("/documents", documentsController.PostDocument)
	v1Auth.Delete("/documents/:id", documentsController.DeleteDocument)
	v1Auth.Get("/documents/:id/download", documentsController.DownloadDocument)

	go startGRPCServer(settings, pdb.DBS, hardwareTemplateService, &logger, ddSvc, eventService)

	logger.Info().Msg("Server started on port " + settings.Port)
	// Start Server from a different go routine
	go func() {
		if err := app.Listen(":" + settings.Port); err != nil {
			logger.Fatal().Err(err)
		}
	}()
	// start kafka consumer for registry processor
	kconf := sarama.NewConfig()
	kconf.Version = sarama.V2_8_1_0

	kclient, err := sarama.NewClient(strings.Split(settings.KafkaBrokers, ","), kconf)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Sarama client")
	}

	store, err := registry.NewProcessor(pdb.DBS, &logger, autoPi)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create registry storage client")
	}

	ctx := context.Background()
	err = registry.RunConsumer(ctx, kclient, &logger, store)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create transaction listener")
	}
	// start task consumer for autopi
	autoPiTaskService.StartConsumer(ctx)
	drivlyTaskService.StartConsumer(ctx)

	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent with length of 1
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel
	<-c                                             // This blocks the main thread until an interrupt is received
	logger.Info().Msg("Gracefully shutting down and running cleanup tasks...")
	_ = ctx.Done()
	_ = app.Shutdown()
	_ = pdb.DBS().Writer.Close()
	_ = pdb.DBS().Reader.Close()
	_ = producer.Close()
}

func healthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	err := c.JSON(res)

	if err != nil {
		return err
	}

	return nil
}

func startGRPCServer(settings *config.Settings, dbs func() *db.ReaderWriter,
	hardwareTemplateService autopi.HardwareTemplateService, logger *zerolog.Logger, deviceDefSvc services.DeviceDefinitionService, eventService services.EventService) {
	lis, err := net.Listen("tcp", ":"+settings.GRPCPort)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Couldn't listen on gRPC port %s", settings.GRPCPort)
	}

	logger.Info().Msgf("Starting gRPC server on port %s", settings.GRPCPort)
	server := grpc.NewServer()
	pb.RegisterUserDeviceServiceServer(server, api.NewUserDeviceService(dbs, settings, hardwareTemplateService, logger, deviceDefSvc, eventService))
	pb.RegisterAftermarketDeviceServiceServer(server, api.NewAftermarketDeviceService(dbs, logger))

	if err := server.Serve(lis); err != nil {
		logger.Fatal().Err(err).Msg("gRPC server terminated unexpectedly")
	}
}
