# Phase 002 - Codegen And Release Workflow

## Goal

Generate and publish language bindings consistently.

## Tasks

- [x] Add Buf config and generation templates for Go and Java.
- [x] Commit generated artifacts under `gen/go` and `gen/java`.
- [x] Add CI step: `buf lint` + `buf generate` + clean-tree check.
- [x] Define tagging/release flow for downstream consumers.

## Notes

During local workspace development, downstream modules can use local path wiring (`replace` in Go and Gradle sourceSet for plugin).
