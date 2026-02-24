# jenkins-step-rpc-plugin

Kotlin-based Jenkins plugin that exposes a constrained HTTP endpoint so external clients (for example Go services) can request pipeline step execution through a controlled Jenkins job/run path.

## Status

Prototype endpoint behavior is implemented for `v1` health/catalog/invoke/run-status with shared protobuf contracts.

## Why This Exists

Pipeline step functions like `archiveArtifacts`, `junit`, `input`, and others are runtime-bound to Jenkins Pipeline execution and cannot be safely called as direct standalone RPC functions from external processes.

This plugin provides:

1. A narrow RPC surface.
2. Generic operation dispatch (`operation` + `args`) with policy controls.
3. Explicit execution-boundary handling for Jenkins CPS runtime.
4. Auditable request and response records.

## Kotlin Choice

This plugin is intentionally Kotlin-first. Jenkins plugins can be authored in Kotlin because plugin binaries are JVM bytecode artifacts, and existing Jenkins plugins in the `jenkinsci` org already use Kotlin.

## Planned Endpoint Model

1. `POST /step-rpc/v1/invoke`
2. `GET /step-rpc/v1/runs/{runId}`
3. `GET /step-rpc/v1/catalog`
4. `GET /step-rpc/v1/bridge/pending?runExternalizableId=<id>`
5. `POST /step-rpc/v1/bridge/complete`

## Critical Constraint

From a plain controller REST action (outside Pipeline CPS context), Jenkins cannot safely emulate arbitrary `steps.*` behavior.

That means:

1. Some operations can be executed directly (for example `SimpleBuildStep`-style operations with explicit run/workspace context).
2. Arbitrary Pipeline steps still require execution inside a running Pipeline CPS context.
3. The plugin must surface this boundary explicitly instead of pretending full Groovy-equivalent dispatch from controller code.

Runtime error semantics:

1. `operation_not_found`: operation not discovered on this controller.
2. `requires_cps_context`: operation exists but cannot run safely from controller lane.

Catalog semantics:

1. `catalog.operations[].executionMode` reports direct lane vs CPS bridge lane.

Bridge lane semantics:

1. CPS-bound operations are queued with state `queued`.
2. A Pipeline-side bridge worker retrieves pending requests and executes them in live CPS context.
3. Bridge worker marks each request complete (`succeeded`/`failed`) through the complete endpoint.

See `docs/bridge-worker-example.md` for a practical Jenkinsfile-side drain pattern.

## Repository Structure

- `src/main/kotlin/` plugin source
- `src/test/kotlin/` plugin tests
- `explore/` research and design exploration
- `plan/` phased delivery plans
- `docs/adr/` architecture decision records

## Next Steps

1. ~~Finalize generic operation resolver and descriptor binding model.~~ Done.
2. ~~Add CPS bridge path for operations that require live Pipeline context.~~ Done.
3. ~~Tighten authz/audit controls per endpoint.~~ Done.
4. Add integration tests with Jenkins test harness.
