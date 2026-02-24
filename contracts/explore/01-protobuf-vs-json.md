# Explore 01 - Protobuf Vs Plain JSON

## Context

The plugin and Go client already communicate over HTTP JSON. The missing piece is a shared schema that stays consistent as both repos evolve independently.

## Why Protobuf Here

1. One canonical contract (`contracts.proto`) drives both generated Go and Java types.
2. We still keep human-readable JSON over HTTP using Proto JSON mapping.
3. Backward-compatible field evolution is easier to enforce with tag discipline.

## Risks

1. Added build complexity (Buf + generated code lifecycle).
2. Developers must understand Proto JSON behavior (`snake_case` fields map to `camelCase` JSON by default).
3. Server/client must be pinned to compatible generated revisions.

## Decision

Use protobuf as the schema source of truth and keep transport as JSON for Jenkins endpoint ergonomics.
