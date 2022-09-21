# Include variables from the .envrc file
include .envrc


# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run: Runs the server
.PHONY: run
run:
	@go run ./cmd/server


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor


# ==================================================================================== #
# BUILD
# ==================================================================================== #
#

## build/server: build the cmd/server application
.PHONY: build/server
build/server:
	@echo 'Building...'
	go build -ldflags="-X main.port=${PORT} -X main.uri=${URI}" -o=./bin/code-harvest-server ./cmd/server
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.port=${PORT} -X main.uri=${URI}" -o=./bin/darwin_arm64/code-harvest-server ./cmd/server
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.port=${PORT} -X main.uri=${URI}" -o=./bin/linux_amd64/code-harvest-server ./cmd/server

## build/client: build the cmd/client application
.PHONY: build/client
build/client:
	@echo 'Building...'
	go build -ldflags="-X main.port=${PORT} -X main.hostname=${HOSTNAME}" -o=./bin/code-harvest-client ./cmd/client
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.port=${PORT} -X main.hostname=${HOSTNAME}" -o=./bin/darwin_arm64/code-harvest-client ./cmd/client
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.port=${PORT} -X main.hostname=${HOSTNAME}" -o=./bin/linux_amd64/code-harvest-client ./cmd/client

## build: builds the server and client applications
.PHONY: build
build: build/server build/client
