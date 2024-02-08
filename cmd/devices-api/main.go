package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/subcommands"

	_ "github.com/DIMO-Network/devices-api/docs"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/kafka"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/autopi"
	"github.com/DIMO-Network/devices-api/internal/services/macaron"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/Shopify/sarama"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/burdiyan/kafkautil"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	mt "github.com/txaty/go-merkletree"
	_ "go.uber.org/automaxprocs"
)

type leaf struct {
	Data []byte
}

func concatSortHash(b1 []byte, b2 []byte) []byte {
	if bytes.Compare(b1, b2) < 0 {
		return concatHash(b1, b2)
	}
	return concatHash(b2, b1)
}

// type onChainAttestation struct {
// 	VehicleTokenId         int
// 	DeviceDefinitionId     string
// 	VerifiableCredentialId string
// 	VCSignature            string
// }

// TODO (ae): will this work for all types? need to check, might not want to infer
func abiEncode(values []interface{}) ([]byte, error) {
	argTypes := []reflect.Type{}
	for _, val := range values {
		argTypes = append(argTypes, reflect.TypeOf(val))
	}

	args := abi.Arguments{}
	for _, argType := range argTypes {
		abiType, err := abi.NewType(argType.Name(), "", nil)
		if err != nil {
			return nil, err
		}
		args = append(args, abi.Argument{Type: abiType})
	}

	return args.Pack(values...)
}

func (l *leaf) Serialize() ([]byte, error) {
	codeBytes, err := abiEncode([]interface{}{string(l.Data)})
	if err != nil {
		return nil, err
	}

	return crypto.Keccak256(crypto.Keccak256(codeBytes)), nil
}

