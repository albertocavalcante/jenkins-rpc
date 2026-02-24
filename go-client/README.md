# go-client

[![CI](https://github.com/albertocavalcante/jenkins-rpc/actions/workflows/ci.yml/badge.svg)](https://github.com/albertocavalcante/jenkins-rpc/actions/workflows/ci.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=albertocavalcante_jenkins-rpc_go-client&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=albertocavalcante_jenkins-rpc_go-client)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=albertocavalcante_jenkins-rpc_go-client&metric=coverage)](https://sonarcloud.io/summary/new_code?id=albertocavalcante_jenkins-rpc_go-client)
[![Go Reference](https://pkg.go.dev/badge/github.com/albertocavalcante/jenkins-rpc/go-client.svg)](https://pkg.go.dev/github.com/albertocavalcante/jenkins-rpc/go-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/albertocavalcante/jenkins-rpc/go-client)](https://goreportcard.com/report/github.com/albertocavalcante/jenkins-rpc/go-client)

Typed Go client for the plugin endpoint family.

## Status

Prototype transport implemented. Invoke, run status, catalog, and CPS bridge APIs all use shared protobuf contracts.

## Scope

1. Build and send invoke requests.
2. Poll run status with timeout/cancellation (`WaitRunTerminal`).
3. Discover operation execution lanes (direct vs CPS bridge required).
4. Drive bridge worker loops for CPS-bound operations.
5. Decode structured error responses.
6. Support idempotency keys and correlation IDs.

## Repository Layout

- `internal/rpcclient/` transport and protocol client primitives
- `docs/api-surface.md` current client methods and error model
- `explore/` research notes
- `plan/` phased implementation plan
- `docs/adr/` architecture decisions

## Next Steps

1. ~~Add status polling API (`runs/{runId}`) with context-aware backoff.~~ Done.
2. ~~Add richer error mapping and retry policies by error code.~~ Done.
3. ~~Add end-to-end fixture tests against plugin responses.~~ Done.
4. Replace local module wiring with tagged contracts releases.
