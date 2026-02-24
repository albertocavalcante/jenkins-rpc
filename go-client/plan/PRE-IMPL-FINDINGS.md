# PRE-IMPL Findings

## Confirmed

1. `init` is a good fit for this Go client scaffold.
2. Core client responsibilities are invoke + poll + decode + retry policy.
3. Contract evolution must be versioned from the start.

## Open Questions

1. Final endpoint auth mechanism per environment.
2. Whether to expose sync convenience APIs over polling.
3. Minimum plugin version support policy.
