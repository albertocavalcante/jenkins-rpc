# Phase 007 - CPS Bridge And Generic Dispatch

## Goal

Ship a generic invocation model now while preparing a robust CPS bridge for full Pipeline-step compatibility.

## Milestones

1. Generic direct lane:
Implement runtime resolution and execution for operations that are safe from controller context.

2. Boundary semantics:
Return deterministic error codes for operations requiring CPS context.

3. CPS bridge contract:
Define and implement plugin-to-pipeline handshake for in-build execution path:
- `GET /v1/bridge/pending`
- `POST /v1/bridge/complete`

4. End-to-end validation:
Add harness tests for direct-lane success, missing plugin, bad args, and CPS-boundary failures.

## Non-Goals (This Phase)

1. Full arbitrary Groovy execution from controller payloads.
2. Bypassing Jenkins security and script-safety constraints.
