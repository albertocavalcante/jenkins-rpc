# Go Client API Surface

## Constructor

1. `New(baseURL, token string, httpClient *http.Client) (*Client, error)`

## Configuration

1. `WithRetryPolicy(p *RetryPolicy) *Client` — returns a copy with retry enabled
2. `WithDebugHook(h *DebugHook) *Client` — returns a copy with debug callbacks

## Invoke + Status

1. `Invoke(ctx, *steprpcv1.InvokeRequest) (*steprpcv1.InvokeResponse, error)`
2. `GetRunStatus(ctx, runID string) (*steprpcv1.RunStatusResponse, error)`
3. `WaitRunTerminal(ctx, runID string, policy PollPolicy) (*steprpcv1.RunStatusResponse, error)`

Terminal states used by `WaitRunTerminal`:

1. `succeeded`
2. `failed`
3. `cancelled`

## Catalog

1. `GetCatalog(ctx) (*steprpcv1.CatalogResponse, error)`
2. `DirectOperations(catalog) []string`
3. `CPSBridgeOperations(catalog) []string`

Execution lane metadata comes from protobuf `execution_mode` on catalog operations.

## CPS Bridge

1. `GetBridgePending(ctx, runExternalizableID string) (*steprpcv1.BridgePendingResponse, error)`
2. `CompleteBridgeRequest(ctx, req *steprpcv1.BridgeCompleteRequest) (*steprpcv1.BridgeCompleteResponse, error)`

## Errors

Non-2xx responses decode to:

1. `*HTTPError` with `StatusCode` and `Category()` method
2. `ProtoError` (`code`, `message`, `details`) when server returns structured error payload

Error categories (`ErrorCategory`): `Network`, `Auth`, `NotFound`, `BadRequest`, `RateLimited`, `ServerError`, `Unknown`.

Helper: `CategoryOf(err) ErrorCategory` — extracts category from error chain.

## Retry

`RetryPolicy` struct:
- `MaxAttempts` — total attempts (including first try)
- `InitialBackoff` — delay before first retry
- `MaxBackoff` — upper bound on backoff duration
- `Classifier` — optional `func(statusCode int, err error) bool` override

Default classifier retries on 429, 502, 503, 504. Exponential backoff with full jitter. Zero value = no retry.

Wired into `Invoke`, `CompleteBridgeRequest`, `GetRunStatus`, `GetCatalog`, `GetBridgePending`.

## Debug Hooks

`DebugHook` struct:
- `OnRequest(req *http.Request, body []byte)` — called before HTTP send
- `OnResponse(resp *http.Response, body []byte, err error)` — called after response read

Both callbacks are optional (nil-safe).

## Polling

`PollPolicy` struct:
- `InitialInterval` — first poll delay (default 2s)
- `MaxInterval` — upper bound after exponential backoff
- `MaxAttempts` — 0 = unlimited
- `MaxDuration` — 0 = unlimited; sets context deadline

`WaitRunTerminal` applies exponential backoff with ±25% jitter between polls.
