# 02 - Client Architecture

## Components

1. `Client` - endpoint methods and auth headers.
2. `Protocol` models - request/response structs with versioning.
3. `Polling` helpers - status loops and timeout handling.
4. `Errors` package - typed transport/protocol/runtime errors.

## Data Flow

1. Build invoke payload.
2. Send request with context.
3. Parse immediate state.
4. Poll status until terminal.
5. Return structured outcome.
