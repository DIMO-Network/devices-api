package main

import (
	"context"
	"math/big"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/DIMO-Network/clickhouse-infra/pkg/connect"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/controllers/user/sd"
	"github.com/DIMO-Network/devices-api/internal/middleware"
	"github.com/DIMO-Network/devices-api/internal/middleware/address"
	"github.com/DIMO-Network/devices-api/internal/middleware/metrics"
	"github.com/DIMO-Network/devices-api/internal/middleware/owner"
	"github.com/DIMO-Network/devices-api/internal/rpc"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/autopi"
	"github.com/DIMO-Network/devices-api/internal/services/fingerprint"
	"github.com/DIMO-Network/devices-api/internal/services/genericad"
	"github.com/DIMO-Network/devices-api/internal/services/integration"
	"github.com/DIMO-Network/devices-api/internal/services/ipfs"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/internal/services/tmpcred"
	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	pbuser "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/DIMO-Network/shared/privileges"
	"github.com/DIMO-Network/shared/redis"
	"github.com/IBM/sarama"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func startWebAPI(logger zerolog.Logger, settings *config.Settings, pdb db.Store, eventService services.EventService, producer sarama.SyncProducer, s3ServiceClient *s3.Client, s3NFTServiceClient *s3.Client) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return helpers.ErrorHandler(c, err, &logger, settings.IsProduction())
		},
		DisableStartupMessage: true,
		ReadBufferSize:        16000,
		BodyLimit:             10 * 1024 * 1024,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
	})

	var cipher shared.Cipher
	if settings.Environment == "dev" || settings.IsProduction() {
		cipher = createKMS(settings, &logger)
	} else {
		logger.Warn().Msg("Using ROT13 encrypter. Only use this for testing!")
		cipher = new(shared.ROT13Cipher)
	}

	registryClient := registry.Client{
		Producer:     producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(settings.DIMORegistryChainID),
			Address: common.HexToAddress(settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	gcon, err := grpc.NewClient(settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed dialing users-api.")
	}
	usersClient := pbuser.NewUserServiceClient(gcon)

	// services
	ddIntSvc := services.NewDeviceDefinitionIntegrationService(pdb.DBS, settings)
	ddSvc := services.NewDeviceDefinitionService(pdb.DBS, &logger, settings)
	ddaSvc := services.NewDeviceDataService(settings.DeviceDataGRPCAddr, &logger)
	ipfsSvc, err := ipfs.NewGateway(settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error creating IPFS client.")
	}
	scTaskSvc := services.NewSmartcarTaskService(settings, producer)
	smartcarClient := services.NewSmartcarClient(settings)
	teslaTaskService := services.NewTeslaTaskService(settings, producer)
	teslaFleetAPISvc, err := services.NewTeslaFleetAPIService(settings, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error constructing Tesla Fleet API client.")
	}
	autoPiSvc := services.NewAutoPiAPIService(settings, pdb.DBS)
	autoPiIngest := services.NewIngestRegistrar(producer)
	deviceDefinitionRegistrar := services.NewDeviceDefinitionRegistrar(producer, settings)
	hardwareTemplateService := autopi.NewHardwareTemplateService(autoPiSvc, pdb.DBS, &logger)
	genericADIntegration := genericad.NewIntegration(pdb.DBS, ddSvc, autoPiIngest, eventService, deviceDefinitionRegistrar, &logger)
	userDeviceSvc := services.NewUserDeviceService(ddSvc, logger, pdb.DBS, eventService, usersClient)

	openAI := services.NewOpenAI(&logger, *settings)

	natsSvc, err := services.NewNATSService(settings, &logger)
	if err != nil {
		logger.Error().Err(err).Msg("unable to create NATS service")
	}

	redisCache := redis.NewRedisCacheService(settings.IsProduction(), redis.Settings{
		URL:       settings.RedisURL,
		Password:  settings.RedisPassword,
		TLS:       settings.RedisTLS,
		KeyPrefix: "devices-api",
	})

	wallet, err := services.NewSyntheticWalletInstanceService(settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't construct wallet client.")
	}

	chConn, err := connect.GetClickhouseConn(&settings.Clickhouse)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't construct ClickHouse client.")
	}

	err = chConn.Ping(context.Background())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to ping ClickHouse.")
	}

	// controllers
	userDeviceController := controllers.NewUserDevicesController(settings, pdb.DBS, &logger, ddSvc, ddIntSvc, eventService,
		smartcarClient, scTaskSvc, teslaTaskService, cipher, autoPiSvc, autoPiIngest,
		deviceDefinitionRegistrar, producer, s3NFTServiceClient, redisCache, openAI, usersClient,
		ddaSvc, natsSvc, wallet, userDeviceSvc, teslaFleetAPISvc, ipfsSvc, chConn)
	geofenceController := controllers.NewGeofencesController(settings, pdb.DBS, &logger, producer, ddSvc, usersClient)
	webhooksController := controllers.NewWebhooksController(settings, pdb.DBS, &logger, autoPiSvc, ddIntSvc)
	documentsController := controllers.NewDocumentsController(settings, &logger, s3ServiceClient, pdb.DBS)
	countriesController := controllers.NewCountriesController()
	userIntegrationAuthController := controllers.NewUserIntegrationAuthController(settings, pdb.DBS, &logger, ddSvc, teslaFleetAPISvc, &tmpcred.Store{
		Redis:  redisCache,
		Cipher: cipher,
	})

	app.Use(metrics.HTTPMetricsMiddleware)

	app.Use(fiberrecover.New(fiberrecover.Config{
		Next:              nil,
		EnableStackTrace:  true,
		StackTraceHandler: nil,
	}))

	// application routes
	app.Get("/", healthCheck)

	v1 := app.Group("/v1")

	v1.Get("/swagger/*", swagger.HandlerDefault)
	// Device Definitions
	nftController := controllers.NewNFTController(settings, pdb.DBS, &logger, s3NFTServiceClient, ddSvc, scTaskSvc, teslaTaskService, ddIntSvc)

	v1.Get("/countries", countriesController.GetSupportedCountries)
	v1.Get("/countries/:countryCode", countriesController.GetCountry)

	// webhooks, performs signature validation
	v1.Post(constants.AutoPiWebhookPath, webhooksController.ProcessCommand)

	privilegeAuth := jwtware.New(jwtware.Config{
		JWKSetURLs: []string{settings.TokenExchangeJWTKeySetURL},
	})

	vPriv := app.Group("/v1/vehicle/:tokenID", privilegeAuth)

	privTokenWare := privilegetoken.New(privilegetoken.Config{Log: &logger})

	vehicleAddr := common.HexToAddress(settings.VehicleNFTAddress)

	// vehicle command privileges
	vPriv.Patch("/vin", privTokenWare.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleCommands}), userDeviceController.UpdateVINV2)
	vPriv.Post("/commands/doors/unlock", privTokenWare.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleCommands}), nftController.UnlockDoors)
	vPriv.Post("/commands/doors/lock", privTokenWare.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleCommands}), nftController.LockDoors)
	vPriv.Post("/commands/trunk/open", privTokenWare.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleCommands}), nftController.OpenTrunk)
	vPriv.Post("/commands/frunk/open", privTokenWare.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleCommands}), nftController.OpenFrunk)

	// Traditional tokens

	jwtAuth := jwtware.New(jwtware.Config{
		JWKSetURLs: []string{settings.JwtKeySetURL},
	})

	v1Auth := app.Group("/v1", jwtAuth)

	// List user's devices.
	v1Auth.Get("/user/devices/me", userDeviceController.GetUserDevices)
	v1Auth.Get("/user/devices/shared", userDeviceController.GetSharedDevices)

	// Device creation.
	v1Auth.Post("/user/devices/fromvin", userDeviceController.RegisterDeviceForUserFromVIN)
	v1Auth.Post("/user/devices/fromsmartcar", userDeviceController.RegisterDeviceForUserFromSmartcar)
	v1Auth.Post("/user/devices", userDeviceController.RegisterDeviceForUser)

	// Autopi specific routes.
	amdOwnerMw := owner.AftermarketDevice(pdb, usersClient, &logger)
	// same as above but AftermarketDevice
	amdOwner := v1Auth.Group("/aftermarket/device/by-serial/:serial", amdOwnerMw)

	amdOwner.Get("/", userDeviceController.GetAftermarketDeviceInfo)

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

	// Vehicle owner routes.
	udOwnerMw := owner.UserDevice(pdb, usersClient, &logger)
	udOwner := v1Auth.Group("/user/devices/:userDeviceID", udOwnerMw)

	udOwner.Delete("/", userDeviceController.DeleteUserDevice)
	udOwner.Get("/commands/mint", userDeviceController.GetMintDevice)
	udOwner.Post("/commands/mint", userDeviceController.PostMintDevice)

	udOwner.Patch("/vin", userDeviceController.UpdateVIN)
	udOwner.Patch("/country-code", userDeviceController.UpdateCountryCode)

	udOwner.Post("/error-codes", userDeviceController.QueryDeviceErrorCodes)
	udOwner.Get("/error-codes", userDeviceController.GetUserDeviceErrorCodeQueries)
	udOwner.Post("/error-codes/clear", userDeviceController.ClearUserDeviceErrorCodeQuery)

	// New-style NFT mint, claim, pair.
	udOwner.Post("/commands/update-nft-image", userDeviceController.UpdateNFTImage)

	// device integrations
	udOwner.Get("/integrations/:integrationID", userDeviceController.GetUserDeviceIntegration)
	udOwner.Delete("/integrations/:integrationID", userDeviceController.DeleteUserDeviceIntegration)
	udOwner.Post("/integrations/:integrationID", userDeviceController.RegisterDeviceIntegration)
	udOwner.Post("/commands/refresh", userDeviceController.RefreshUserDeviceStatus)

	// Vehicle owner routes.
	vehicleOwnerMw := owner.VehicleToken(pdb, usersClient, &logger)
	{
		addr := address.New(usersClient, &logger)

		v1Auth.Post("/integration/:tokenID/credentials", addr, userIntegrationAuthController.CompleteOAuthExchange)

		sdc := sd.Controller{
			DBS:         pdb,
			Smartcar:    scTaskSvc,
			Tesla:       teslaTaskService,
			IntegClient: &integration.Client{Service: ddSvc},
			Store: &tmpcred.Store{
				Redis:  redisCache,
				Cipher: cipher,
			},
			TeslaAPI: teslaFleetAPISvc,
			Cipher:   cipher,
		}

		v1Auth.Post("/user/synthetic/device/:tokenID/commands/reauthenticate", addr, sdc.PostReauthenticate)
	}

	vOwner := v1Auth.Group("/user/vehicle/:tokenID", vehicleOwnerMw)
	vOwner.Get("/commands/burn", userDeviceController.GetBurnDevice)
	vOwner.Post("/commands/burn", userDeviceController.PostBurnDevice)

	syntheticController := controllers.NewSyntheticDevicesController(settings, pdb.DBS, &logger, ddSvc, usersClient, wallet, registryClient)

	udOwner.Get("/integrations/:integrationID/commands/mint", syntheticController.GetSyntheticDeviceMintingPayload)
	udOwner.Post("/integrations/:integrationID/commands/mint", syntheticController.MintSyntheticDevice)

	udOwner.Get("/integrations/:integrationID/commands/burn", syntheticController.GetSyntheticDeviceBurnPayload)
	udOwner.Post("/integrations/:integrationID/commands/burn", syntheticController.BurnSyntheticDevice)

	// Vehicle commands.
	udOwner.Post("/integrations/:integrationID/commands/doors/unlock", userDeviceController.UnlockDoors)
	udOwner.Post("/integrations/:integrationID/commands/doors/lock", userDeviceController.LockDoors)
	udOwner.Post("/integrations/:integrationID/commands/trunk/open", userDeviceController.OpenTrunk)
	udOwner.Post("/integrations/:integrationID/commands/frunk/open", userDeviceController.OpenFrunk)
	udOwner.Get("/integrations/:integrationID/commands/:requestID", userDeviceController.GetCommandRequestStatus)

	if !settings.IsProduction() {
		udOwner.Post("/integrations/:integrationID/commands/telemetry/subscribe", userDeviceController.TelemetrySubscribe)
	}

	udOwner.Post("/commands/opt-in", userDeviceController.DeviceOptIn)

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

	ctx := context.Background()

	if err := fingerprint.RunConsumer(ctx, settings, &logger, pdb); err != nil {
		logger.Fatal().Err(err).Msg("Failed to create vin credentialer listener")
	}

	startContractEventsConsumer(logger, settings, pdb, genericADIntegration, ddSvc, eventService, scTaskSvc, teslaTaskService)

	store, err := registry.NewProcessor(pdb.DBS, &logger, settings, eventService, scTaskSvc, teslaTaskService, ddSvc)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create registry storage client")
	}

	if err := registry.RunConsumer(ctx, kclient, &logger, store); err != nil {
		logger.Fatal().Err(err).Msg("Failed to create transaction listener")
	}

	go startGRPCServer(settings, pdb.DBS, hardwareTemplateService, &logger, ddSvc, eventService, userDeviceSvc, teslaTaskService, scTaskSvc)

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

