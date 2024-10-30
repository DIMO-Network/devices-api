.PHONY: all deps docker gen-proto docker-cgo clean docs test test-race fmt lint install deploy-docs

TAGS =

INSTALL_DIR        = $(GOPATH)/bin
DEST_DIR           = ./target
PATHINSTBIN        = $(DEST_DIR)/bin
PATHINSTDOCKER     = $(DEST_DIR)/docker

VERSION   := $(shell git describe --tags || echo "v0.0.0")
VER_CUT   := $(shell echo $(VERSION) | cut -c2-)
VER_MAJOR := $(shell echo $(VER_CUT) | cut -f1 -d.)
VER_MINOR := $(shell echo $(VER_CUT) | cut -f2 -d.)
VER_PATCH := $(shell echo $(VER_CUT) | cut -f3 -d.)
VER_RC    := $(shell echo $(VER_PATCH) | cut -f2 -d-)
DATE      := $(shell date +"%Y-%m-%dT%H:%M:%SZ")

LD_FLAGS   =
GO_FLAGS   =
DOCS_FLAGS =

APPS = devices-api
all: $(APPS)

install: $(APPS)
	@mkdir -p bin
	@cp $(PATHINSTBIN)/devices-api ./bin/

deps:
	@go mod tidy
	@go mod vendor

SOURCE_FILES = $(shell find lib internal -type f -name "*.go")


$(PATHINSTBIN)/%: $(SOURCE_FILES) 
	@go build $(GO_FLAGS) -tags "$(TAGS)" -ldflags "$(LD_FLAGS) " -o $@ ./cmd/$*

$(APPS): %: $(PATHINSTBIN)/%

gen-mocks:
	@go generate ./...

gen-proto:
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/grpc/*.proto

gen-swag:
	@swag init -g cmd/devices-api/main.go --parseDependency --parseInternal

docker-tags:
	@echo "latest,$(VER_CUT),$(VER_MAJOR).$(VER_MINOR),$(VER_MAJOR)" > .tags

docker-rc-tags:
	@echo "latest,$(VER_CUT),$(VER_MAJOR)-$(VER_RC)" > .tags

docker-cgo-tags:
	@echo "latest-cgo,$(VER_CUT)-cgo,$(VER_MAJOR).$(VER_MINOR)-cgo,$(VER_MAJOR)-cgo" > .tags

docker: deps
	@docker build -f ./resources/docker/Dockerfile . -t dimozone/devices-api:$(VER_CUT)
	@docker tag dimozone/devices-api:$(VER_CUT) dimozone/devices-api:latest

docker-cgo: deps
	@docker build -f ./resources/docker/Dockerfile.cgo . -t dimozone/devices-api:$(VER_CUT)-cgo
	@docker tag dimozone/devices-api:$(VER_CUT)-cgo dimozone/devices-api:latest-cgo

fmt:
	@go list -f {{.Dir}} ./... | xargs -I{} gofmt -w -s {}
	@go mod tidy

lint:
	@golangci-lint run

test: $(APPS)
	@go test $(GO_FLAGS) -timeout 3m -race ./...
	@$(PATHINSTBIN)/devices-api test ./config/test/...

clean:
	rm -rf $(PATHINSTBIN)
	rm -rf $(DEST_DIR)/dist
	rm -rf $(PATHINSTDOCKER)
