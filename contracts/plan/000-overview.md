# Contracts Plan Overview

## Goal

Provide a shared, versioned protobuf contract consumed by both:

1. `jenkins-step-rpc-plugin`
2. `jenkins-step-rpc-go-client`

## Phases

1. Schema baseline and governance rules.
2. Multi-language code generation.
3. Consumption wiring in plugin/client.
4. CI guards for compatibility.

## Exit Criteria

1. Generated Go and Java types compile in downstream repos.
2. Endpoint/request handling uses generated message types.
3. Lint and generation are deterministic.
