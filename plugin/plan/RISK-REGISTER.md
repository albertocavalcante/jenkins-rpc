# Risk Register

1. `RISK-001` - Over-broad operation exposure
- Mitigation: explicit allowlist and per-op validator.

2. `RISK-002` - Duplicate execution on retries
- Mitigation: idempotency keys + request store.

3. `RISK-003` - Plugin/version drift across Jenkins instances
- Mitigation: `/catalog` endpoint with capability flags.

4. `RISK-004` - Sensitive data leakage in logs
- Mitigation: field-level redaction and structured audit policy.
