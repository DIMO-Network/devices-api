package config

import (
	"github.com/DIMO-Network/shared/db"
)

// Settings contains the application config
type Settings struct {
	Environment                       string      `yaml:"ENVIRONMENT"`
	Port                              string      `yaml:"PORT"`
	GRPCPort                          string      `yaml:"GRPC_PORT"`
	UsersAPIGRPCAddr                  string      `yaml:"USERS_API_GRPC_ADDR"`
	LogLevel                          string      `yaml:"LOG_LEVEL"`
	DB                                db.Settings `yaml:"DB"`
	ServiceName                       string      `yaml:"SERVICE_NAME"`
	JwtKeySetURL                      string      `yaml:"JWT_KEY_SET_URL"`
	DeploymentBaseURL                 string      `yaml:"DEPLOYMENT_BASE_URL"`
	SmartcarClientID                  string      `yaml:"SMARTCAR_CLIENT_ID"`
	SmartcarClientSecret              string      `yaml:"SMARTCAR_CLIENT_SECRET"`
	SmartcarTestMode                  bool        `yaml:"SMARTCAR_TEST_MODE"`
	SmartcarWebhookID                 string      `yaml:"SMARTCAR_WEBHOOK_ID"`
	RedisURL                          string      `yaml:"REDIS_URL"`
	RedisPassword                     string      `yaml:"REDIS_PASSWORD"`
	RedisTLS                          bool        `yaml:"REDIS_TLS"`
	IngestSmartcarURL                 string      `yaml:"INGEST_SMARTCAR_URL"`
	IngestSmartcarTopic               string      `yaml:"INGEST_SMARTCAR_TOPIC"`
	KafkaBrokers                      string      `yaml:"KAFKA_BROKERS"`
	DeviceStatusTopic                 string      `yaml:"DEVICE_STATUS_TOPIC"`
	PrivacyFenceTopic                 string      `yaml:"PRIVACY_FENCE_TOPIC"`
	TaskRunNowTopic                   string      `yaml:"TASK_RUN_NOW_TOPIC"`
	TaskStopTopic                     string      `yaml:"TASK_STOP_TOPIC"`
	TaskCredentialTopic               string      `yaml:"TASK_CREDENTIAL_TOPIC"`
	TaskStatusTopic                   string      `yaml:"TASK_STATUS_TOPIC"`
	EventsTopic                       string      `yaml:"EVENTS_TOPIC"`
	ElasticSearchAppSearchHost        string      `yaml:"ELASTIC_SEARCH_APP_SEARCH_HOST"`
	ElasticSearchAppSearchToken       string      `yaml:"ELASTIC_SEARCH_APP_SEARCH_TOKEN"`
	DeviceDataIndexName               string      `yaml:"DEVICE_DATA_INDEX_NAME"`
	AWSRegion                         string      `yaml:"AWS_REGION"`
	KMSKeyID                          string      `yaml:"KMS_KEY_ID"`
	AutoPiAPIToken                    string      `yaml:"AUTO_PI_API_TOKEN"`
	AutoPiAPIURL                      string      `yaml:"AUTO_PI_API_URL"`
	SmartcarManagementToken           string      `yaml:"SMARTCAR_MANAGEMENT_TOKEN"`
	CIOSiteID                         string      `yaml:"CIO_SITE_ID"`
	CIOApiKey                         string      `yaml:"CIO_API_KEY"`
	AWSDocumentsBucketName            string      `yaml:"AWS_DOCUMENTS_BUCKET_NAME"`
	NFTS3Bucket                       string      `yaml:"NFT_S3_BUCKET"`
	DocumentsAWSAccessKeyID           string      `yaml:"DOCUMENTS_AWS_ACCESS_KEY_ID"`
	DocumentsAWSSecretsAccessKey      string      `yaml:"DOCUMENTS_AWS_SECRET_ACCESS_KEY"`
	DocumentsAWSEndpoint              string      `yaml:"DOCUMENTS_AWS_ENDPOINT"`
	NFTAWSAccessKeyID                 string      `yaml:"NFT_AWS_ACCESS_KEY_ID"`
	NFTAWSSecretsAccessKey            string      `yaml:"NFT_AWS_SECRET_ACCESS_KEY"`
	IPFSNodeEndpoint                  string      `yaml:"IPFS_NODE_ENDPOINT"`
	DrivlyAPIKey                      string      `yaml:"DRIVLY_API_KEY"`
	DrivlyVINAPIURL                   string      `yaml:"DRIVLY_VIN_API_URL"`
	DrivlyOfferAPIURL                 string      `yaml:"DRIVLY_OFFER_API_URL"`
	DefinitionsGRPCAddr               string      `yaml:"DEFINITIONS_GRPC_ADDR"`
	DeviceDefinitionTopic             string      `yaml:"DEVICE_DEFINITION_TOPIC"`
	DeviceDefinitionMetadataTopic     string      `yaml:"DEVICE_DEFINITION_METADATA_TOPIC"`
	ElasticDeviceStatusIndex          string      `yaml:"ELASTIC_DEVICE_STATUS_INDEX"`
	ElasticSearchEnrichStatusHost     string      `yaml:"ELASTIC_SEARCH_ENRICH_STATUS_HOST"`
	ElasticSearchEnrichStatusUsername string      `yaml:"ELASTIC_SEARCH_ENRICH_STATUS_USERNAME"`
	ElasticSearchEnrichStatusPassword string      `yaml:"ELASTIC_SEARCH_ENRICH_STATUS_PASSWORD"`
	DIMORegistryAddr                  string      `yaml:"DIMO_REGISTRY_ADDR"`
	DIMORegistryChainID               int64       `yaml:"DIMO_REGISTRY_CHAIN_ID"`
	MonitoringServerPort              string      `yaml:"MONITORING_SERVER_PORT"`
	TokenExchangeJWTKeySetURL         string      `yaml:"TOKEN_EXCHANGE_JWK_KEY_SET_URL"`
	EnablePrivileges                  bool        `yaml:"ENABLE_PRIVILEGES"`
	GoogleMapsAPIKey                  string      `yaml:"GOOGLE_MAPS_API_KEY"`
	VehicleNFTAddress                 string      `yaml:"VEHICLE_NFT_ADDRESS"`
	ContractsEventTopic               string      `yaml:"CONTRACT_EVENT_TOPIC"`
	AutoPiNFTImage                    string      `yaml:"AUTOPI_NFT_IMAGE"`
}
