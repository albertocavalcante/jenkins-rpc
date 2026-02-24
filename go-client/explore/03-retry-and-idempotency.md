# 03 - Retry And Idempotency

## Objective

Prevent duplicate logical operations while preserving resilience for transient failures.

## Strategy Draft

1. Client always sends `requestId` and optional `idempotencyKey`.
2. Retries only for network errors and 5xx responses.
3. No automatic retry for schema errors, 4xx authorization errors, or explicit plugin rejection.
4. Polling uses bounded exponential backoff with jitter.
