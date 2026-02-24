# Phase 001 - Schema Baseline

## Goal

Define `contracts.proto` with minimum viable RPC models.

## Scope

- `HealthResponse`
- `CatalogResponse` and `CatalogOperation`
- `InvokeRequest` and `InvokeResponse`
- `RunStatusResponse`
- `Error` and `ErrorResponse`

## Acceptance

1. Tags are unique and stable.
2. Messages support current plugin endpoints.
3. Error schema supports machine-readable codes.
