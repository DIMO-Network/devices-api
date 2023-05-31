module github.com/DIMO-Network/devices-api

go 1.19

require (
	github.com/DIMO-Network/device-definitions-api v1.0.7
	github.com/DIMO-Network/go-mnemonic v0.0.0-20230406181942-6ddfe6f8c21c
	github.com/DIMO-Network/shared v0.9.1
	github.com/DIMO-Network/zflogger v1.0.0-beta
	github.com/Shopify/sarama v1.38.1
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/ThreeDotsLabs/watermill-kafka/v2 v2.2.1
	github.com/aws/aws-sdk-go-v2 v1.17.6
	github.com/aws/aws-sdk-go-v2/config v1.15.10
	github.com/aws/aws-sdk-go-v2/service/kms v1.20.7
	github.com/aws/aws-sdk-go-v2/service/s3 v1.26.11
	github.com/btcsuite/btcd/btcutil v1.1.3
	github.com/burdiyan/kafkautil v0.0.0-20190131162249-eaf83ed22d5b
	github.com/docker/go-connections v0.4.0
	github.com/elastic/go-elasticsearch/v8 v8.4.0
	github.com/friendsofgo/errors v0.9.2
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gofiber/contrib/jwt v1.0.0
	github.com/gofiber/fiber/v2 v2.46.0
	github.com/gofiber/swagger v0.1.12
	github.com/golang-jwt/jwt/v4 v4.4.3
	github.com/golang/mock v1.6.0
	github.com/google/subcommands v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/jarcoal/httpmock v1.1.0
	github.com/lib/pq v1.10.7
	github.com/lovoo/goka v1.1.7
	github.com/nats-io/nats-server/v2 v2.9.16
	github.com/nats-io/nats.go v1.25.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/piprate/json-gold v0.5.0
	github.com/pkg/errors v0.9.1
	github.com/pressly/goose/v3 v3.5.3
	github.com/prometheus/client_golang v1.15.1
	github.com/rs/zerolog v1.28.0
	github.com/segmentio/ksuid v1.0.4
	github.com/stretchr/testify v1.8.2
	github.com/swaggo/swag v1.16.1
	github.com/testcontainers/testcontainers-go v0.14.0
	github.com/tidwall/gjson v1.14.3
	github.com/vmihailenco/taskq/v3 v3.2.8
	github.com/volatiletech/null/v8 v8.1.2
	github.com/volatiletech/sqlboiler/v4 v4.13.0
	github.com/volatiletech/strmangle v0.0.4
	go.uber.org/automaxprocs v1.5.1
	golang.org/x/mod v0.10.0
)

require (
	github.com/MicahParks/keyfunc/v2 v2.0.3 // indirect
	github.com/avast/retry-go v3.0.0+incompatible // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/deckarep/golang-set/v2 v2.1.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/elastic/elastic-transport-go/v8 v8.1.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.0.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/holiman/uint256 v1.2.2-0.20230321075855-87b91420868c // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/nats-io/jwt/v2 v2.4.1 // indirect
	github.com/nats-io/nkeys v0.4.4 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/savsgio/dictpool v0.0.0-20221023140959-7bf2e61cea94 // indirect
	github.com/savsgio/gotils v0.0.0-20230208104028-c358bd845dee // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/swaggo/files/v2 v2.0.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tinylib/msgp v1.1.8 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/sync v0.2.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Microsoft/hcsshim v0.9.4 // indirect
	github.com/avast/retry-go/v4 v4.3.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.2 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.12.5
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.6 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.24 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.7 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/bsm/redislock v0.7.2 // indirect
	github.com/capnm/sysinfo v0.0.0-20130621111458-5909a53897f3 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/containerd/cgroups v1.0.4 // indirect
	github.com/containerd/containerd v1.6.8 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.18+incompatible // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/ericlagergren/decimal v0.0.0-20211103172832-aca2edc11f73
	github.com/go-redis/redis_rate/v9 v9.1.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/moby/sys/mount v0.3.3 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/moby/term v0.0.0-20220808134915-39b0c02b01ae // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20211202183452-c5a74bcca799 // indirect
	github.com/opencontainers/runc v1.1.4 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/RichardKnop/logging v0.0.0-20190827224416-1a693bdd4fae // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/eapache/go-resiliency v1.3.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230111030713-bf00bc1b83b6 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.3.0
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.3 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect; indirectn
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/crypto v0.8.0 // indirect
	google.golang.org/genproto v0.0.0-20220916172020-2692e8806bfa // indirect
	google.golang.org/grpc v1.49.0
	google.golang.org/protobuf v1.30.0
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/RichardKnop/machinery v1.10.6
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/customerio/go-customerio/v3 v3.4.1
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ethereum/go-ethereum v1.12.0
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/spec v0.20.9 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gofrs/uuid v4.3.0+incompatible // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.16.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/smartcar/go-sdk v1.4.0
	github.com/spf13/cast v1.5.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.47.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/volatiletech/inflect v0.0.1 // indirect
	github.com/volatiletech/randomize v0.0.1 // indirect
	golang.org/x/exp v0.0.0-20230206171751-46f607a40771
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/tools v0.9.1 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
)

replace github.com/vmihailenco/taskq/v3 v3.2.8 => github.com/DIMO-Network/taskq/v3 v3.2.9-0.20220518233332-179b5552605f

replace github.com/ericlagergren/decimal => github.com/ericlagergren/decimal v0.0.0-20181231230500-73749d4874d5
