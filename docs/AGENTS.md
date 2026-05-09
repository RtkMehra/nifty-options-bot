# AI Agent Guidelines for This Codebase

This file provides context for AI coding assistants working on the Nifty Options Buying Bot.

## Project Context

- **Language**: Go 1.22+
- **Architecture**: Goroutine pipeline with typed channels — no shared mutable state
- **Broker API**: Dhan REST + WebSocket
- **Data Source**: NSE option chain API (public, no auth needed, cookie-based session)
- **Database**: SQLite via `go-sqlite3`
- **Config**: YAML for parameters, env vars for secrets
- **Build**: Single binary deployment (`go build -o bot cmd/bot/main.go`)

## How to Read This Codebase

Start with the following files in order:
1. `docs/VISION.md` — project goals and scope
2. `docs/ARCHITECTURE.md` — system architecture and data flow
3. `cmd/bot/main.go` — entry point that wires all goroutines
4. `internal/datafeed/fetcher.go` — data ingestion layer
5. `internal/signals/engine.go` — signal generation
6. `internal/strategy/selector.go` — strategy selection

## Repository Structure

```
cmd/bot/main.go           # Entry point
internal/datafeed/        # NSE + Dhan data ingestion
internal/signals/         # IV analysis + directional conviction
internal/strategy/        # Strategy selection + sizing
internal/execution/       # Dhan order execution
internal/risk/            # Position & portfolio risk
internal/store/           # SQLite persistence
internal/notify/          # Telegram alerts
pkg/bsm/                  # Black-Scholes pricer
config/                   # YAML configuration
```

## Commands

```bash
go build ./...          # Build all packages
go vet ./...            # Static analysis
go test ./...           # Run all tests
go test ./... -race     # Race detection
go test -cover ./...    # Coverage report
gofmt -w .              # Format all code
go mod tidy             # Clean dependencies
```

## Agent Rules

1. **Read before writing** — always read existing files in the relevant package to understand patterns before making changes
2. **Match existing style** — follow the naming, formatting, and patterns in adjacent files
3. **No generated secrets** — never generate API keys, tokens, or credentials. Use placeholder comments like `// TODO: read from env`
4. **Test with mocks** — never write tests that hit external APIs. Use the mock interface pattern from `internal/testutil`
5. **Channel ownership** — the creating goroutine closes the channel. Never close a channel you didn't create
6. **Magic numbers** — any number that isn't 0, 1, or 2 must be a named constant
7. **Commit discipline** — only commit when explicitly asked. Keep commits atomic.
8. **Error wrapping** — always wrap errors with `fmt.Errorf("context: %w", err)`
9. **No comments that explain "what"** — code should be self-documenting. Comments explain "why"
10. **Context everywhere** — every goroutine function must accept `context.Context`

## If Unsure

Ask before:
- Adding a new dependency
- Changing the project structure
- Modifying the config schema
- Changing the channel data flow between layers
- Adding a new strategy type
