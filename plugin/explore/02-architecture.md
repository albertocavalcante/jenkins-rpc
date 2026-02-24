# 02 - Architecture

## Proposed Components

1. `RpcRootAction` - HTTP entrypoint and protocol routing.
2. `InvocationService` - payload validation, operation policy, and dispatch decisions.
3. `OperationExecutor` - generic direct execution for compatible Jenkins operations.
4. `CpsBridge` (future) - in-pipeline execution path for CPS-bound operations.
5. `ResultStore` - structured result and error payload generation.
6. `AuditRecorder` - security and observability events.

## High-Level Flow

1. Client calls `invoke` with operation + args + run context + idempotency key.
2. Plugin authenticates user and validates policy.
3. Plugin tries direct generic execution lane.
4. If lane is unsupported, plugin returns explicit CPS-boundary error.
5. Plugin records request/result and returns run handle + state.

## Design Principles

1. Narrow API, strict schema.
2. Generic transport, explicit runtime boundary.
3. No unsafe dynamic Groovy evaluation from controller payloads.
4. Full request/response traceability.
