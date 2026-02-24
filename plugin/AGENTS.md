# jenkins-step-rpc-plugin

Kotlin Jenkins plugin scaffold for a step-RPC bridge.

## Commands

- `JAVA_HOME=/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home ./gradlew compileKotlin` - compile plugin sources.
- `JAVA_HOME=/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home ./gradlew test` - run unit tests.
- `JAVA_HOME=/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home ./gradlew server` - run Jenkins with plugin loaded for local testing.

## Structure

- `src/main/kotlin` plugin implementation
- `src/test/kotlin` tests
- `explore` research docs
- `plan` phased plans and risk tracking
- `docs/adr` architecture decisions

## Conventions

- Keep external RPC surface minimal and explicitly allowlisted.
- Never execute arbitrary Groovy from endpoint payloads.
- Treat every RPC request as untrusted input.
- Preserve deterministic JSON contracts for client compatibility.
