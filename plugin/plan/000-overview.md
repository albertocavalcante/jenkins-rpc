# Phase 000 - Overview

## Objective

Deliver a production-grade Jenkins plugin that provides a constrained RPC bridge for approved pipeline step operations.

## Phase Graph

1. `001-project-setup`
2. `002-http-contract-and-routing`
3. `003-operation-registry-and-validation`
4. `004-run-orchestration`
5. `005-security-and-audit`
6. `006-test-harness-and-e2e`

## Quality Gates

1. No arbitrary script execution path.
2. Request schema validation is exhaustive.
3. Every operation is permission-gated and audited.
4. Integration tests verify success and rejection cases.
