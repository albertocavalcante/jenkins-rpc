# 04 - Contract Notes

## Initial Request Fields

1. `requestId`
2. `operation`
3. `args`
4. `idempotencyKey` (optional)

## Initial Response Fields

1. `requestId`
2. `runId`
3. `state`
4. `error` (optional)

## Validation Policy

1. Required fields must be non-empty.
2. Unknown operation names are rejected before request dispatch when possible.
3. Unknown response fields should be captured for diagnostics.
