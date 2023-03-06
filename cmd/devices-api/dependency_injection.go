package main

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/elasticsearch"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	es "github.com/DIMO-Network/devices-api/internal/elasticsearch"
	"github.com/Shopify/sarama"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog"
)

// dependencyContainer way to hold different dependencies we need for our app. We could put all our deps and follow this pattern for everything.
type dependencyContainer struct {
	kafkaProducer      sarama.SyncProducer
	settings           *config.Settings
	logger             *zerolog.Logger
	s3ServiceClient    *s3.Client
	s3NFTServiceClient *s3.Client
	nhtsaSvc           services.INHTSAService
	ddSvc              services.DeviceDefinitionService
	dbs                func() *db.ReaderWriter
	elasticSearch      elasticsearch.ElasticSearch
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

		cfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(dc.settings.AWSRegion),
			// Comment the below out if not using localhost
			awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {

					if dc.settings.Environment == "local" {
						return aws.Endpoint{PartitionID: "aws", URL: dc.settings.DocumentsAWSEndpoint, SigningRegion: dc.settings.AWSRegion}, nil // The SigningRegion key was what's was missing! D'oh.
					}

					// returning EndpointNotFoundError will allow the service to fallback to its default resolution
					return aws.Endpoint{}, &aws.EndpointNotFoundError{}
				})))

		if err != nil {
			dc.logger.Fatal().Err(err).Msg("Could not load aws config, terminating")
		}

		dc.s3ServiceClient = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.Region = dc.settings.AWSRegion
			o.Credentials = credentials.NewStaticCredentialsProvider(dc.settings.DocumentsAWSAccessKeyID, dc.settings.DocumentsAWSSecretsAccessKey, "")
		})
	}
	return dc.s3ServiceClient
}

func (dc *dependencyContainer) getS3NFTServiceClient(ctx context.Context) *s3.Client {
	if dc.s3NFTServiceClient == nil {

		cfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(dc.settings.AWSRegion),
			// Comment the below out if not using localhost
			awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {

					if dc.settings.Environment == "local" {
						return aws.Endpoint{PartitionID: "aws", URL: dc.settings.DocumentsAWSEndpoint, SigningRegion: dc.settings.AWSRegion}, nil // The SigningRegion key was what's was missing! D'oh.
					}

					// returning EndpointNotFoundError will allow the service to fallback to its default resolution
					return aws.Endpoint{}, &aws.EndpointNotFoundError{}
				})))

		if err != nil {
			dc.logger.Fatal().Err(err).Msg("Could not load aws config, terminating")
		}

		dc.s3NFTServiceClient = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.Region = dc.settings.AWSRegion
			o.Credentials = credentials.NewStaticCredentialsProvider(dc.settings.NFTAWSAccessKeyID, dc.settings.NFTAWSSecretsAccessKey, "")
		})
	}
	return dc.s3NFTServiceClient
}

func (dc *dependencyContainer) getNHTSAService() services.INHTSAService {
	dc.nhtsaSvc = services.NewNHTSAService()
	return dc.nhtsaSvc
}

func (dc *dependencyContainer) getDeviceDefinitionService() services.DeviceDefinitionService {
	dc.ddSvc = services.NewDeviceDefinitionService(dc.dbs, dc.logger, dc.getNHTSAService(), dc.settings)
	return dc.ddSvc
}

func (dc *dependencyContainer) getElasticSearchService() elasticsearch.ElasticSearch {
	esInstance, err := es.NewElasticSearch(*dc.settings, dc.logger)
	if err != nil {
		dc.logger.Fatal().Err(err).Msgf("Couldn't instantiate Elasticsearch client.")
	}

	dc.elasticSearch = esInstance

	return dc.elasticSearch
}
