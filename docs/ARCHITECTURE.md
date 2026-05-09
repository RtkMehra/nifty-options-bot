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

### Four Strategies
| Strategy | IVR | Conviction | DTE | Max Capital |
|----------|-----|------------|-----|-------------|
| A — Long Call/Put | < 30 | > 0.70 | 7–14 | 8–10% |
| B — Debit Spread | 30–50 | > 0.65 | 10–21 | 10–12% |
| C — Long Straddle | < 25 | Any (event) | 5–10 | 6–8% |
| D — OTM Momentum | < 20 | > 0.80 | 7–14 | 3% |

### Position Sizing (Half-Kelly)
```
Kelly = (winRate × winMult − (1−winRate) × lossMult) / winMult
HalfKelly = Kelly / 2
Allocation = min(HalfKelly, maxCapByStrategy) × capital
Lots = floor(allocation / (premium × 75))
```

### Expected Value Gate
Trade proceeds only if:
```
EV = (winRate × avgWin) − (lossRate × avgLoss) > 0
```

## Layer 4 — Execution Engine (Dhan API)

### Order Placement
- REST API: `POST /orders` with access token header
- Limit orders at mid-price with 3 retry attempts
- 30s fill wait per attempt, price improvement of 0.5 points per retry
- Multi-leg sequencing: BUY leg first, then SELL leg for spreads

### Smart Fill
```
1. Place limit order at mid-price
2. Wait 30s for fill
3. If unfilled: cancel, improve price by 0.5, retry (max 3 attempts)
4. If still unfilled: log and skip
```

## Layer 5 — Risk Management

### Position-Level Rules
| Rule | Implementation |
|------|---------------|
| Hard Stop Loss | Exit at −50% of premium paid |
| Profit Target (1) | Exit 50% at +100% gain |
| Profit Target (2) | Trail remaining 50% at 50% of peak gain |
| Time Stop | Exit at 3 DTE remaining |
| Delta Stop | Exit if delta < 0.15 |

### Portfolio-Level Rules
- Max 3 simultaneous open trades
- Max 20% capital at risk across all positions
- Max 3% daily loss → block new entries
- Max 7% weekly loss → reduce sizing 50%
- Correlation limit: no long call + long straddle simultaneously
- 2-hour cooldown after any stop loss

## Project Structure

```
nifty-options-bot/
├── cmd/
│   └── bot/
│       └── main.go              # Entry point — wires all goroutines
├── internal/
│   ├── datafeed/
│   │   ├── fetcher.go           # NSE option chain poller (5s)
│   │   ├── dhan_ws.go           # Dhan WebSocket for spot price
│   │   └── session.go           # NSE cookie session manager
│   ├── signals/
│   │   ├── iv_rank.go           # IVR + IV Z-Score
│   │   ├── expected_move.go     # BSM expected move
│   │   ├── conviction.go        # Directional conviction scorer
│   │   ├── max_pain.go          # Max pain calculator
│   │   └── engine.go            # Combines signals → Signal struct
│   ├── strategy/
│   │   ├── selector.go          # Picks strategy A/B/C/D
│   │   ├── strikes.go           # Computes strikes to trade
│   │   ├── sizing.go            # Kelly sizing + lots
│   │   └── ev.go                # Expected value gate
│   ├── execution/
│   │   ├── dhan_client.go       # Dhan REST API wrapper
│   │   ├── smart_fill.go        # Limit order + retry
│   │   └── order_tracker.go     # Live order state
│   ├── risk/
│   │   ├── monitor.go           # Position-level checks
│   │   └── portfolio.go         # Portfolio-level limits
│   ├── store/
│   │   ├── db.go                # SQLite connection
│   │   ├── vix_history.go       # 52-week VIX store
│   │   └── trade_log.go         # Trade recording
│   └── notify/
│       └── telegram.go          # Trade alerts
├── pkg/
│   └── bsm/
│       └── bsm.go               # Black-Scholes-Merton pricer
├── config/
│   └── config.yaml              # API keys, thresholds, lot size
└── go.mod
```

## Dependencies

```
go 1.22
github.com/mattn/go-sqlite3       # SQLite
github.com/gorilla/websocket       # Dhan WS
gopkg.in/yaml.v3                   # Config
go.uber.org/zap                    # Logging
github.com/go-telegram-bot-api/telegram-bot-api/v5  # Alerts
gonum.org/v1/gonum                 # Normal CDF for BSM
```

## Shutdown Sequence

1. OS signal (SIGINT/SIGTERM) received
2. `context.WithCancel` triggers `ctx.Done()`
3. All goroutines receive cancellation via select
4. Each goroutine completes current iteration, then returns
5. 2-second grace period for cleanup
6. Process exits