// concatHash concatenates two byte slices, b1 and b2.
func concatHash(b1 []byte, b2 []byte) []byte {
	result := make([]byte, len(b1)+len(b2))
	copy(result, b1)
	copy(result[len(b1):], b2)
	return result
}

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

	if len(os.Args) >= 1 && os.Args[len(os.Args)-1] == "merkle-test" {

		conf := mt.Config{
			SortSiblingPairs: true, // parameter for OpenZeppelin compatibility
			HashFunc: func(data []byte) ([]byte, error) {
				return crypto.Keccak256(data), nil
				// return data, nil

			},
			Mode:               mt.ModeProofGenAndTreeBuild,
			RunInParallel:      true,
			DisableLeafHashing: true,
			NumRoutines:        0,
		}

		//  need to make sure this is sorted BEFORE creating the tree bc openzep will hash this
		valid := []string{
			"bob -> dave",
			"dave -> bob",
			"carol -> alice",
			"alice -> bob",
		}
		// the following order would also work OTHERS WILL NOT
		// "carol -> alice",
		// "alice -> bob",
		// "bob -> dave",
		// "dave -> bob",

		blocks := []mt.DataBlock{}
		for _, vc := range valid { // Actual data
			blocks = append(blocks, &leaf{Data: []byte(vc)})
		}

		tree, err := mt.New(&conf, blocks)
		if err != nil {
			fmt.Println("couldnt make tree")
			panic(err)
		}

		fmt.Println(" Tree Root: ", hexutil.Encode(tree.Root))

		for n, b := range blocks {
			s, _ := b.Serialize()
			fmt.Println("Block ", n, "\t\t", hexutil.Encode(s))

		}

		for _, l := range blocks {
			// The following proofs can all be validated using openzeppelin MerkleProof verify function
			for _, p := range tree.Proofs {

				for _, x := range blocks {
					s2, _ := x.Serialize()
					s1, _ := x.Serialize()
					if string(s2) != string(s1) {
						fmt.Println("\t\t", hexutil.Encode(concatHash(s2, s1)))
					}
				}

				valid, err := mt.Verify(l, p, tree.Root, &conf)
				if err != nil {
					fmt.Println("couldnt verify")
					panic(err)
				}

				if valid {
					a := l.(*leaf)
					data, _ := l.Serialize()
					fmt.Println("\nValidated: ", hexutil.Encode(data), data, string(a.Data))
					codeBytes, err := abiEncode([]interface{}{string(a.Data)})
					if err != nil {
						panic(err)
					}
					fmt.Println("\tSingle Hash:", hexutil.Encode(crypto.Keccak256(codeBytes)))
					fmt.Println("\tPath: ", p.Path)
					fmt.Println("\tSiblings: ")
					for _, s := range p.Siblings {
						fmt.Println("\t\t", hexutil.Encode(s))
					}
				}
			}
		}

		return
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
		subcommands.Register(&generateEventCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "events")
		subcommands.Register(&setCommandCompatibilityCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&remakeAutoPiTopicCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&remakeAftermarketTopicCmd{logger: logger, settings: settings, pdb: pdb, container: deps}, "device integrations")
		subcommands.Register(&remakeUserDeviceTokenTableCmd{logger: logger, settings: settings, pdb: pdb, container: deps}, "device integrations")
		subcommands.Register(&remakeFenceTopicCmd{logger: logger, settings: settings, pdb: pdb}, "device integrations")
		subcommands.Register(&remakeDeviceDefinitionTopicsCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&startSmartcarFromRefreshCmd{logger: logger, settings: settings, pdb: pdb, ddSvc: deps.getDeviceDefinitionService()}, "device integrations")
		subcommands.Register(&autopiToolsCmd{logger: logger, settings: settings, pdb: pdb}, "device integrations")
		subcommands.Register(&updateStateCmd{logger: logger, settings: settings, pdb: pdb}, "device integrations")
		subcommands.Register(&web2PairCmd{logger: logger, settings: settings, pdb: pdb, container: deps}, "device integrations")
		subcommands.Register(&autoPiKTableDeleteCmd{logger: logger, container: deps}, "device integrations")

		subcommands.Register(&populateESDDDataCmd{logger: logger, settings: settings, pdb: pdb, esInstance: deps.getElasticSearchService(), ddSvc: deps.getDeviceDefinitionService()}, "populate data")
		subcommands.Register(&populateESRegionDataCmd{logger: logger, settings: settings, pdb: pdb, esInstance: deps.getElasticSearchService(), ddSvc: deps.getDeviceDefinitionService()}, "populate data")
		subcommands.Register(&populateUSAPowertrainCmd{logger: logger, settings: settings, pdb: pdb, nhtsaService: deps.getNHTSAService(), deviceDefSvc: deps.getDeviceDefinitionService()}, "populate data")

		subcommands.Register(&stopTaskByKeyCmd{logger: logger, settings: settings, container: deps}, "tasks")

		subcommands.Register(&syncDeviceTemplatesCmd{logger: logger, settings: settings, pdb: pdb}, "user devices")

		subcommands.Register(&fixSignalTimestamps{logger: logger, settings: settings, pdb: pdb}, "data-fixes")

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

	nhtsaSvc := services.NewNHTSAService()
	ddSvc := services.NewDeviceDefinitionService(pdb.DBS, &logger, nhtsaSvc, settings)

	taskStatusService := services.NewTaskStatusListener(pdb.DBS, &logger, ddSvc)
	consumer.Start(context.Background(), taskStatusService.ProcessTaskUpdates)

	logger.Info().Msg("Task status consumer started")
}

func startContractEventsConsumer(logger zerolog.Logger, settings *config.Settings, pdb db.Store, autoPi *autopi.Integration, macaron *macaron.Integration, ddSvc services.DeviceDefinitionService) {
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

	cevConsumer := services.NewContractsEventsConsumer(pdb, &logger, settings, autoPi, macaron, ddSvc)
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

func permute(a []string, l, r int, result *[][]string) {
	if l == r {
		// Make a copy of the current permutation and add it to the result.
		permutation := make([]string, len(a))
		copy(permutation, a)
		*result = append(*result, permutation)
	} else {
		for i := l; i <= r; i++ {
			a[l], a[i] = a[i], a[l] // swap
			permute(a, l+1, r, result)
			a[l], a[i] = a[i], a[l] // backtrack: swap back
		}
	}
}
