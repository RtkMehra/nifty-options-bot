# AI Agent Guidelines for This Codebase

This file provides context for AI coding assistants working on the Nifty Options Buying Bot.

## Project Context

- **Language**: Go 1.26.3
- **Architecture**: Goroutine pipeline with typed channels — no shared mutable state
- **Broker API**: Dhan REST API (client implemented, dry-run mode active)
- **Data Source**: NSE option chain API (public, cookie-based session with 25min refresh)
- **Database**: SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **Config**: YAML for parameters (`config/config.yaml`), env vars for secrets
- **Build**: Single binary deployment (`go build -o bot.exe ./cmd/bot`)
- **Current Phase**: P4 complete — full pipeline wired in dry-run mode

## Pipeline (Wired in main.go)

```
NSE API → DataFetcher (5s) → MarketSnapshot → SignalEngine → Signal
  → StrategyEngine → TradeDecision → ExecutionEngine (dry-run) → RiskAction
```

All goroutines use `context.Context` for lifecycle management. The `Fetcher` also stops the session cookie refresh goroutine via `defer f.session.Stop()`.

## How to Read This Codebase

Start with the following files in order:
1. `docs/VISION.md` — project goals, scope, roadmap
2. `docs/ARCHITECTURE.md` — system architecture and data flow
3. `internal/core/types.go` — all shared data structures
4. `cmd/bot/main.go` — entry point that wires all goroutines
5. `internal/datafeed/fetcher.go` — data ingestion layer
6. `internal/signals/engine.go` — signal generation
7. `internal/strategy/selector.go` — strategy selection matrix
8. `internal/strategy/ev.go` — expected value gate + sizing
9. `internal/execution/dhan_client.go` — broker integration

## Repository Structure

```
cmd/bot/main.go                     # Entry point — wires 4 goroutines
internal/
├── core/types.go                   # Shared types (MarketSnapshot, Signal, Config, etc.)
├── datafeed/
│   ├── fetcher.go                  # NSE option chain poller (5s interval, context-cancelable)
│   ├── session.go                  # NSE cookie session (25min auto-refresh, stop channel)
│   └── dhan_ws.go                  # Scaffolded Dhan WebSocket
├── signals/
│   ├── iv_rank.go                  # IVR (52wk percentile) + IV Z-Score
│   ├── expected_move.go            # BSM expected move, target/stop strikes
│   ├── conviction.go               # Directional conviction (SMC/ICT: BOS, OB, FVG, PCR, EMA)
│   ├── max_pain.go                 # Max pain + skew calculator
│   ├── engine.go                   # Combines signals, interfaces with store
├── strategy/
│   ├── selector.go                 # Priority-ordered strategy matrix (7 strategies)
│   ├── strikes.go                  # Strike computation per strategy
│   ├── sizing.go                   # Half-Kelly, lot calc, max loss
│   ├── ev.go                       # Expected value with IVR/conviction adjustments
│   ├── engine.go                   # Goroutine: signal → select → size → trade decision
├── execution/
│   ├── dhan_client.go              # Dhan REST (place/cancel/status, access-token auth)
│   ├── smart_fill.go               # 3-retry limit fill + dry-run mode
│   ├── order_tracker.go            # Thread-safe order state (sync.RWMutex)
├── store/
│   ├── db.go                       # SQLite init + migrations (modernc.org/sqlite)
│   ├── vix_history.go              # 52wk VIX store + 20d history
│   ├── trade_log.go                # Trade recording + open trade queries
├── risk/                           # Scaffolded (P5)
├── notify/                         # Scaffolded (P5)
pkg/bsm/bsm.go                      # Black-Scholes-Merton pricer + IV solver
config/config.yaml                  # Parameters and thresholds
docs/                               # VISION, ARCHITECTURE, RULES, CODING_STANDARDS, AGENTS
```

## Commands

```powershell
go build ./...          # Build all packages
go vet ./...            # Static analysis
go test ./...           # Run all tests (~100 tests)
go test ./... -v        # Verbose output
go test -cover ./...    # Coverage report
go run ./cmd/bot        # Run the bot pipeline
go mod tidy             # Clean dependencies
```

Note: `-race` flag requires CGO which is unavailable on this Windows environment. The code is designed to be race-free (channels only, no shared mutex state beyond tracker/store).

## Strategy Selection Matrix (Priority Order)

| Priority | Strategy | IVR | Conviction | DTE | Max Capital |
|----------|----------|-----|------------|-----|-------------|
| 1 | OTM Call/Put | < 20 | ≥ 0.80 | 7–14 | 3% |
| 2 | Long Straddle | < 25 | ≥ 0 | 5–10 | 8% |
| 3 | Long Call/Put | < 30 | ≥ 0.70 | 7–14 | 10% |
| 4 | Bull/Bear Spread | 30–50 | ≥ 0.65 | 10–21 | 12% |

## Agent Rules

1. **Read before writing** — always read existing files in the relevant package to understand patterns before making changes
2. **Match existing style** — follow the naming, formatting, and patterns in adjacent files
3. **No generated secrets** — never generate API keys, tokens, or credentials. Use placeholder comments or `os.Getenv()`
4. **Test with mocks** — never write tests that hit real external APIs. Mock HTTP servers (`httptest`) for API tests, mock `Store` interface for signal tests
5. **Channel ownership** — the creating goroutine initializes the channel. The sender closes it when done. Never close a channel you didn't create
6. **Magic numbers** — any number that isn't 0, 1, or 2 must be a named constant or config value
7. **Commit discipline** — only commit when explicitly asked. Keep commits atomic.
8. **Error wrapping** — always wrap errors with `fmt.Errorf("context: %w", err)`
9. **No comments that explain "what"** — code should be self-documenting. Comments explain "why" (design decisions, tradeoffs)
10. **Context everywhere** — every goroutine function must accept `context.Context` and respect cancellation via `select { case <-ctx.Done(): return }`

## If Unsure

Ask before:
- Adding a new dependency (currently only `modernc.org/sqlite` and `gonum.org/v1/gonum`)
- Changing the project structure
- Modifying the config schema
- Changing the channel data flow between layers
- Adding a new strategy type or modifying the matrix

## If Unsure

Ask before:
- Adding a new dependency
- Changing the project structure
- Modifying the config schema
- Changing the channel data flow between layers
- Adding a new strategy type
