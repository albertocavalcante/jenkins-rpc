# contracts

Shared protobuf contracts for server/client type consistency. Part of the [jenkins-rpc](https://github.com/albertocavalcante/jenkins-rpc) monorepo.

## Commands

- `buf lint`
- `buf generate --template buf.gen.go.yaml`
- `buf generate --template buf.gen.java.yaml`

## Rules

- Edit only `.proto` files manually.
- Regenerate code after every schema change.
- Keep field evolution backward compatible (no tag reuse).
