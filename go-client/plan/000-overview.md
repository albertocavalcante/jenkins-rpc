# Phase 000 - Overview

## Objective

Ship a stable Go client for the Jenkins step-RPC plugin with clear contracts, retries, and observability hooks.

## Phases

1. `001-project-setup`
2. `002-protocol-models`
3. `003-invoke-and-status-api`
4. `004-retry-polling-and-timeouts`
5. `005-errors-and-observability`
6. `006-compatibility-and-e2e-tests`

## Quality Gates

1. Context cancellation honored everywhere.
2. Deterministic error taxonomy.
3. Contract tests for plugin response variants.
4. Backward compatibility checks for JSON schema changes.
