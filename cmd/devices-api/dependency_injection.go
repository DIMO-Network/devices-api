package main

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/IBM/sarama"

	"github.com/DIMO-Network/devices-api/internal/config"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog"
)

// dependencyContainer way to hold different dependencies we need for our app. We could put all our deps and follow this pattern for everything.
type dependencyContainer struct {
	kafkaProducer   sarama.SyncProducer
	settings        *config.Settings
	logger          *zerolog.Logger
	s3ServiceClient *s3.Client
	ddSvc           services.DeviceDefinitionService
	dbs             func() *db.ReaderWriter
}

func newDependencyContainer(settings *config.Settings, logger zerolog.Logger, dbs func() *db.ReaderWriter) dependencyContainer {
	return dependencyContainer{
		settings: settings,
		logger:   &logger,
		dbs:      dbs,
	}
}

// getKafkaProducer instantiates a new kafka producer if not already set in our container and returns
func (dc *dependencyContainer) getKafkaProducer() sarama.SyncProducer {
	if dc.kafkaProducer == nil {
		p, err := createKafkaProducer(dc.settings)
		if err != nil {
			dc.logger.Fatal().Err(err).Msg("Could not initialize Kafka producer, terminating")
		}
		dc.kafkaProducer = p
	}
	return dc.kafkaProducer
}

// getS3ServiceClient instantiates a new default config and then a new s3 services client if not already set. Takes context in, although it could likely use a context from container passed in on instantiation
func (dc *dependencyContainer) getS3ServiceClient(ctx context.Context) *s3.Client {
	if dc.s3ServiceClient == nil {

		cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(dc.settings.AWSRegion))
		if err != nil {
			dc.logger.Fatal().Err(err).Msg("Could not load aws config, terminating")
		}

		dc.s3ServiceClient = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.Region = dc.settings.AWSRegion
			o.Credentials = credentials.NewStaticCredentialsProvider(dc.settings.DocumentsAWSAccessKeyID, dc.settings.DocumentsAWSSecretsAccessKey, "")

			if dc.settings.Environment == "local" {
				o.BaseEndpoint = &dc.settings.DocumentsAWSEndpoint
				o.UsePathStyle = true
			}
		})
	}
	return dc.s3ServiceClient
}

func (dc *dependencyContainer) getDeviceDefinitionService() services.DeviceDefinitionService {
	dc.ddSvc = services.NewDeviceDefinitionService(dc.dbs, dc.logger, dc.settings)
	return dc.ddSvc
}
