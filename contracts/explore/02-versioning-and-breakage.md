# Explore 02 - Versioning And Breakage Controls

## Versioning Model

- Package version is encoded in proto namespace: `steprpc.v1`.
- HTTP route version remains aligned: `/step-rpc/v1/...`.

## Compatibility Rules

1. Never reuse field numbers.
2. Additive fields only for minor contract expansion.
3. Removal requires deprecation window and major namespace bump (`v2`).

## Validation Controls

- `buf lint` for style/consistency checks.
- `buf breaking` (future CI step) to block incompatible schema edits.

## Notes

The first working baseline includes health, catalog, invoke, run status, and structured error payloads.
