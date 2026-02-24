# Risk Register

1. `RISK-001` - Contract drift with plugin responses
- Mitigation: schema fixtures + compatibility tests.

2. `RISK-002` - Duplicate logical operations on retry
- Mitigation: idempotency key guidance + guarded retry policy.

3. `RISK-003` - Poor diagnosability under failures
- Mitigation: typed errors + structured debug hooks.
