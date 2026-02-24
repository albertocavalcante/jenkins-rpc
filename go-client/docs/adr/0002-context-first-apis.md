# ADR 0002 - Context-First APIs

## Status

Accepted

## Decision

Every network operation requires `context.Context`.

## Rationale

Jenkins RPC calls can block on queue/run lifecycle; callers must be able to cancel and bound runtime.
