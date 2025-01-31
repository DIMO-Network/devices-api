package config

import (
	"github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"
	"github.com/DIMO-Network/shared/db"
)

// Settings contains the application config
type Settings struct {
	Environment          string      `yaml:"ENVIRONMENT"`
	Port                 string      `yaml:"PORT"`
	GRPCPort             string      `yaml:"GRPC_PORT"`
	UsersAPIGRPCAddr     string      `yaml:"USERS_API_GRPC_ADDR"`
	LogLevel             string      `yaml:"LOG_LEVEL"`
	DB                   db.Settings `yaml:"DB"`
	ServiceName          string      `yaml:"SERVICE_NAME"`
	JwtKeySetURL         string      `yaml:"JWT_KEY_SET_URL"`
	DeploymentBaseURL    string      `yaml:"DEPLOYMENT_BASE_URL"`
	SmartcarClientID     string      `yaml:"SMARTCAR_CLIENT_ID"`
	SmartcarClientSecret string      `yaml:"SMARTCAR_CLIENT_SECRET"`
	RedisURL             string      `yaml:"REDIS_URL"`
	RedisPassword        string      `yaml:"REDIS_PASSWORD"`
	RedisTLS             bool        `yaml:"REDIS_TLS"`
	IngestSmartcarURL    string      `yaml:"INGEST_SMARTCAR_URL"`
	IngestSmartcarTopic  string      `yaml:"INGEST_SMARTCAR_TOPIC"`
	KafkaBrokers         string      `yaml:"KAFKA_BROKERS"`
	// DeviceStatusTopic                 string      `yaml:"DEVICE_STATUS_TOPIC"`
	PrivacyFenceTopic                 string `yaml:"PRIVACY_FENCE_TOPIC"`
	PrivacyFenceTopicV2               string `yaml:"PRIVACY_FENCE_TOPIC_V2"`
	TaskRunNowTopic                   string `yaml:"TASK_RUN_NOW_TOPIC"`
	TaskStopTopic                     string `yaml:"TASK_STOP_TOPIC"`
	TaskCredentialTopic               string `yaml:"TASK_CREDENTIAL_TOPIC"`
	TaskStatusTopic                   string `yaml:"TASK_STATUS_TOPIC"`
	EventsTopic                       string `yaml:"EVENTS_TOPIC"`
	AWSRegion                         string `yaml:"AWS_REGION"`
	KMSKeyID                          string `yaml:"KMS_KEY_ID"`
	AutoPiAPIToken                    string `yaml:"AUTO_PI_API_TOKEN"`
	AutoPiAPIURL                      string `yaml:"AUTO_PI_API_URL"`
	AWSDocumentsBucketName            string `yaml:"AWS_DOCUMENTS_BUCKET_NAME"`
	NFTS3Bucket                       string `yaml:"NFT_S3_BUCKET"`
	DocumentsAWSAccessKeyID           string `yaml:"DOCUMENTS_AWS_ACCESS_KEY_ID"`
	DocumentsAWSSecretsAccessKey      string `yaml:"DOCUMENTS_AWS_SECRET_ACCESS_KEY"`
	DocumentsAWSEndpoint              string `yaml:"DOCUMENTS_AWS_ENDPOINT"`
	NFTAWSAccessKeyID                 string `yaml:"NFT_AWS_ACCESS_KEY_ID"`
	NFTAWSSecretsAccessKey            string `yaml:"NFT_AWS_SECRET_ACCESS_KEY"`
	DefinitionsGRPCAddr               string `yaml:"DEFINITIONS_GRPC_ADDR"`
	DeviceDataGRPCAddr                string `yaml:"DEVICE_DATA_GRPC_ADDR"`
	DeviceDefinitionTopic             string `yaml:"DEVICE_DEFINITION_TOPIC"`
	DeviceDefinitionMetadataTopic     string `yaml:"DEVICE_DEFINITION_METADATA_TOPIC"`
	ElasticDeviceStatusIndex          string `yaml:"ELASTIC_DEVICE_STATUS_INDEX"`
	ElasticSearchEnrichStatusHost     string `yaml:"ELASTIC_SEARCH_ENRICH_STATUS_HOST"`
	ElasticSearchEnrichStatusUsername string `yaml:"ELASTIC_SEARCH_ENRICH_STATUS_USERNAME"`
	ElasticSearchEnrichStatusPassword string `yaml:"ELASTIC_SEARCH_ENRICH_STATUS_PASSWORD"`
	DIMORegistryAddr                  string `yaml:"DIMO_REGISTRY_ADDR"`
	DIMORegistryChainID               int64  `yaml:"DIMO_REGISTRY_CHAIN_ID"`
	MonitoringServerPort              string `yaml:"MONITORING_SERVER_PORT"`
	TokenExchangeJWTKeySetURL         string `yaml:"TOKEN_EXCHANGE_JWK_KEY_SET_URL"`
	GoogleMapsAPIKey                  string `yaml:"GOOGLE_MAPS_API_KEY"`
	VehicleNFTAddress                 string `yaml:"VEHICLE_NFT_ADDRESS"`
	SyntheticDeviceNFTAddress         string `yaml:"SYNTHETIC_DEVICE_NFT_ADDRESS"`
	ContractsEventTopic               string `yaml:"CONTRACT_EVENT_TOPIC"`
	AutoPiNFTImage                    string `yaml:"AUTOPI_NFT_IMAGE"`
	MacaronNFTImage                   string `yaml:"MACARON_NFT_IMAGE"`
	OpenAISecretKey                   string `yaml:"OPENAI_SECRET_KEY"`
	ChatGPTURL                        string `yaml:"CHATGPT_URL"`
	AftermarketDeviceContractAddress  string `yaml:"AFTERMARKET_DEVICE_CONTRACT_ADDRESS"`
	NATSURL                           string `yaml:"NATS_URL"`
	NATSStreamName                    string `yaml:"NATS_STREAM_NAME"`
	NATSValuationSubject              string `yaml:"NATS_VALUATION_SUBJECT"`
	NATSOfferSubject                  string `yaml:"NATS_OFFER_SUBJECT"`
	NATSAckTimeout                    string `yaml:"NATS_ACK_TIMEOUT"`
	NATSDurableConsumer               string `yaml:"NATS_DURABLE_CONSUMER"`
	ValuationsAPIGRPCAddr             string `yaml:"VALUATIONS_GRPC_ADDR"`

	DIMOContractAPIURL string `yaml:"DIMO_CONTRACT_APIURL"`

	MetaTransactionProcessorGRPCAddr string `yaml:"META_TRANSACTION_PROCESSOR_GRPC_ADDR"`

	// IssuerPrivateKey is a base64-encoded secp256k1 private key, used to sign
	// VIN verifiable credentials.
	IssuerPrivateKey string `yaml:"ISSUER_PRIVATE_KEY"`

	SyntheticWalletGRPCAddr string `yaml:"SYNTHETIC_WALLET_GRPC_ADDR"`
	DeviceFingerprintTopic  string `yaml:"DEVICE_FINGERPRINT_TOPIC"`

	DeviceFingerprintConsumerGroup    string `yaml:"DEVICE_FINGERPRINT_CONSUMER_GROUP"`
	SyntheticFingerprintTopic         string `yaml:"SYNTHETIC_FINGERPRINT_TOPIC"`
	SyntheticFingerprintConsumerGroup string `yaml:"SYNTHETIC_FINGERPRINT_CONSUMER_GROUP"`
	TeslaClientID                     string `yaml:"TESLA_CLIENT_ID"`
	TeslaClientSecret                 string `yaml:"TESLA_CLIENT_SECRET"`
	TeslaTokenURL                     string `yaml:"TESLA_TOKEN_URL"`
	TeslaFleetURL                     string `yaml:"TESLA_FLEET_URL"`
	TeslaTelemetryHostName            string `yaml:"TESLA_TELEMETRY_HOST_NAME"`
	TeslaTelemetryPort                int    `yaml:"TESLA_TELEMETRY_PORT"`
	TeslaTelemetryCACertificate       string `yaml:"TESLA_TELEMETRY_CA_CERTIFICATE"`

	IPFSURL string `yaml:"IPFS_URL"`

	SDInfoTopic string `yaml:"SD_INFO_TOPIC"`
	MainRPCURL  string `yaml:"MAIN_RPC_URL"`

	VehicleDecodingGRPCAddr string `yaml:"VEHICLE_DECODING_GRPC_ADDR"`

	Clickhouse config.Settings `yaml:",inline"`

	DeviceDefinitionsGetByKSUIDEndpoint string `yaml:"DEVICE_DEFINITIONS_GET_BY_KSUID_ENDPOINT"`

	TeslaRequiredScopes string `json:"TESLA_REQUIRED_SCOPES"`

	AccountsAPIGRPCAddr string `json:"ACCOUNTS_API_GRPC_ADDR"`
	CustomerIOAPIKey    string `yaml:"CUSTOMER_IO_API_KEY"`
}

func (s *Settings) IsProduction() bool {
	return s.Environment == "prod" // this string is set in the helm chart values-prod.yaml
}
