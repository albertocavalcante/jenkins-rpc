# jenkins-step-rpc-go-client

Go client scaffold for Jenkins step-RPC plugin integration.

## Commands

- `just setup` - install hooks and local tooling.
- `just check` - run fmt, lint, vet, and tests.
- `just test` - run test suite.

## Structure

- `internal/rpcclient` typed client and protocol models
- `explore` design exploration
- `plan` phased delivery docs
- `docs/adr` architecture decisions

## Conventions

- Keep API structs explicit and versioned.
- Do not silently ignore unknown response fields.
- Propagate context cancellation to all HTTP operations.
- Keep retry logic deterministic and observable.
