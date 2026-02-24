# Contributing to jenkins-rpc

Thank you for your interest in contributing to the Jenkins Step RPC project.

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [JDK 17](https://openjdk.org/)
- [buf](https://buf.build/) (for protobuf generation)
- [just](https://github.com/casey/just) (command runner)

## Getting Started

```bash
# Clone the repository
git clone https://github.com/albertocavalcante/jenkins-rpc.git
cd jenkins-rpc

# Install hooks and tools
just setup

# Run all checks
just check-all
```

## Development Workflow

1. Fork and create a feature branch from `main`.
2. Make your changes in the appropriate module (`contracts/`, `plugin/`, or `go-client/`).
3. Run checks for the module you changed:
   ```bash
   just contracts-lint    # protobuf changes
   just contracts-gen     # regenerate after proto changes
   just plugin-build      # plugin changes
   just plugin-test
   just go-client-check   # go-client changes
   ```
4. Run the full suite before submitting: `just check-all`
5. Open a pull request against `main`.

## Protobuf Contract Changes

If you modify `.proto` files in `contracts/proto/`:

1. Edit only the `.proto` source files.
2. Regenerate Go and Java code: `just contracts-gen`
3. Commit both the `.proto` changes and the regenerated code.
4. Keep field evolution backward compatible — never reuse tag numbers.

## Code Style

- **Go**: Follow standard `gofmt` formatting. The linter config is in `go-client/.golangci.toml`.
- **Kotlin**: Follow standard Kotlin conventions. The plugin uses Gradle for builds.
- **Protobuf**: Follow the [buf style guide](https://buf.build/docs/best-practices/style-guide/). Run `buf lint` before committing.

## Reporting Issues

Open an issue on [GitHub Issues](https://github.com/albertocavalcante/jenkins-rpc/issues) with a clear description of the problem or feature request.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
