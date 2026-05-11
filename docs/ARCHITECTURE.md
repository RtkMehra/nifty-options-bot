# System Architecture

## Overview

The bot is a pipeline of goroutines communicating via typed channels. No shared mutable state — each layer owns its data and passes it to the next via channels.

```
┌─────────────────────────────────────────────────────────────┐
│                    NIFTY OPTIONS BUYING BOT                  │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
      [NSE API]        [Dhan WebSocket]   [VIX API]
              │               │               │
              ▼               ▼               │
      DataFetcher ───────► MarketSnapshot ◄──┘
              │
              ▼
      SignalEngine ──────────► Signal
              │
              ▼
     StrategyEngine ──────► TradeDecision
              │
              ▼
     ExecutionEngine ─────► Dhan REST API
              │
              ▼
      RiskMonitor ◄──────── Positions
```

## Layer 1 — Data Pipeline

### Data Sources
| Source | Frequency | Used For |
|--------|-----------|----------|
| Nifty 50 Spot (Dhan WS) | Live tick | Delta calc, strike selection |
| Option Chain (NSE API) | Every 5s | IV, Greeks, OI, liquidity |
| India VIX (NSE/Yahoo) | Every 60s | IV rank, regime detection |
| VIX History (SQLite) | Daily | IVR calculation |
| Option Prices (BSM) | Daily backfill | Backtesting |

### DataFetcher Goroutine
- Polls NSE option chain every 5 seconds
- Maintains NSE session with cookie refresh every 20-30 minutes
- Filters rows: OI > 50000, volume > 5000, spread < 5 points
- Computes MidPrice = (Bid + Ask) / 2
- Emits `MarketSnapshot` on channel

## Layer 2 — Signal Engine

### Entry Equation
```
IV Rank < 35  AND  IV Z-Score < -0.5  AND  Directional Conviction > 0.65  AND  DTE in [7, 21]
```

### Signals
1. **IV Rank** (0–100): Current VIX percentile in 52-week window. < 35 = buy zone.
2. **IV Z-Score**: Std deviations from 20-day IV mean. < -0.5 = cheap.
3. **Expected Move**: `spot × IV × √(DTE/365)`. Defines strike selection.
4. **Directional Conviction** (0–1): Combines BOS, OB retest, FVG fill, EMA alignment, PCR extreme, proximity to key level. Minimum 0.65 to trade.
5. **DTE Filter**: 7–14 preferred, 4–7 only with high conviction, 0–3 blocked.

### IV Regime
| IVR Range | Action |
|-----------|--------|
| 0–25 | Strong BUY |
| 25–35 | BUY with confirmation |
| 35–55 | Debit spreads only |
| 55–70 | Avoid buying |
| 70–100 | DO NOT BUY |

## Layer 3 — Strategy Engine

### Seven Strategies (Priority-Ordered)
| Strategy | IVR | Conviction | DTE | Max Capital |
|----------|-----|------------|-----|-------------|
| D — OTM Call/Put | < 20 | ≥ 0.80 | 7–14 | 3% |
| C — Long Straddle | < 25 | ≥ 0 | 5–10 | 8% |
| A — Long Call/Put | < 30 | ≥ 0.70 | 7–14 | 10% |
| B — Bull/Bear Spread | 30–50 | ≥ 0.65 | 10–21 | 12% |

More specific strategies take priority (higher conviction floor = earlier check).

### Strategy Selection Flow
1. `Signal` arrives at `StrategyEngine.Run()`
2. `Select(sig, dte)` iterates strategy matrix in priority order
3. First match by IVR, conviction, DTE, and direction wins
4. `SelectStrikes()` computes exact legs (spread width = 3 × strike step)
5. `CalcEV()` computes expected value with IVR-adjusted win rate and conviction-boosted win multiplier
6. `BuildTradeDecision()` gates on positive EV, then computes lots

### Position Sizing (Half-Kelly)
```
Kelly = (winRate × winMult − (1−winRate) × lossMult) / winMult
HalfKelly = Kelly / 2
Allocation = min(HalfKelly × capital, maxCapByStrategy × capital)
Lots = floor(allocation / (premium × 75))
```

### Expected Value Adjustments
- **Win rate boost**: IVR < 20 → ×1.2 (cap 65%), IVR < 30 → ×1.1 (cap 60%), IVR > 40 → ×0.9
- **Win multiplier boost**: Conviction > 0.85 → ×1.3, Conviction > 0.75 → ×1.1
- Trade proceeds only if `EV = (winRate × avgWin) − (lossRate × avgLoss) > 0`

## Layer 4 — Execution Engine (Dhan API)

