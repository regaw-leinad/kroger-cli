# kroger-cli

A Go CLI for the Kroger API. Designed for AI agents to shop for groceries.

## Build & Run

Uses [Task](https://taskfile.dev) for commands:

```bash
task build    # Build the binary
task test     # Run tests
task lint     # Run all checks (vet, fmt, test)
task fmt      # Format code
```

## Project Structure

- `main.go` - entry point
- `cmd/` - Cobra command definitions (auth, store, product, cart)
- `internal/api/` - Kroger API client (HTTP, auth injection, token refresh)
- `internal/auth/` - OAuth2 localhost callback flow
- `internal/secrets/` - Credential and token storage (~/.config/kroger/credentials.json)
- `internal/config/` - Non-sensitive config (~/.config/kroger/config.json)
- `internal/output/` - Printer interface with JSON and human implementations

## Code Style

- Early returns over nested if/else
- Dependencies injected via constructors, no globals
- Errors returned up, formatted once at command layer
- `fmt.Errorf("context: %w", err)` for wrapping
