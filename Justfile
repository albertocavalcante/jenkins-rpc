# jenkins-rpc monorepo
# Run `just` to see all commands

set shell := ["bash", "-cu"]

default:
    @just --list --unsorted

# ─── Setup ───────────────────────────────────────────────

# Initial project setup
setup:
    lefthook install

# ─── Contracts ───────────────────────────────────────────

# Lint proto files
contracts-lint:
    cd contracts && buf lint

# Generate Go protobuf code
contracts-gen-go:
    cd contracts && buf generate --template buf.gen.go.yaml

# Generate Java protobuf code
contracts-gen-java:
    cd contracts && buf generate --template buf.gen.java.yaml

# Generate all protobuf code
contracts-gen: contracts-gen-go contracts-gen-java

# ─── Plugin ──────────────────────────────────────────────

# Build plugin
plugin-build:
    cd plugin && ./gradlew build

# Test plugin
plugin-test:
    cd plugin && ./gradlew test

# Run plugin check
plugin-check:
    cd plugin && ./gradlew check

# ─── Go Client ───────────────────────────────────────────

# Build go-client
go-client-build:
    cd go-client && go build ./...

# Test go-client
go-client-test:
    cd go-client && go test ./...

# Test go-client with race detector
go-client-test-race:
    cd go-client && go test -race ./...

# Lint go-client
go-client-lint:
    cd go-client && GOWORK=off go tool -modfile=tools/lint/go.mod golangci-lint run

# Vet go-client
go-client-vet:
    cd go-client && go vet ./...

# Format check go-client
go-client-fmt-check:
    @test -z "$(cd go-client && gofmt -l .)" || (echo "gofmt needed on:"; cd go-client && gofmt -l .; exit 1)

# Run all go-client checks
go-client-check: go-client-fmt-check go-client-vet go-client-lint go-client-test

# Tidy go modules
go-client-tidy:
    cd go-client && go mod tidy
    cd go-client && go mod tidy -modfile=tools.go.mod
    cd go-client/tools/lint && go mod tidy

# ─── E2E ─────────────────────────────────────────────────

# Copy plugin HPI to e2e docker context
e2e-prepare: plugin-build
    cp plugin/build/libs/jenkins-step-rpc-plugin.hpi tests/e2e/docker/

# Start Jenkins container for e2e tests
e2e-up: e2e-prepare
    cd tests/e2e/docker && docker compose up --build --wait

# Stop Jenkins container and remove volumes
e2e-down:
    cd tests/e2e/docker && docker compose down -v

# Run e2e integration tests (starts Jenkins, runs tests, tears down)
e2e-test: e2e-prepare
    cd tests/e2e && go test -tags e2e -v -count=1 -timeout 300s ./...

# Clean up e2e artifacts
e2e-clean: e2e-down
    rm -f tests/e2e/docker/jenkins-step-rpc-plugin.hpi

# ─── All ─────────────────────────────────────────────────

# Run all checks across the monorepo
check-all: contracts-lint go-client-check plugin-check
