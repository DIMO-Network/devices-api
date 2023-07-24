package main

import (
	"context"
	"encoding/base64"
	"math/big"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/DIMO-Network/devices-api/internal/rpc"

	"github.com/DIMO-Network/devices-api/internal/middleware/metrics"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/DIMO-Network/shared/redis"

	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"

	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/middleware/owner"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/autopi"
	"github.com/DIMO-Network/devices-api/internal/services/fingerprint"
	"github.com/DIMO-Network/devices-api/internal/services/issuer"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	pbuser "github.com/DIMO-Network/shared/api/users"
	pr "github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/DIMO-Network/zflogger"
	"github.com/Shopify/sarama"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
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

	gcon, err := grpc.Dial(settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed dialing users-api.")
	}
	usersClient := pbuser.NewUserServiceClient(gcon)

	// services
	nhtsaSvc := services.NewNHTSAService()
	ddIntSvc := services.NewDeviceDefinitionIntegrationService(pdb.DBS, settings)
	ddSvc := services.NewDeviceDefinitionService(pdb.DBS, &logger, nhtsaSvc, settings)
	ddaSvc := services.NewDeviceDataService(settings.DeviceDataGRPCAddr, &logger)

	scTaskSvc := services.NewSmartcarTaskService(settings, producer)
	smartcarClient := services.NewSmartcarClient(settings)
	teslaTaskService := services.NewTeslaTaskService(settings, producer)
	teslaSvc := services.NewTeslaService(settings)
	autoPiSvc := services.NewAutoPiAPIService(settings, pdb.DBS)
	autoPiIngest := services.NewIngestRegistrar(producer)
	deviceDefinitionRegistrar := services.NewDeviceDefinitionRegistrar(producer, settings)
	autoPiTaskService := services.NewAutoPiTaskService(settings, autoPiSvc, pdb.DBS, logger)
	hardwareTemplateService := autopi.NewHardwareTemplateService(autoPiSvc, pdb.DBS, &logger)
	autoPi := autopi.NewIntegration(pdb.DBS, ddSvc, autoPiSvc, autoPiTaskService, autoPiIngest, eventService, deviceDefinitionRegistrar, hardwareTemplateService, &logger)
	openAI := services.NewOpenAI(&logger, *settings)
	dcnSvc := registry.NewDcnService(settings)

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

	// controllers
	userDeviceController := controllers.NewUserDevicesController(settings, pdb.DBS, &logger, ddSvc, ddIntSvc, eventService, smartcarClient, scTaskSvc, teslaSvc, teslaTaskService, cipher, autoPiSvc, services.NewNHTSAService(), autoPiIngest, deviceDefinitionRegistrar, autoPiTaskService, producer, s3NFTServiceClient, autoPi, redisCache, openAI, usersClient, ddaSvc, natsSvc)
	geofenceController := controllers.NewGeofencesController(settings, pdb.DBS, &logger, producer, ddSvc)
	webhooksController := controllers.NewWebhooksController(settings, pdb.DBS, &logger, autoPiSvc, ddIntSvc)
	documentsController := controllers.NewDocumentsController(settings, &logger, s3ServiceClient, pdb.DBS)

	// commenting this out b/c the library includes the path in the metrics which saturates prometheus queries - need to fork / make our own
	//prometheus := fiberprometheus.New("devices-api")
	//app.Use(prometheus.Middleware)
	app.Use(metrics.HTTPMetricsMiddleware)

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
	nftController := controllers.NewNFTController(settings, pdb.DBS, &logger, s3NFTServiceClient, ddSvc, scTaskSvc, teslaTaskService, ddIntSvc, dcnSvc, ddaSvc)
	v1.Get("/vehicle/:tokenID", nftController.GetNFTMetadata)
	v1.Get("/vehicle/:tokenID/image", nftController.GetNFTImage)

	v1.Get("/aftermarket/device/:tokenID", cacheHandler, nftController.GetAftermarketDeviceNFTMetadata)
	v1.Get("/aftermarket/device/:tokenID/image", nftController.GetAftermarketDeviceNFTImage)
	v1.Get("/manufacturer/:tokenID", nftController.GetManufacturerNFTMetadata)

	v1.Get("/dcn/:nodeID", nftController.GetDcnNFTMetadata)
	v1.Get("/dcn/:nodeID/image", nftController.GetDCNNFTImage)
	v1.Get("/integration/:tokenID", nftController.GetIntegrationNFTMetadata)

	// webhooks, performs signature validation
	v1.Post(constants.AutoPiWebhookPath, webhooksController.ProcessCommand)

	privilegeAuth := jwtware.New(jwtware.Config{
		JWKSetURLs: []string{settings.TokenExchangeJWTKeySetURL},
	})

	vPriv := app.Group("/v1/vehicle/:tokenID", privilegeAuth)

	tk := pr.New(pr.Config{
		Log: &logger,
	})

	vehicleAddr := common.HexToAddress(settings.VehicleNFTAddress)

	// vehicle command privileges
	vPriv.Get("/status", tk.OneOf(vehicleAddr, []int64{controllers.NonLocationData, controllers.CurrentLocation, controllers.AllTimeLocation}), nftController.GetVehicleStatus)
	if !settings.IsProduction() {
		vPriv.Get("/vin-credential", tk.OneOf(vehicleAddr, []int64{controllers.VinCredential}), nftController.GetVinCredential)
	}
	vPriv.Post("/commands/doors/unlock", tk.OneOf(vehicleAddr, []int64{controllers.Commands}), nftController.UnlockDoors)
	vPriv.Post("/commands/doors/lock", tk.OneOf(vehicleAddr, []int64{controllers.Commands}), nftController.LockDoors)
	vPriv.Post("/commands/trunk/open", tk.OneOf(vehicleAddr, []int64{controllers.Commands}), nftController.OpenTrunk)
	vPriv.Post("/commands/frunk/open", tk.OneOf(vehicleAddr, []int64{controllers.Commands}), nftController.OpenFrunk)

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

	v1Auth.Get("/integrations", userDeviceController.GetIntegrations)

	// Autopi specific routes.
	amdOwnerMw := owner.AftermarketDevice(pdb, usersClient, &logger)
	apOwner := v1Auth.Group("/autopi/unit/:serial", amdOwnerMw)
	// same as above but AftermarketDevice
	amdOwner := v1Auth.Group("/aftermarket/device/by-serial/:serial", amdOwnerMw)

	apOwner.Get("/", userDeviceController.GetAutoPiUnitInfo)
	amdOwner.Get("/", userDeviceController.GetAutoPiUnitInfo)

	apOwner.Post("/update", userDeviceController.StartAutoPiUpdateTask)
	amdOwner.Post("/update", userDeviceController.StartAutoPiUpdateTask)

	// AftermarketDevice claiming, formerly AutoPi
	apOwner.Get("/commands/claim", userDeviceController.GetAutoPiClaimMessage)
	amdOwner.Get("/commands/claim", userDeviceController.GetAutoPiClaimMessage)

	apOwner.Post("/commands/claim", userDeviceController.PostClaimAutoPi).Name("PostClaimAutoPi")
	amdOwner.Post("/commands/claim", userDeviceController.PostClaimAutoPi).Name("PostClaimAutoPi")
	if !settings.IsProduction() {
		// Used by mobile to test. Easy to misuse.
		apOwner.Post("/commands/unclaim", userDeviceController.PostUnclaimAutoPi)
		amdOwner.Post("/commands/unclaim", userDeviceController.PostUnclaimAutoPi)
	}

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

	if settings.SyntheticDevicesEnabled {
		syntheticDeviceSvc, err := services.NewSyntheticWalletInstanceService(pdb.DBS, settings)
		if err != nil {
			logger.Error().Err(err).Msg("unable to create Synthetic Device service")
		}

		syntheticController := controllers.NewSyntheticDevicesController(settings, pdb.DBS, &logger, ddSvc, usersClient, syntheticDeviceSvc, registryClient, smartcarClient, teslaSvc, cipher)

		sdAuth := v1Auth.Group("/synthetic/device")

		sdAuth.Get("/mint/:integrationNode/:vehicleNode", syntheticController.GetSyntheticDeviceMintingPayload)
		sdAuth.Post("/mint/:integrationNode/:vehicleNode", syntheticController.MintSyntheticDevice)

		sdAuth.Get("/:syntheticDeviceNode/burn", syntheticController.GetSyntheticDeviceBurnPayload)
		sdAuth.Post("/:syntheticDeviceNode/burn", syntheticController.BurnSyntheticDevice)

		sdAuth.Post("/:syntheticDeviceNode/re-authenticate", syntheticController.ReAuthenticate)
	}

	// Vehicle owner routes.
	udOwnerMw := owner.UserDevice(pdb, usersClient, &logger)
	udOwner := v1Auth.Group("/user/devices/:userDeviceID", udOwnerMw)

	udOwner.Get("/status", userDeviceController.GetUserDeviceStatus)
	udOwner.Delete("/", userDeviceController.DeleteUserDevice)
	udOwner.Get("/commands/mint", userDeviceController.GetMintDevice)
	udOwner.Post("/commands/mint", userDeviceController.PostMintDevice)

	udOwner.Patch("/vin", userDeviceController.UpdateVIN)
	udOwner.Patch("/name", userDeviceController.UpdateName)
	udOwner.Patch("/country-code", userDeviceController.UpdateCountryCode)
	udOwner.Get("/valuations", userDeviceController.GetValuations)
	udOwner.Get("/offers", userDeviceController.GetOffers)
	udOwner.Get("/range", userDeviceController.GetRange)

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

	// Vehicle commands.
	udOwner.Post("/integrations/:integrationID/commands/doors/unlock", userDeviceController.UnlockDoors)
	udOwner.Post("/integrations/:integrationID/commands/doors/lock", userDeviceController.LockDoors)
	udOwner.Post("/integrations/:integrationID/commands/trunk/open", userDeviceController.OpenTrunk)
	udOwner.Post("/integrations/:integrationID/commands/frunk/open", userDeviceController.OpenFrunk)
	udOwner.Get("/integrations/:integrationID/commands/:requestID", userDeviceController.GetCommandRequestStatus)

	udOwner.Post("/commands/opt-in", userDeviceController.DeviceOptIn)

	// AftermarketDevice pairing and unpairing.
	udOwner.Get("/autopi/commands/pair", userDeviceController.GetAutoPiPairMessage)
	udOwner.Get("/aftermarket/commands/pair", userDeviceController.GetAutoPiPairMessage)
	udOwner.Post("/autopi/commands/pair", userDeviceController.PostPairAutoPi)
	udOwner.Post("/aftermarket/commands/pair", userDeviceController.PostPairAutoPi)
	udOwner.Get("/autopi/commands/unpair", userDeviceController.GetAutoPiUnpairMessage)
	udOwner.Get("/aftermarket/commands/unpair", userDeviceController.GetAutoPiUnpairMessage)
	udOwner.Post("/autopi/commands/unpair", userDeviceController.UnpairAutoPi)
	udOwner.Post("/aftermarket/commands/unpair", userDeviceController.UnpairAutoPi)

	udOwner.Post("/autopi/commands/cloud-repair", userDeviceController.CloudRepairAutoPi)
	udOwner.Post("/aftermarket/commands/cloud-repair", userDeviceController.CloudRepairAutoPi)

	go startValuationConsumer(settings, pdb.DBS, &logger, ddSvc, natsSvc)

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

	iss, err := createVCIssuer(settings, pdb, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create issuer.")
	}

	if err := fingerprint.RunConsumer(ctx, settings, &logger, iss, pdb); err != nil {
		logger.Fatal().Err(err).Msg("Failed to create vin credentialer listener")
	}

	store, err := registry.NewProcessor(pdb.DBS, &logger, autoPi, settings, scTaskSvc, teslaTaskService, ddSvc)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create registry storage client")
	}

	if err := registry.RunConsumer(ctx, kclient, &logger, store); err != nil {
		logger.Fatal().Err(err).Msg("Failed to create transaction listener")
	}

	go startGRPCServer(settings, pdb.DBS, hardwareTemplateService, &logger, ddSvc, eventService, iss)

	// start task consumer for autopi
	autoPiTaskService.StartConsumer(ctx)

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
	vcIss *issuer.Issuer,
) {
	lis, err := net.Listen("tcp", ":"+settings.GRPCPort)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Couldn't listen on gRPC port %s", settings.GRPCPort)
	}

	logger.Info().Msgf("Starting gRPC server on port %s", settings.GRPCPort)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			metrics.GRPCMetricsMiddleware(),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		)),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	pb.RegisterUserDeviceServiceServer(server, rpc.NewUserDeviceService(dbs, settings, hardwareTemplateService, logger, deviceDefSvc, eventService, vcIss))
	pb.RegisterAftermarketDeviceServiceServer(server, rpc.NewAftermarketDeviceService(dbs, logger))

	if err := server.Serve(lis); err != nil {
		logger.Fatal().Err(err).Msg("gRPC server terminated unexpectedly")
	}
}

func createVCIssuer(settings *config.Settings, dbs db.Store, logger *zerolog.Logger) (*issuer.Issuer, error) {
	pk, err := base64.RawURLEncoding.DecodeString(settings.IssuerPrivateKey)
	if err != nil {
		return nil, err
	}

	return issuer.New(
		issuer.Config{
			PrivateKey:        pk,
			ChainID:           big.NewInt(settings.DIMORegistryChainID),
			VehicleNFTAddress: common.HexToAddress(settings.VehicleNFTAddress),
			DBS:               dbs,
		},
		logger,
	)
}

func startValuationConsumer(settings *config.Settings, pdb func() *db.ReaderWriter, logger *zerolog.Logger, ddSvc services.DeviceDefinitionService, natsSvc *services.NATSService) {
	if settings.IsProduction() {

		valuationService := services.NewValuationService(logger, pdb, ddSvc, natsSvc)

		go func() {
			err := valuationService.ValuationConsumer(context.Background())

			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to start valuation consumer")
			}
		}()
	}
}
