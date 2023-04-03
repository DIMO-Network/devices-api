package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/subcommands"

	_ "github.com/DIMO-Network/devices-api/docs"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/kafka"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/Shopify/sarama"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/burdiyan/kafkautil"
	"github.com/customerio/go-customerio/v3"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	_ "github.com/lib/pq"
	"github.com/lovoo/goka"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	_ "go.uber.org/automaxprocs"
)

// @title                      DIMO Devices API
// @version                    1.0
// @BasePath                   /v1
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
func main() {
	gitSha1 := os.Getenv("GIT_SHA1")
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Str("git-sha1", gitSha1).
		Logger()

	config.SetupMachineryLogging(&logger)

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("could not load settings")
	}
	level, err := zerolog.ParseLevel(settings.LogLevel)
	if err != nil {
		logger.Fatal().Err(err).Msgf("could not parse LOG_LEVEL: %s", settings.LogLevel)
	}
	zerolog.SetGlobalLevel(level)

	pdb := db.NewDbConnectionFromSettings(ctx, &settings.DB, true)
	// check db ready, this is not ideal btw, the db connection handler would be nicer if it did this.
	totalTime := 0
	for !pdb.IsReady() {
		if totalTime > 30 {
			logger.Fatal().Msg("could not connect to postgres after 30 seconds")
		}
		time.Sleep(time.Second)
		totalTime++
	}

	deps := newDependencyContainer(&settings, logger, pdb.DBS)

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")

	// Run API
	if len(os.Args) == 1 {
		if settings.EnablePrivileges {
			startContractEventsConsumer(logger, &settings, pdb)
		}
		startMonitoringServer(logger, &settings)
		eventService := services.NewEventService(&logger, &settings, deps.getKafkaProducer())
		startDeviceStatusConsumer(logger, &settings, pdb, eventService)
		startCredentialConsumer(logger, &settings, pdb)
		startTaskStatusConsumer(logger, &settings, pdb)
		startWebAPI(logger, &settings, pdb, eventService, deps.getKafkaProducer(), deps.getS3ServiceClient(ctx), deps.getS3NFTServiceClient(ctx))
	} else {

		subcommands.Register(&migrateDBCmd{logger: logger, settings: settings}, "database")
		subcommands.Register(&generateEventCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "events")
		subcommands.Register(&setCommandCompatibilityCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&remakeSmartcarTopicCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&remakeAutoPiTopicCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&remakeFenceTopicCmd{logger: logger, settings: settings, pdb: pdb}, "device integrations")
		subcommands.Register(&remakeDeviceDefinitionTopicsCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&startSmartcarFromRefreshCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&autopiToolsCmd{logger: logger, settings: settings, pdb: pdb}, "device integrations")
		subcommands.Register(&updateStateCmd{logger: logger, settings: settings, pdb: pdb}, "device integrations")
		subcommands.Register(&web2PairCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&autoPiKTableDeleteCmd{logger: logger, container: deps}, "device integrations")

		subcommands.Register(&populateESDDDataCmd{logger: logger, settings: settings, pdb: pdb, esInstance: deps.getElasticSearchService(), ddSvc: deps.getDeviceDefinitionService()}, "populate data")
		subcommands.Register(&populateESRegionDataCmd{logger: logger, settings: settings, pdb: pdb, esInstance: deps.getElasticSearchService(), ddSvc: deps.getDeviceDefinitionService()}, "populate data")
		subcommands.Register(&populateUSAPowertrainCmd{logger: logger, settings: settings, pdb: pdb, nhtsaService: deps.getNHTSAService()}, "populate data")

		subcommands.Register(&stopTaskByKeyCmd{logger: logger, settings: settings}, "tasks")

		subcommands.Register(&loadValuationsCmd{logger: logger, settings: settings, pdb: pdb}, "user devices")
		subcommands.Register(&syncDeviceTemplatesCmd{logger: logger, settings: settings, pdb: pdb}, "user devices")

		subcommands.Register(&autopiClearVINCmd{logger: logger, settings: settings, pdb: pdb}, "autopi")

		flag.Parse()
		os.Exit(int(subcommands.Execute(ctx)))
	}

}

func createKafkaProducer(settings *config.Settings) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_1_0
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = kafkautil.NewJVMCompatiblePartitioner
	p, err := sarama.NewSyncProducer(strings.Split(settings.KafkaBrokers, ","), config)
	if err != nil {
		return nil, fmt.Errorf("failed to construct producer with broker list %s: %w", settings.KafkaBrokers, err)
	}
	return p, nil
}

func createKMS(settings *config.Settings, logger *zerolog.Logger) shared.Cipher {
	// Need AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY to be set.
	// TODO(elffjs): Can we let the SDK grab the region too?
	awscfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(settings.AWSRegion))
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create AWS config.")
	}

	return &shared.KMSCipher{
		KeyID:  settings.KMSKeyID,
		Client: kms.NewFromConfig(awscfg),
	}
}

