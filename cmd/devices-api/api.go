package main

import (
	"context"
	"math/big"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
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
	"github.com/DIMO-Network/devices-api/internal/services/integration"
	"github.com/DIMO-Network/devices-api/internal/services/ipfs"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/internal/services/tmpcred"
	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	cip "github.com/DIMO-Network/shared/pkg/cipher"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/shared/pkg/middleware/privilegetoken"
	"github.com/DIMO-Network/shared/pkg/privileges"
	"github.com/DIMO-Network/shared/pkg/redis"
	pb_oracle "github.com/DIMO-Network/tesla-oracle/pkg/grpc"
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

func startWebAPI(logger zerolog.Logger, settings *config.Settings, pdb db.Store, producer sarama.SyncProducer, s3ServiceClient *s3.Client) {
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

	var cipher cip.Cipher
	if settings.Environment == "dev" || settings.IsProduction() {
		cipher = createKMS(settings, &logger)
	} else {
		logger.Warn().Msg("Using ROT13 encrypter. Only use this for testing!")
		cipher = new(cip.ROT13Cipher)
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

	oracleConn, err := grpc.NewClient(settings.TeslaOracleGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed dialing tesla-oracle.")
	}
	teslaOracle := pb_oracle.NewTeslaOracleClient(oracleConn)

	// services
	ddIntSvc := services.NewDeviceDefinitionIntegrationService(pdb.DBS, settings)
	ddSvc := services.NewDeviceDefinitionService(pdb.DBS, &logger, settings)
	ipfsSvc, err := ipfs.NewGateway(settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error creating IPFS client.")
	}
	teslaTaskService := services.NewTeslaTaskService(settings, producer)
	teslaFleetAPISvc, err := services.NewTeslaFleetAPIService(settings, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error constructing Tesla Fleet API client.")
	}
	autoPiSvc := services.NewAutoPiAPIService(settings, pdb.DBS)
	autoPiIngest := services.NewIngestRegistrar(producer)
	hardwareTemplateService := autopi.NewHardwareTemplateService(autoPiSvc, pdb.DBS, &logger)
	userDeviceSvc := services.NewUserDeviceService(ddSvc, logger, pdb.DBS)

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
	userDeviceController := controllers.NewUserDevicesController(settings, pdb.DBS, &logger, ddSvc, ddIntSvc,
		teslaTaskService, teslaOracle, cipher, autoPiSvc, autoPiIngest,
		producer, redisCache, openAI,
		natsSvc, wallet, userDeviceSvc, teslaFleetAPISvc, ipfsSvc, chConn)
	webhooksController := controllers.NewWebhooksController(settings, pdb.DBS, &logger, autoPiSvc, ddIntSvc)
	documentsController := controllers.NewDocumentsController(settings, &logger, s3ServiceClient, pdb.DBS)
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
	nftController := controllers.NewNFTController(settings, pdb.DBS, &logger, ddSvc, teslaTaskService, ddIntSvc, teslaOracle)

	// webhooks, performs signature validation
	v1.Post(constants.AutoPiWebhookPath, webhooksController.ProcessCommand)

	privilegeAuth := jwtware.New(jwtware.Config{
		JWKSetURLs: []string{settings.TokenExchangeJWTKeySetURL},
	})

	newNFTHost, err := url.ParseRequestURI(settings.NewNFTHost)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't parse the new NFT host.")
	}

	app.Get("/v1/vehicle/:tokenID", func(c *fiber.Ctx) error {
		tokenID, err := c.ParamsInt("tokenID")
		if err != nil || tokenID <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid token id.")
		}

		return c.Redirect(newNFTHost.JoinPath("vehicle", strconv.Itoa(tokenID)).String(), fiber.StatusMovedPermanently)
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
	vPriv.Post("/commands/charge/start", privTokenWare.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleCommands}), nftController.ChargeStart)
	vPriv.Post("/commands/charge/stop", privTokenWare.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleCommands}), nftController.ChargeStop)

	// Vehicle owner routes.
	vPriv.Get("/error-codes", privTokenWare.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleNonLocationData}), userDeviceController.GetUserDeviceErrorCodeQueriesByTokenID)
	vPriv.Post("/error-codes", privTokenWare.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleNonLocationData}), userDeviceController.QueryDeviceErrorCodesByTokenID)
	vPriv.Post("/error-codes/clear", privTokenWare.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleNonLocationData}), userDeviceController.ClearUserDeviceErrorCodeQueryByTokenID)

	// Traditional tokens

	jwtAuth := jwtware.New(jwtware.Config{
		JWKSetURLs: []string{settings.JwtKeySetURL},
	})

	v1Auth := app.Group("/v1", jwtAuth)

	// List user's devices.
	v1Auth.Get("/user/devices/me", userDeviceController.GetUserDevices)

	// Device creation.
	v1Auth.Post("/user/devices", userDeviceController.RegisterDeviceForUser)

	// documents
	v1Auth.Get("/documents", documentsController.GetDocuments)
	v1Auth.Get("/documents/:id", documentsController.GetDocumentByID)
	v1Auth.Post("/documents", documentsController.PostDocument)
	v1Auth.Delete("/documents/:id", documentsController.DeleteDocument)
	v1Auth.Get("/documents/:id/download", documentsController.DownloadDocument)

	// Vehicle owner routes.
	udOwnerMw := owner.UserDevice(pdb, &logger)
	udOwner := v1Auth.Group("/user/devices/:userDeviceID", udOwnerMw)

	udOwner.Delete("/", userDeviceController.DeleteUserDevice)
	udOwner.Get("/commands/mint", userDeviceController.GetMintDevice)
	udOwner.Post("/commands/mint", userDeviceController.PostMintDevice)

	udOwner.Post("/error-codes", userDeviceController.QueryDeviceErrorCodes)
	udOwner.Get("/error-codes", userDeviceController.GetUserDeviceErrorCodeQueries)
	udOwner.Post("/error-codes/clear", userDeviceController.ClearUserDeviceErrorCodeQuery)

	// device integrations
	udOwner.Get("/integrations/:integrationID", userDeviceController.GetUserDeviceIntegration)
	udOwner.Delete("/integrations/:integrationID", userDeviceController.DeleteUserDeviceIntegration)
	udOwner.Post("/integrations/:integrationID", userDeviceController.RegisterDeviceIntegration)

	{
		addr := address.New(&logger)

		v1Auth.Post("/integration/:tokenID/credentials", addr, userIntegrationAuthController.CompleteOAuthExchange)

		sdc := sd.Controller{
			DBS:         pdb,
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

	syntheticController := controllers.NewSyntheticDevicesController(settings, pdb.DBS, &logger, ddSvc, wallet, registryClient, teslaOracle)

	udOwner.Get("/integrations/:integrationID/commands/mint", syntheticController.GetSyntheticDeviceMintingPayload)
	udOwner.Post("/integrations/:integrationID/commands/mint", syntheticController.MintSyntheticDevice)

	udOwner.Get("/integrations/:integrationID/commands/burn", syntheticController.GetSyntheticDeviceBurnPayload)
	udOwner.Post("/integrations/:integrationID/commands/burn", syntheticController.BurnSyntheticDevice)

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

	store, err := registry.NewProcessor(pdb.DBS, &logger, settings, teslaTaskService, ddSvc)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create registry storage client")
	}

	if err := registry.RunConsumer(ctx, kclient, &logger, store); err != nil {
		logger.Fatal().Err(err).Msg("Failed to create transaction listener")
	}

	go startGRPCServer(settings, pdb.DBS, hardwareTemplateService, &logger, ddSvc, userDeviceSvc, teslaTaskService, cipher, teslaFleetAPISvc, producer)

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
	userDeviceSvc services.UserDeviceService,
	teslaTaskSvc services.TeslaTaskService,
	cipher cip.Cipher,
	teslaAPI services.TeslaFleetAPIService,
	producer sarama.SyncProducer,
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
		deviceDefSvc, userDeviceSvc, teslaTaskSvc))
	pb.RegisterAftermarketDeviceServiceServer(server, rpc.NewAftermarketDeviceService(dbs, logger))
	pb.RegisterTeslaServiceServer(server, rpc.NewTeslaRPCService(dbs, settings, cipher, teslaAPI, logger, producer))

	if err := server.Serve(lis); err != nil {
		logger.Fatal().Err(err).Msg("gRPC server terminated unexpectedly")
	}
}
