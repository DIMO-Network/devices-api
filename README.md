# devices-api

API & worker for managing devices on the DIMO platform.

For an overview of the project, see the [DIMO technical documentation site.](https://docs.dimo.zone/docs/overview/intro)

## Table of contents

- [Developing locally](#developing-locally)
  - [Kafka test producer](#kafka-test-producer)
  - [Linting](#linting)
  - [Database ORM](#database-orm)
- [Migrations](#migrations)
  - [Managing migrations from k8s](#managing-migrations-from-k8s)
- [Mocks](#mocks)
- [Helm requirements](#helm-requirements)
- [API](#api)
  - [Generating Swagger / OpenAPI spec](#generating-swagger--openapi-spec)
- [gRPC library](#gRPC-library)

## Developing locally

**TL;DR**
```bash
cp settings.sample.yaml settings.yaml
docker compose up -d
go run ./cmd/devices-api migrate
go run ./cmd/devices-api
```

1. Create a settings file by copying the sample
   ```sh
   cp settings.sample.yaml settings.yaml
   ```
   Adjust these as necessary—the sample file should have what you need for local development. (Make sure you do this step each time you run `git pull` in case there have been any changes to the sample settings file.)

2. Start the services
   ```sh
   docker compose up -d
   ```
   This will start a bunch of services. Briefly:

   - Postgres, used to store the basic data models, on port 5432.
   - [Redis](https://redis.io), used by the [taskq library](https://taskq.uptrace.dev) to enqueue interactions with the AutoPi API, on port 6379.
   - [ElasticSearch](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html), only used by the sub-command `search-sync-dds`, on port 9200. Kibana provides a UI for this on port 5601.
   - [LocalStack](https://localstack.cloud), for testing our use of AWS S3 to store user documents and NFTs, takes up ports 4566–4583.
   - [IPFS](https://ipfs.tech), which we hope to use to store device definitions, takes up ports 4001, 8080, 8081, and 5001.
   - [Kafka](https://kafka.apache.org) is used to receive vehicle and task status updates, and emit events. It lives on port 9092, and the supporting Zookeeper service lives on port 2181.

   If you get a port conflict, you can find the existing process using the port with, e.g., `lsof -i :5432`. Most of these containers have attached volumes, so their data will persist across restarts. To check container status, run `docker ps`.

3. You can log into the database now with
   ```sh
   psql -h localhost -p 5432 -U dimo
   ```
   using password `dimo`, or use your favorite UI like [DataGrip](https://www.jetbrains.com/datagrip/). To do anything useful, you'll have to apply the database migrations from the `migrations` folder: 
   ```sh
   go run ./cmd/devices-api migrate
   ```

5. You are now ready to run the application:
   ```sh
   go run ./cmd/devices-api
   ```
It may be helpful to seed the database with test data:

8. Sync Smartcar integration compatibility:
   ```sh
   go run ./cmd/devices-api smartcar-sync
   ```
Finally, if you want to test document uploads:

9. Execute the following command to point the AWS CLI at LocalStack:
   ```sh
   aws --endpoint-url=http://localhost:4566 s3 mb s3://documents
   ```

### Authenticating

One of the variables set in `settings.yaml` is `JWT_KEY_SET_URL`. By default this is set to `http://127.0.0.1:5556/dex/keys`. To make use of this, clone the DIMO Dex fork:
```sh
git clone git@github.com:DIMO-Network/dex.git
cd dex
make build examples
./bin/dex serve examples/config-dev.yaml
```
This will start up the Dex identity server on port 5556. Next, start up the example interface by running
```sh
./bin/example-app
```
You can reach this on port 5555. The "Log in with Example" option is probably the easiest. This will give you an ID token you can provide to the [API](#api).

### Kafka test producer

This tool can be useful to test the consumer when running locally.
`$ go run ./cmd/test-producer <integrationID> <userDeviceID>`

Above integration and vehicle ID's aka userDeviceID should exist in your local DB. 

### Linting

`brew install golangci-lint`

`golangci-lint run`

This should use the settings from `.golangci.yml`, which you can override.

### Database ORM

This is using [sqlboiler](https://github.com/volatiletech/sqlboiler). The ORM models are code generated. If the db changes,
you must update the models.

Make sure you have sqlboiler installed:
```bash
go install github.com/volatiletech/sqlboiler/v4@latest
go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
```

To generate the models:
```bash
sqlboiler psql --no-tests --wipe
```
*Make sure you're running the docker image (ie. docker compose up)*

If you get a command not found error with sqlboiler, make sure your go install is correct. 
[Instructions here](https://jimkang.medium.com/install-go-on-mac-with-homebrew-5fa421fc55f5)

## Migrations

To install goose in GO:
```bash
$ go get github.com/pressly/goose/v3/cmd/goose@v3.5.3
export GOOSE_DRIVER=postgres
```

To install goose CLI:
```bash
$ go install github.com/pressly/goose/v3/cmd/goose
export GOOSE_DRIVER=postgres
```

Add a migrations:
`$ goose -dir migrations create <migration_name> sql`

Migrate DB to latest:
`$ go run ./cmd/devices-api migrate`

Clear DB to start over:
```bash
docker ps
docker stop <container_id>
rm -R ./resources/data/ && mkdir ./resources/data/ 
docker compose up -d
```

If we have code base migrations in the migrations folder, we must import `_ "github.com/DIMO-Network/devices-api/migrations"` in the runner so that
it can find the migrations, otherwise get error.

### Managing migrations from k8s
```bash
kc get pods -n dev
kc exec devices-api-dev-65f8f47ff5-94dp4 -n dev -it -- /bin/sh
./devices-api migrate -down # brings the last migration down
```

## Mocks

To regenerate a mock, you can use go gen since the files that are mocked have a `//go:generate mockgen ...` at the top. For example:
`nhtsa_api_service.go`

## API

Swagger docs at: http://localhost:3000/v1/swagger/index.html

Example curl commands:
```bash
curl http://localhost:3000/v1/user/devices/me
  -H "Authorization: Bearer {token}"
curl -X POST http://localhost:3000/v1/user/devices
   -H 'Content-Type: application/json'
   -H "Authorization: Bearer {token}"
   -d '{"device_definition_id":"{existing device def id}"}'
```

To prettify json, pipe to json_pp: `| json_pp`

Some test VINs:
5YJYGDEE5MF085533
5YJ3E1EA6MF873863

### Generating swagger / openapi spec

Note that swagger must be served from fiber-swagger library v2.31.1 +, since they fixed an issue in previous version. 

To check what cli version you have installed: `swag --version`. As of this writing v1.8.1 is working for us. 
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/devices-api/main.go --parseDependency --parseInternal --generatedTime true 
# optionally add `--parseDepth 2` if have issues
```

[declarative_comments_format](https://swaggo.github.io/swaggo.io/declarative_comments_format/)

### Testing file upload

Replace the file with a file in your system, userDeviceID to one your account controls, and the Authorization header token to yours - get from mobile app
```bash
curl -X POST -F "file=@./some-test-image.png" -F "name=test file" -F "type=VehicleMaintenance" -F "userDeviceID=2Bz5Wv4icb5Il1vBsaFjJKeILN7" \
-H "Authorization: Bearer XXX" \
-H "content-type: application/x-www-form-urlencoded" \
https://devices-api.dimo.zone/v1/documents
```

## gRPC library

We should probably put these in the repositories of the services that own them, but we are putting this off for now. To make changes to the current suite for, e.g., the devices API, run

```
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/grpc/*.proto
```