func startGRPCServer(
	settings *config.Settings,
	dbs func() *db.ReaderWriter,
	hardwareTemplateService autopi.HardwareTemplateService,
	logger *zerolog.Logger,
	deviceDefSvc services.DeviceDefinitionService,
	eventService services.EventService,
	userDeviceSvc services.UserDeviceService,
	teslaTaskSvc services.TeslaTaskService,
	smartcarTaskSvc services.SmartcarTaskService,
) {
	lis, err := net.Listen("tcp", ":"+settings.GRPCPort)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Couldn't listen on gRPC port %s", settings.GRPCPort)
	}

	logger.Info().Msgf("Starting gRPC server on port %s", settings.GRPCPort)
	gp := middleware.GRPCPanicker{Logger: logger}
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			metrics.GRPCMetricsMiddleware(),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(gp.GRPCPanicRecoveryHandler)),
		)),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	pb.RegisterUserDeviceServiceServer(server, rpc.NewUserDeviceRPCService(dbs, settings, hardwareTemplateService, logger,
		deviceDefSvc, eventService, userDeviceSvc, teslaTaskSvc, smartcarTaskSvc))
	pb.RegisterAftermarketDeviceServiceServer(server, rpc.NewAftermarketDeviceService(dbs, logger))

	if err := server.Serve(lis); err != nil {
		logger.Fatal().Err(err).Msg("gRPC server terminated unexpectedly")
	}
}
