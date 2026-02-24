# jenkins-step-rpc-contracts

Source-of-truth Protobuf contracts for Jenkins Step RPC.

## Purpose

Defines the shared message schema consumed by:

1. `jenkins-step-rpc-plugin` (Jenkins server)
2. `jenkins-step-rpc-go-client` (Go client)

Catalog operations now include typed execution lane metadata via `execution_mode`:

1. `OPERATION_EXECUTION_MODE_DIRECT`
2. `OPERATION_EXECUTION_MODE_CPS_BRIDGE_REQUIRED`

## Layout

- `proto/` canonical `.proto` files
- `gen/go/` generated Go types
- `gen/java/` generated Java types for JVM/Kotlin usage
- `explore/` research notes
- `plan/` delivery phases
- `docs/adr/` architecture decisions

## Regeneration

```bash
buf lint
buf generate --template buf.gen.go.yaml
buf generate --template buf.gen.java.yaml
```
