##############
# DEPRECATED #
# ############

SHELL=/usr/bin/env bash -e -o pipefail
PWD = $(shell pwd)
GO_BUILD= go build
GOFLAGS= CGO_ENABLED=0

PG_DB=rmx-dev-test
PG_USER=rmx
PG_PASSWORD=postgrespw
PG_HOST=localhost
PG_PORT=5432

PG_CONN_STRING="postgresql://$(PG_USER):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DB)?sslmode=disable"

PG_CONTAINER_NAME=postgres-rmx

## test: Run tests
.PHONY: test
test:
	go test -race -v ./...

## cover: Run tests and show coverage result
.PHONY: cover
cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

## tidy: Cleanup and download missing dependencies
.PHONY: tidy
tidy:
	go mod tidy
	go mod verify

## vet: Examine Go source code and reports suspicious constructs
.PHONY: vet
vet:
	go vet ./...

## fmt: Format all go source files
.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: install_deps
install_deps:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: migration_generate
migration_generate:
	migrate create -dir internal/db/migrations -ext sql schema

.PHONY: migrate_up
.migrate_up:
	migrate -path internal/db/migrations -database $(PG_CONN_STRING) -verbose up

.PHONY: migrate_down
.migrate_down:
	migrate -path internal/db/migrations -database $(PG_CONN_STRING) -verbose down


