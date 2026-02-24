# ADR 0001 - Protobuf Contract Source Of Truth

## Status

Accepted

## Context

The Jenkins plugin (Kotlin/JVM) and Go client need strict type alignment for request/response payloads. Hand-written JSON structs in separate repos drift quickly.

## Decision

Adopt protobuf as the single source of truth and generate:

1. Java classes for plugin-side request/response handling.
2. Go structs/messages for client-side transport.

Transport remains JSON over HTTP using protobuf JSON mapping to preserve endpoint ergonomics.

## Consequences

1. Stronger server/client type consistency.
2. Additional generation tooling in contributor workflow.
3. Clear migration path for future `v2` contracts.
