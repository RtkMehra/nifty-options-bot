# Nifty Options Buying Bot — Vision & Scope

## Vision
Build a fully automated, capital-efficient Nifty 50 options buying bot that systematically captures asymmetric upside using pure long option strategies — no naked short positions, no margin risk, no blowup potential.

## Core Philosophy
**"Right big, wrong small."** Win rate of 40–55% is acceptable because winning trades run to 100–200% while losing trades are cut at 50%. The mathematical edge comes from asymmetric R-multiple (>2.0), not prediction accuracy.

## Scope

### In Scope
- Pure option buying only (Long Call, Long Put, Debit Spreads, Long Straddle)
- Nifty 50 index options on NSE
- Directional signals using SMC/ICT concepts (BOS, OB, FVG, EMA alignment)
- IV-based entry filters (IV Rank, IV Z-Score)
- 7–21 DTE sweet spot window
- Dhan API for order execution
- NSE option chain polling every 5 seconds
- Live WebSocket price feed via Dhan
- Risk management: hard stops, profit targets, time stops, delta stops
- Portfolio-level limits: max 3 open trades, 20% capital at risk
- Kelly-based position sizing (half-Kelly)
- SQLite-backed trade log and VIX history
- Telegram alerts for trade events
- Walk-forward backtesting framework

### Out of Scope (v1)
- Option selling / writing strategies
- Futures or equity trading
- Multi-leg exotic options (iron condors, butterflies)
- Machine learning / AI price prediction
- Real-time charting or GUI dashboard (CLI-only initially)
- Broker other than Dhan
- Markets other than Nifty 50 (Bank Nifty, Finnifty, etc.)
- 0–3 DTE trading
- High-frequency trading / market making
- Sentiment analysis from news or social media

## Success Metrics

| Metric | Target |
|--------|--------|
| Win Rate | 40–55% |
| Average R-Multiple | > 2.0 |
| Profit Factor | > 1.8 |
| Max Consecutive Losses | Plan for 8–10 |
| Monthly Return (scaled) | 5–12% (after paper trading validation) |
| Sharpe Ratio | > 1.0 |

## Build Roadmap

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| P1 — Foundation | 1 week | Project structure, types, NSE fetcher, SQLite store |
| P2 — Signal Engine | 1 week | IVR, Z-Score, conviction scorer, unit tests |
| P3 — Backtester | 2 weeks | BSM pricer, walk-forward engine, 4 strategy backtest |
| P4 — Strategy + Execution | 1 week | Strategy selector, Dhan client, smart fill, dry-run |
| P5 — Risk + Alerting | 1 week | Risk monitor, portfolio limits, Telegram alerts |
| P6 — Paper Trading | 4 weeks | Full live run, simulated fills, track P&L |
| P7 — Live Micro | 4–8 weeks | 1 lot per trade, max ₹15k risk, daily review |
| P8 — Scale | Ongoing | Increase size via Kelly, add straddle strategy |

## Guiding Principles
1. **Loss is known at entry** — premium paid is the max loss. Always.
2. **Speed of move matters as much as direction** — theta kills slow moves.
3. **IV cheap + directional conviction must BOTH be true** — never one without the other.
4. **Cut losses, let winners run** — 50% hard stop, 100% partial exit, trailing stop remainder.
5. **No revenge trading** — mandatory 2-hour cooldown after any stop loss.
6. **Paper trade first** — minimum 4 weeks paper before 1 lot live.
