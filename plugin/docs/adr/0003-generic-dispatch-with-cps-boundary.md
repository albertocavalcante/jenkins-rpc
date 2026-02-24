# ADR 0003 - Generic Dispatch With CPS Boundary

## Status

Accepted

## Context

We want language-agnostic invocation (`operation` + `args`) so clients do not require per-plugin hardcoded contracts.

At the same time, Jenkins Pipeline steps are CPS/runtime-bound. A plain controller REST action does not have a live `CpsScript`/`StepContext` and therefore cannot safely emulate arbitrary `steps.*` calls.

## Decision

Adopt a two-lane execution model:

1. Generic direct lane for operations that can be safely executed from controller-side code using explicit run/workspace context.
2. CPS bridge lane for operations that require live Pipeline CPS semantics.

`invoke` stays generic at protocol level, but runtime decides whether the operation is executable in the direct lane.

## Why

1. Keeps client protocol stable and generic.
2. Makes the execution boundary explicit and auditable.
3. Avoids unsafe “magic Groovy emulation” from controller code.

## Consequences

1. Not all Jenkins steps are directly executable from the controller lane.
2. Errors like `requires_cps_context` are expected and correct for unsupported operations.
3. A follow-up CPS bridge component is required for full `steps.*`-equivalent behavior.

## Alternatives Considered

1. Hardcoded operation handlers per plugin.
Result: type-safe but high maintenance and poor extensibility.

2. Attempt to dynamically call arbitrary step descriptors from controller.
Result: appears flexible but breaks CPS assumptions and causes correctness/security risks.
