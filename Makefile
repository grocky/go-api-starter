PROJECT_NAME := $(shell basename $(CURDIR))
.DEFAULT_GOAL := help

SOURCE_FILES := $(shell find . -name '*.go')

DB_PORT := $(shell docker-compose ps db --format json | jq ".[0].Publishers[] | select(.TargetPort == 3306) | .PublishedPort" || 3306)

help: ## print this help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ { printf "${GREEN}%-20s${RESET}%s\n", $$1, $$NF }' $(MAKEFILE_LIST)

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

.PHONY: tidy
tidy: ## format code and tidy modfile
	go fmt ./...
	go mod tidy -v

.PHONY: audit
audit: ## run quality control checks
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go test -race -vet=off ./...
	go mod verify

# ==================================================================================== #
# BUILD
# ==================================================================================== #
TARGET := bin/$(PROJECT_NAME)-api
$(TARGET): $(SOURCE_FILES)
	go build -ldflags='-s' -o=$@ ./cmd/api

TARGET_LINUX_AMD64 := bin/linux_amd64/$(PROJECT_NAME)-api
$(TARGET_LINUX_AMD64): $(SOURCE_FILES)
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=$@ ./cmd/api

.PHONY: build
build: $(TARGET) $(TARGET_LINUX_AMD64) ## build the cmd/api application

# ==================================================================================== #
# SERVER
# ==================================================================================== #

server-run:
	@DB_HOST=127.0.0.1 \
	DB_PORT=$(DB_PORT) \
	DB_USER=go-api-starter-user \
	DB_PASS=go-api-starter-password \
	go run ./cmd/api

server/run: ## run the server with live reload enabled
	@echo "$(SOURCE_FILES)" | tr " " "\n" | entr -r make server-run

server/run-bin: $(TARGET) ## run the binary
	@DB_HOST=127.0.0.1 \
	DB_PORT=$(DB_PORT) \
	DB_USER=go-api-starter-user \
	DB_PASS=go-api-starter-password \
	./$(TARGET)


# ==================================================================================== #
# SERVER
# ==================================================================================== #

# ==================================================================================== #
# SQL MIGRATIONS
# ==================================================================================== #

db/apply: ## Apply DB schema changes
	DB_PORT=$(DB_PORT) atlas schema apply --env local

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

GREEN  := $(shell tput -Txterm setaf 2)
RESET  := $(shell tput -Txterm sgr0)
