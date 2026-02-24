# jenkins-rpc

Monorepo for the Jenkins Step RPC system.

## Modules

| Module | Description |
|--------|-------------|
| `contracts/` | Protobuf definitions and generated Go/Java types |
| `plugin/` | Kotlin Jenkins plugin exposing constrained HTTP RPC endpoints |
| `go-client/` | Typed Go client for the plugin endpoint family |

## Quick Start

```bash
# Install hooks and tools
just setup

# Run all checks
just check-all
```

## Module Commands

```bash
# Contracts
just contracts-lint
just contracts-gen

# Plugin
just plugin-build
just plugin-test

# Go Client
just go-client-build
just go-client-test
just go-client-lint
```

## Architecture

External Go services invoke pipeline step execution through the plugin's constrained HTTP endpoints. Shared protobuf contracts ensure type consistency between the Kotlin plugin and the Go client.

Operations are classified into two execution lanes:
1. **Direct** — `SimpleBuildStep`-style operations executable from the controller
2. **CPS Bridge** — Operations requiring a live Pipeline CPS context, queued and executed by a bridge worker

See module-level `docs/` and `plan/` directories for detailed architecture decisions and delivery plans.

## License

[MIT](LICENSE)