func changeLogLevel(c *fiber.Ctx) error {
	payload := struct {
		LogLevel string `json:"logLevel"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return err
	}
	level, err := zerolog.ParseLevel(payload.LogLevel)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(level)
	return c.Status(fiber.StatusOK).SendString("log level set to: " + level.String())
}

func startDeviceStatusConsumer(logger zerolog.Logger, settings *config.Settings, pdb db.Store, eventService services.EventService) {
	nhtsaSvc := services.NewNHTSAService()
	ddSvc := services.NewDeviceDefinitionService(pdb.DBS, &logger, nhtsaSvc, settings)
	autoPISvc := services.NewAutoPiAPIService(settings, pdb.DBS)
	ingestSvc := services.NewDeviceStatusIngestService(pdb.DBS, &logger, eventService, ddSvc, autoPISvc)

	sc := goka.DefaultConfig()
	sc.Version = sarama.V2_8_1_0
	goka.ReplaceGlobalConfig(sc)

	group := goka.DefineGroup("devices-vin-fraud",
		goka.Input(goka.Stream(settings.DeviceStatusTopic), new(shared.JSONCodec[services.DeviceStatusEvent]), ingestSvc.ProcessDeviceStatusMessages),
		goka.Persist(new(shared.JSONCodec[shared.CloudEvent[services.RegisteredVIN]])),
	)

	processor, err := goka.NewProcessor(strings.Split(settings.KafkaBrokers, ","),
		group,
		goka.WithHasher(kafkautil.MurmurHasher),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not start device status processor")
	}

	go func() {
		err = processor.Run(context.Background())
		if err != nil {
			logger.Fatal().Err(err).Msg("could not run device status processor")
		}
	}()

	logger.Info().Msg("Device status update consumer started")
}

func startCredentialConsumer(logger zerolog.Logger, settings *config.Settings, pdb db.Store) {
	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = sarama.V2_8_1_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	cfg := &kafka.Config{
		ClusterConfig:   clusterConfig,
		BrokerAddresses: strings.Split(settings.KafkaBrokers, ","),
		Topic:           settings.TaskCredentialTopic,
		GroupID:         "user-devicesYY",
		MaxInFlight:     int64(5),
	}
	consumer, err := kafka.NewConsumer(cfg, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not start credential update consumer")
	}
	credService := services.NewCredentialListener(pdb.DBS, &logger)
	consumer.Start(context.Background(), credService.ProcessCredentialsMessages)

	logger.Info().Msg("Credential update consumer started")
}

func startTaskStatusConsumer(logger zerolog.Logger, settings *config.Settings, pdb db.Store) {
	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = sarama.V2_8_1_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	cfg := &kafka.Config{
		ClusterConfig:   clusterConfig,
		BrokerAddresses: strings.Split(settings.KafkaBrokers, ","),
		Topic:           settings.TaskStatusTopic,
		GroupID:         "user-devices",
		MaxInFlight:     int64(5),
	}
	consumer, err := kafka.NewConsumer(cfg, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not start credential update consumer")
	}
	cio := customerio.NewTrackClient(
		settings.CIOSiteID,
		settings.CIOApiKey,
		customerio.WithRegion(customerio.RegionUS),
	)

	nhtsaSvc := services.NewNHTSAService()
	ddSvc := services.NewDeviceDefinitionService(pdb.DBS, &logger, nhtsaSvc, settings)

	taskStatusService := services.NewTaskStatusListener(pdb.DBS, &logger, cio, ddSvc)
	consumer.Start(context.Background(), taskStatusService.ProcessTaskUpdates)

	logger.Info().Msg("Task status consumer started")
}

func startContractEventsConsumer(logger zerolog.Logger, settings *config.Settings, pdb db.Store) {
	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = sarama.V2_8_1_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	cfg := &kafka.Config{
		ClusterConfig:   clusterConfig,
		BrokerAddresses: strings.Split(settings.KafkaBrokers, ","),
		Topic:           settings.ContractsEventTopic,
		GroupID:         "user-devices",
		MaxInFlight:     int64(5), // TODO(elffjs): Probably need to bump this up.
	}
	consumer, err := kafka.NewConsumer(cfg, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not start contract event consumer")
	}

	cevConsumer := services.NewContractsEventsConsumer(pdb, &logger, settings)
	consumer.Start(context.Background(), cevConsumer.ProcessContractsEventsMessages)

	logger.Info().Msg("Contracts events consumer started")
}

func startMonitoringServer(logger zerolog.Logger, config *config.Settings) {
	monApp := fiber.New(fiber.Config{DisableStartupMessage: true})

	monApp.Use(pprof.New())

	monApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	monApp.Put("/loglevel", changeLogLevel)

	go func() {
		if err := monApp.Listen(":" + config.MonitoringServerPort); err != nil {
			logger.Fatal().Err(err).Str("port", config.MonitoringServerPort).Msg("Failed to start monitoring web server.")
		}
	}()

	logger.Info().Str("port", config.MonitoringServerPort).Msg("Started monitoring web server.")
}
