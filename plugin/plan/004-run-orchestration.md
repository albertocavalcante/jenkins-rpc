# Phase 004 - Run Context And Execution Lanes

## Goal

Translate accepted RPC requests into execution against explicit build context with clear lane boundaries.

## Tasks

- [ ] Define required run context shape for direct execution lane.
- [ ] Implement generic operation resolution for compatible Jenkins descriptors.
- [ ] Implement explicit CPS-boundary error mapping for unsupported operations.
- [ ] Persist request/result metadata for diagnostics and retries.

## Risks

- Missing/invalid run context from clients.
- False expectation of full `steps.*` emulation outside CPS.
- Duplicate execution when retries race.
