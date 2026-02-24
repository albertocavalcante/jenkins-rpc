# 04 - API Surface

## Candidate Endpoints

1. `POST /step-rpc/v1/invoke`
2. `GET /step-rpc/v1/runs/{runId}`
3. `GET /step-rpc/v1/catalog`
4. `GET /step-rpc/v1/health`

## Invoke Request Draft

```json
{
  "requestId": "uuid",
  "operation": "archiveArtifacts",
  "args": {
    "artifacts": "build/**/*.jar",
    "allowEmptyArchive": false
  },
  "idempotencyKey": "optional-key"
}
```

## Response Draft

```json
{
  "requestId": "uuid",
  "runId": "job#42",
  "state": "queued",
  "links": {
    "status": "/step-rpc/v1/runs/job%2342"
  }
}
```

## Contract Rules

1. Strongly typed JSON schema.
2. Unknown fields rejected.
3. Operation names must be from allowlist.
4. Sensitive fields redacted in logs.