### Order Placement
- REST API: `POST /orders` with access token header
- Limit orders at mid-price with 3 retry attempts
- 30s fill wait per attempt, price improvement of 0.5 points per retry
- Multi-leg sequencing: BUY leg first, then SELL leg for spreads
- **Dry-run mode**: null Dhan client, always fills at entry price

### Smart Fill
```
1. Place limit order at mid-price
2. Wait 30s for fill (poll every 500ms)
3. If unfilled or rejected: cancel, improve price by 0.5, retry (max 3)
4. If still unfilled: log, emit RiskAlert, skip
```

### Dhan Client Interface
```
PlaceOrder(OrderRequest) → (OrderResponse, error)
CancelOrder(orderID) → error
GetOrderStatus(orderID) → (status string, error)
```

### Order Tracker
- Thread-safe with `sync.RWMutex`
- Tracks: orderID, leg details, status, timestamps, fill price, strategy type
- Methods: Add(), Update(), GetOpen(), GetByStrategy()

## Layer 5 — Risk Management (Not Yet Implemented)
- `internal/risk/` directory scaffolded
- Position-level checks, portfolio limits, and automated stop logic pending

## Shutdown Sequence
1. OS signal (SIGINT/SIGTERM) received
2. `context.WithCancel` triggers `ctx.Done()`
3. All goroutines receive cancellation via select statements
4. `Fetcher.Run()` calls `defer f.session.Stop()` to stop cookie refresh goroutine
5. Each goroutine completes current iteration, then returns
6. 2-second grace period for cleanup
7. Process exits

## Project Structure

```
nifty-options-bot/
├── cmd/
│   └── bot/
│       └── main.go              # Entry point — wires all goroutines
├── internal/
│   ├── core/
│   │   └── types.go             # All shared types (snapshot, signal, config, etc.)
│   ├── datafeed/
│   │   ├── fetcher.go           # NSE option chain poller (5s)
│   │   ├── dhan_ws.go           # Dhan WebSocket for spot price (scaffolded)
│   │   └── session.go           # NSE cookie session manager (context-aware)
│   ├── signals/
│   │   ├── iv_rank.go           # IVR (52wk percentile) + IV Z-Score
│   │   ├── expected_move.go     # BSM expected move, target strike
│   │   ├── conviction.go        # Directional conviction scorer (SMC/ICT)
│   │   ├── max_pain.go          # Max pain + skew calculator
│   │   └── engine.go            # Combines all signals → Signal struct
│   ├── strategy/
│   │   ├── selector.go          # Priority-ordered strategy matrix
│   │   ├── strikes.go           # Strike computation per strategy type
│   │   ├── sizing.go            # Half-Kelly + lot calculation
│   │   ├── ev.go                # Expected value with IVR/conviction adjustments
│   │   └── engine.go            # Goroutine: signal → select → size → trade
│   ├── execution/
│   │   ├── dhan_client.go       # Dhan REST API wrapper (place/cancel/status)
│   │   ├── smart_fill.go        # Dry-run mode + 3-retry limit fill
│   │   └── order_tracker.go     # Thread-safe order state (sync.RWMutex)
│   ├── risk/                    # Scaffolded (not yet implemented)
│   ├── store/
│   │   ├── db.go                # SQLite with modernc.org/sqlite (pure Go, no CGO)
│   │   ├── vix_history.go       # 52-week VIX store + IV history
│   │   └── trade_log.go         # Trade recording + open trade queries
│   └── notify/                  # Scaffolded (not yet implemented)
├── pkg/
│   └── bsm/
│       └── bsm.go               # Black-Scholes-Merton pricer + IV solver
├── config/
│   └── config.yaml              # API keys, thresholds, risk params
├── docs/
│   ├── VISION.md                # Project scope, roadmap, success metrics
│   ├── ARCHITECTURE.md          # System architecture and data flow
│   ├── RULES.md                 # Development conventions and rules
│   ├── CODING_STANDARDS.md      # Go coding standards
│   └── AGENTS.md                # AI agent guidelines
├── go.mod
└── go.sum
```

## Dependencies

```
go 1.26.3
modernc.org/sqlite                 # Pure Go SQLite (no CGO)
gonum.org/v1/gonum                 # Normal CDF for BSM (optional, custom impl used)
```

Note: No external logging, WebSocket, or Telegram dependencies yet — kept lean initially.

## Shutdown Sequence

1. OS signal (SIGINT/SIGTERM) received
2. `context.WithCancel` triggers `ctx.Done()`
3. All goroutines receive cancellation via select statements
4. `Fetcher.Run()` calls `defer f.session.Stop()` to stop cookie refresh goroutine
5. Each goroutine completes current iteration, then returns
6. 2-second grace period for cleanup
7. Process exits
