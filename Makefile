# Include variables from the .envrc file
include .envrc

.DEFAULT_GOAL := build

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
.PHONY:help

confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]
.PHONY:confirm

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run: Runs the server
run:
	@go run ./cmd/server
.PHONY:run

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy and vendor dependencies and format, vet and test all code
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...
.PHONY:audit

## vendor: tidy and vendor dependencies
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor
.PHONY:vendor


# ==================================================================================== #
# BUILD
# ==================================================================================== #
#

## build/server: build cmd/server
build/server:
	@echo 'Compiling server...'
	go build -ldflags="-X main.serverName=${SERVER_NAME} -X main.port=${PORT}" -o=./bin/code-harvest-server ./cmd/server
.PHONY:build/server

## build/client: build cmd/client
build/client:
	@echo 'Compiling client...'
	go build -ldflags="-X main.serverName=${SERVER_NAME} -X main.port=${PORT} -X main.hostname=${HOSTNAME}" -o=./bin/code-harvest-client ./cmd/client
.PHONY:build/client

## build/aggregate: build cmd/aggregate
build/aggregate:
	@echo 'Compiling aggregate...'
	go build -ldflags="-X main.uri=${URI} -X main.db=${DB}" -o=./bin/code-harvest-aggregate ./cmd/aggregate
.PHONY:build/client

## build: builds the server and client applications
build: audit build/server build/client build/aggregate
.PHONY:build
