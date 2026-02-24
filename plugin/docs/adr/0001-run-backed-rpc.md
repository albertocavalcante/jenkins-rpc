# ADR 0001 - Run-Aware RPC Execution

## Status

Accepted

## Context

Pipeline steps require Jenkins runtime context and cannot be safely treated as free-standing RPC methods.

## Decision

Treat execution as run-aware rather than run-spawning by default:

1. Prefer operating against an existing build context when provided.
2. Keep run metadata for traceability and status retrieval.
3. Avoid pretending that controller code can fully reproduce CPS semantics.

## Consequences

1. Better alignment with Jenkins execution model.
2. Requires explicit run context from callers for direct execution modes.
3. Enables gradual CPS bridge support for unsupported operations.
