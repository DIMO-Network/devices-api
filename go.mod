module github.com/DIMO-Network/devices-api

go 1.21

toolchain go1.21.5

require (
	github.com/DIMO-Network/device-data-api v0.4.23-0.20230718210255-60a2c7be0aa7
	github.com/DIMO-Network/device-definitions-api v1.0.32
	github.com/DIMO-Network/go-mnemonic v0.0.0-20230406181942-6ddfe6f8c21c
	github.com/DIMO-Network/shared v0.10.6
	github.com/DIMO-Network/synthetic-wallet-instance v0.0.0-20230601233541-6a4c8afb27d3
	github.com/DIMO-Network/valuations-api v0.2.0
	github.com/DIMO-Network/zflogger v1.0.0-beta
	github.com/Shopify/sarama v1.38.1
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/ThreeDotsLabs/watermill-kafka/v2 v2.2.1
	github.com/aws/aws-sdk-go-v2 v1.24.1
	github.com/aws/aws-sdk-go-v2/config v1.26.3
	github.com/aws/aws-sdk-go-v2/service/kms v1.27.9
	github.com/aws/aws-sdk-go-v2/service/s3 v1.48.0
	github.com/btcsuite/btcd/btcutil v1.1.3
	github.com/burdiyan/kafkautil v0.0.0-20190131162249-eaf83ed22d5b
	github.com/docker/go-connections v0.4.0
	github.com/elastic/go-elasticsearch/v8 v8.6.0
	github.com/friendsofgo/errors v0.9.2
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gofiber/fiber/v2 v2.52.0
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/google/subcommands v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.1
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/jarcoal/httpmock v1.1.0
	github.com/lib/pq v1.10.9
	github.com/nats-io/nats-server/v2 v2.9.16
	github.com/nats-io/nats.go v1.25.0
	github.com/piprate/json-gold v0.5.0
	github.com/pkg/errors v0.9.1
	github.com/pressly/goose/v3 v3.17.0
	github.com/prometheus/client_golang v1.17.0
	github.com/rs/zerolog v1.31.0
	github.com/segmentio/ksuid v1.0.4
	github.com/stretchr/testify v1.8.4
	github.com/swaggo/swag v1.16.2
	github.com/testcontainers/testcontainers-go v0.26.0
	github.com/tidwall/gjson v1.17.0
	github.com/vmihailenco/taskq/v3 v3.2.8
	github.com/volatiletech/null/v8 v8.1.2
	github.com/volatiletech/sqlboiler/v4 v4.15.0
	github.com/volatiletech/strmangle v0.0.5
	go.uber.org/automaxprocs v1.5.1
	go.uber.org/mock v0.3.0
	golang.org/x/mod v0.14.0
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/MicahParks/keyfunc/v2 v2.1.0 // indirect
	github.com/Microsoft/hcsshim v0.11.4 // indirect
	github.com/avast/retry-go v3.0.0+incompatible // indirect
	github.com/aws/aws-sdk-go-v2/service/route53 v1.36.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.21.6 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/deckarep/golang-set/v2 v2.1.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/elastic/elastic-transport-go/v8 v8.1.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/holiman/uint256 v1.2.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/lovoo/goka v1.1.7 // indirect
	github.com/lufia/plan9stats v0.0.0-20231016141302-07b5767bb0ed // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/nats-io/jwt/v2 v2.4.1 // indirect
	github.com/nats-io/nkeys v0.4.4 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/power-devops/perfstat v0.0.0-20221212215047-62379fc7944b // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/sethvargo/go-retry v0.2.4 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/shirou/gopsutil/v3 v3.23.11 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/swaggo/files/v2 v2.0.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tinylib/msgp v1.1.8 // indirect
	github.com/tklauser/go-sysconf v0.3.13 // indirect
	github.com/tklauser/numcpus v0.7.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231212172506-995d672761c0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/avast/retry-go/v4 v4.5.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.5.4 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.16.14
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.16.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.18.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.26.7 // indirect
	github.com/aws/smithy-go v1.19.0 // indirect
	github.com/bsm/redislock v0.7.2 // indirect
	github.com/capnm/sysinfo v0.0.0-20130621111458-5909a53897f3 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/containerd/containerd v1.7.11 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker v24.0.7+incompatible // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/ericlagergren/decimal v0.0.0-20211103172832-aca2edc11f73
	github.com/go-redis/redis_rate/v9 v9.1.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc5 // indirect
	github.com/opencontainers/runc v1.1.10 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/eapache/go-resiliency v1.3.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230111030713-bf00bc1b83b6 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.5.0
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.3 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	golang.org/x/crypto v0.16.0 // indirect
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/andybalholm/brotli v1.0.6 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ethereum/go-ethereum v1.12.0
	github.com/go-openapi/jsonpointer v0.20.1 // indirect
	github.com/go-openapi/jsonreference v0.20.3 // indirect
	github.com/go-openapi/spec v0.20.12 // indirect
	github.com/go-openapi/swag v0.22.5 // indirect
	github.com/gofiber/contrib/jwt v1.0.2
	github.com/gofiber/swagger v0.1.14
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/smartcar/go-sdk v1.4.0
	github.com/spf13/cast v1.6.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/volatiletech/inflect v0.0.1 // indirect
	github.com/volatiletech/randomize v0.0.1 // indirect
	golang.org/x/exp v0.0.0-20231219180239-dc181d75b848
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.16.1 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
)

replace github.com/vmihailenco/taskq/v3 v3.2.8 => github.com/DIMO-Network/taskq/v3 v3.2.9-0.20220518233332-179b5552605f

replace github.com/ericlagergren/decimal => github.com/ericlagergren/decimal v0.0.0-20181231230500-73749d4874d5
