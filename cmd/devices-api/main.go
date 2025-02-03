package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/subcommands"

	_ "github.com/DIMO-Network/devices-api/docs"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/kafka"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/IBM/sarama"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/burdiyan/kafkautil"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	_ "github.com/lib/pq"
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
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "devices-api").Logger()

	// TODO(elffjs): Extract this somewhere.
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				if len(s.Value) >= 7 {
					logger = logger.With().Str("commit", s.Value[:7]).Logger()
				}
				break
			}
		}
	}

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
		startMonitoringServer(logger, &settings)
		eventService := services.NewEventService(&logger, &settings, deps.getKafkaProducer())
		startCredentialConsumer(logger, &settings, pdb)
		startTaskStatusConsumer(logger, &settings, pdb)
		startWebAPI(logger, &settings, pdb, eventService, deps.getKafkaProducer(), deps.getS3ServiceClient(ctx), deps.getS3NFTServiceClient(ctx))
	} else {
		subcommands.Register(&migrateDBCmd{logger: logger, settings: settings}, "database")
		subcommands.Register(&findOldStyleTasks{logger: logger, settings: settings, pdb: pdb}, "events")

		subcommands.Register(&generateEventCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "events")
		subcommands.Register(&setCommandCompatibilityCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&remakeAutoPiTopicCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&remakeAftermarketTopicCmd{logger: logger, settings: settings, pdb: pdb, container: deps}, "device integrations")
		subcommands.Register(&remakeUserDeviceTokenTableCmd{logger: logger, settings: settings, pdb: pdb, container: deps}, "device integrations")
		subcommands.Register(&remakeFenceTopicCmd{logger: logger, settings: settings, pdb: pdb}, "device integrations")

		{
			var cipher shared.Cipher
			if settings.Environment == "dev" || settings.IsProduction() {
				cipher = createKMS(&settings, &logger)
			} else {
				logger.Warn().Msg("Using ROT13 encrypter. Only use this for testing!")
				cipher = new(shared.ROT13Cipher)
			}
			subcommands.Register(&checkVirtualKeyCmd{logger: logger, settings: settings, pdb: pdb, cipher: cipher}, "device integrations")
			subcommands.Register(&enableTelemetryCmd{logger: logger, settings: settings, pdb: pdb, cipher: cipher}, "device integrations")
		}

		subcommands.Register(&populateSDInfoTopicCmd{logger: logger, settings: settings, pdb: pdb, container: deps}, "device integrations")
		subcommands.Register(&populateTeslaTelemetryMapCmd{logger: logger, settings: settings, pdb: pdb, container: deps}, "device integrations")
		subcommands.Register(&populatePrivacyV2Topic{logger: logger, settings: settings, pdb: pdb, container: deps}, "device integrations")
		subcommands.Register(&remakeDeviceDefinitionTopicsCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&populateSDFingerprintTable{logger: logger, settings: settings, pdb: pdb, container: deps}, "device integrations")
		subcommands.Register(&updateStateCmd{logger: logger, settings: settings, pdb: pdb}, "device integrations")
		subcommands.Register(&web2PairCmd{logger: logger, settings: settings, pdb: pdb, container: deps}, "device integrations")
		subcommands.Register(&autoPiKTableDeleteCmd{logger: logger, container: deps}, "device integrations")
		subcommands.Register(&startSDTask{logger: logger, container: deps, settings: settings, pdb: pdb}, "device integrations")
		subcommands.Register(&startIntegrationTask{logger: logger, container: deps, settings: settings, pdb: pdb}, "device integrations")
		subcommands.Register(&smartcarStopConnectionsCmd{logger: logger, settings: settings, pdb: pdb, smartcarTaskSvc: services.NewSmartcarTaskService(&settings, deps.getKafkaProducer())}, "device integrations")

		subcommands.Register(&populateESDDDataCmd{logger: logger, settings: settings, pdb: pdb, esInstance: deps.getElasticSearchService(), ddSvc: deps.getDeviceDefinitionService()}, "populate data")
		subcommands.Register(&populateESRegionDataCmd{logger: logger, settings: settings, pdb: pdb, esInstance: deps.getElasticSearchService(), ddSvc: deps.getDeviceDefinitionService()}, "populate data")

		subcommands.Register(&stopTaskByKeyCmd{logger: logger, settings: settings, container: deps, pdb: pdb}, "tasks")

		subcommands.Register(&syncDeviceTemplatesCmd{logger: logger, settings: settings, pdb: pdb}, "user devices")
		subcommands.Register(&vinDecodeCompareCmd{logger: logger, settings: settings, pdb: pdb}, "user devices")

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
	clusterConfig.Version = sarama.V3_6_0_0
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

	kcf := sarama.NewConfig()
	kcf.Version = sarama.V3_6_0_0
	kcf.Producer.Partitioner = kafkautil.NewJVMCompatiblePartitioner
	kcf.Producer.Return.Successes = true

	kp, err := sarama.NewSyncProducer(strings.Split(settings.KafkaBrokers, ","), kcf)
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not create Kafka producer.")
	}

	ddSvc := services.NewDeviceDefinitionService(pdb.DBS, &logger, settings)

	taskStatusService := services.NewTaskStatusListener(pdb.DBS, &logger, ddSvc, kp, settings)
	consumer.Start(context.Background(), taskStatusService.ProcessTaskUpdates)

	logger.Info().Msg("Task status consumer started")
}

func startContractEventsConsumer(logger zerolog.Logger, settings *config.Settings, pdb db.Store, genericADInteg services.Integration, ddSvc services.DeviceDefinitionService, evtSvc services.EventService, scTask services.SmartcarTaskService, teslaTask services.TeslaTaskService) {
	cevConsumer := services.NewContractsEventsConsumer(pdb, &logger, settings, genericADInteg, ddSvc, evtSvc, scTask, teslaTask)
	if err := cevConsumer.RunConsumer(); err != nil {
		logger.Fatal().Err(err).Msg("error occurred processing contract events")
	}

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
