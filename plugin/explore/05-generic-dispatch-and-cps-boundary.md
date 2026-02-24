# 05 - Generic Dispatch And CPS Boundary

## User Goal

Allow non-Groovy clients to request Jenkins step behavior using one generic call shape:

1. `operation` as a string.
2. `args` as structured payload.

## Key Constraint

From a plain controller endpoint, we do not have live Pipeline CPS context. Therefore we cannot fully reproduce arbitrary `steps.*` behavior with correctness and security guarantees.

## Tradeoffs

### Option A - Hardcoded typed handlers

1. Strongly typed and predictable.
2. Requires per-step maintenance and schema churn.
3. Poor extensibility as plugins evolve.

### Option B - Fully dynamic controller-side invocation

1. Looks flexible.
2. Violates CPS/runtime assumptions for many steps.
3. Higher risk for incorrect behavior and security issues.

### Option C - Generic protocol + explicit runtime boundary (chosen)

1. Keep protocol generic and stable.
2. Execute only operations safe in direct controller lane.
3. Queue CPS-bound operations and execute via bridge lane in live Pipeline context.
4. Keep explicit errors when operation is missing or bridge cannot serve.

## Assumptions For Current Phase

1. Caller can provide run context (`runExternalizableId` or `job/build` + node/workspace).
2. Direct lane targets operations that can run as `SimpleBuildStep`-style executions.
3. Unsupported operations fail clearly with machine-readable error codes:
`operation_not_found` when absent, `requires_cps_context` when installed but CPS-bound.

## Bridge Handshake

1. `invoke` queues CPS-bound requests with state `queued`.
2. Pipeline-side bridge polls `/bridge/pending`.
3. Bridge executes requested step in CPS context.
4. Bridge reports result to `/bridge/complete`, which updates run status.

## Slide-Friendly Summary

1. Generic API shape is feasible now.
2. Full Groovy-equivalent runtime is not feasible from controller alone.
3. Two-lane model gives immediate value without hiding architectural limits.
