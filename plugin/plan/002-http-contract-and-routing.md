# Phase 002 - HTTP Contract And Routing

## Goal

Implement stable JSON endpoints and deterministic error mapping.

## Tasks

- [x] Implement `/v1/invoke` request parsing.
- [x] Implement `/v1/runs/{runId}` status endpoint.
- [x] Implement `/v1/catalog` operation discovery endpoint.
- [x] Add structured error payload format.

## Tests

1. `InvokeRejectsMalformedJson`
2. `InvokeRejectsUnknownOperation`
3. `StatusReturnsKnownRunState`
