# jenkins-rpc

Monorepo for Jenkins Step RPC system: protobuf contracts, Kotlin plugin, and Go client.

## Structure

- `contracts/` — Protobuf definitions + generated Go/Java code
- `plugin/` — Kotlin Jenkins plugin (HTTP RPC endpoints)
- `go-client/` — Go client library

## Commands

- `just check-all` — run all checks across the monorepo
- `just contracts-lint` — lint proto files
- `just contracts-gen` — regenerate Go + Java from proto
- `just plugin-build` — build Kotlin plugin
- `just go-client-check` — fmt, vet, lint, test the Go client

## Conventions

- Edit only `.proto` files manually; regenerate code after every schema change.
- Keep field evolution backward compatible (no tag reuse).
- Keep external RPC surface minimal and explicitly allowlisted.
- Keep API structs explicit and versioned.
- Propagate context cancellation to all HTTP operations.
