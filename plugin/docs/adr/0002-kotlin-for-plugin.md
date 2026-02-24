# ADR 0002 - Kotlin For Plugin Implementation

## Status

Accepted

## Context

Team preference is Kotlin, but Jenkins plugin compatibility is mandatory.

## Decision

Use Kotlin for plugin code with standard Jenkins `hpi` packaging and plugin parent/tooling.

## Consequences

1. Kotlin ergonomics and consistency with other workspace projects.
2. Need to keep compatibility checks in CI for plugin ecosystem updates.
