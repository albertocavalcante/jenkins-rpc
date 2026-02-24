# 01 - Problem Space

## Goal

Define a safe and operationally realistic way for non-Jenkins clients to request Pipeline step behavior without opening unsafe execution channels.

## Problem

Pipeline step calls are resolved inside Jenkins Pipeline runtime, with `StepDescriptor`, `StepContext`, CPS thread state, and run-scoped execution requirements. External systems need a supported bridge, not an internal API shortcut.

## Key Constraints

1. Step invocation depends on run context and required environment.
2. Security model must reject arbitrary script execution.
3. Calls must be auditable, replay-aware, and permission-checked.
4. API contracts must remain stable for Go clients.

## Success Criteria

1. Request validation and allowlist are explicit.
2. Every invocation maps to a traceable Jenkins run or node.
3. Endpoint usage requires Jenkins permissions and CSRF-safe handling.
4. Failure modes are deterministic and documented.
