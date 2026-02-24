# ADR 0001 - Versioned Protocol Models

## Status

Accepted

## Decision

All request and response types are namespaced by API version, starting with `v1`.

## Rationale

Versioned models reduce coupling and allow explicit migration handling.
