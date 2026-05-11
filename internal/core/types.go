package core

import "time"

type Regime int

const (
	RegimeBuyIV  Regime = iota
	RegimeSellIV
	RegimeNeutral
)

type Direction int

const (
	DirectionBullish Direction = iota
	DirectionBearish
	DirectionNeutral
)

type StrategyType int

const (
	StrategyLongCall StrategyType = iota
	StrategyLongPut
	StrategyBullCallSpread
	StrategyBearPutSpread
	StrategyLongStraddle
	StrategyOTMCall
	StrategyOTMPut
)

type OptionType string

const (
	CE OptionType = "CE"
	PE OptionType = "PE"
)

type MarketSnapshot struct {
	Timestamp time.Time
	SpotPrice float64
	IndiaVIX  float64
	Chain     []OptionData
	ATMStrike int
	Expiries  []time.Time
}

type OptionData struct {
	Strike     int
	Expiry     time.Time
	OptionType OptionType
	LTP        float64
	IV         float64
	Delta      float64
	Gamma      float64
	Theta      float64
	Vega       float64
	OI         int64
	Volume     int64
	Bid        float64
	Ask        float64
	MidPrice   float64
}

type Signal struct {
	IVRank       float64
	IVZScore     float64
	ExpectedMove float64
	Regime       Regime
	Direction    Direction
	Conviction   float64
	MaxPain      int
	Skew         float64
	Timestamp    time.Time
	SpotPrice    float64
}

type TradeDecision struct {
	Strategy     StrategyType
	Legs         []Leg
	Lots         int
	MaxLoss      float64
	ProfitTarget float64
	StopLoss     float64
	ExpectedEV   float64
	Reason       string
}

type Leg struct {
	Strike       int
	Expiry       time.Time
	OptionType   OptionType
	Action       string
	Quantity     int
	Price        float64
	TradingSymbol string
	SecurityID   string
}

type Position struct {
	ID           string
	EntryValue   float64
	CurrentValue float64
	Lots         int
	EntryPrice   float64
	CurrentPrice float64
	CurrentDelta float64
	Expiry       time.Time
	HalfExited   bool
	Strategy     StrategyType
}

type RiskAction struct {
	Type       RiskActionType
	PositionID string
	Qty        int
	Reason     string
}

type RiskActionType int

const (
	RiskClose RiskActionType = iota
	RiskPartialClose
	RiskAlert
)

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Package   string
	Message   string
	Data      map[string]any
}

type Config struct {
	Dhan    DhanConfig    `yaml:"dhan"`
	Trading TradingConfig `yaml:"trading"`
	Risk    RiskConfig    `yaml:"risk"`
	Notify  NotifyConfig  `yaml:"notify"`
}

type DhanConfig struct {
	ClientID    string `yaml:"client_id"`
	AccessToken string `yaml:"access_token"`
	BaseURL     string `yaml:"base_url"`
}

type TradingConfig struct {
	MaxOpenTrades     int     `yaml:"max_open_trades"`
	MinDTE            int     `yaml:"min_dte"`
	MaxDTE            int     `yaml:"max_dte"`
	MinIVRank         float64 `yaml:"min_iv_rank"`
	MinConviction     float64 `yaml:"min_conviction"`
	ProfitTargetMult  float64 `yaml:"profit_target_multiplier"`
	StopLossMult      float64 `yaml:"stop_loss_multiplier"`
	MaxCapitalPerTrade float64 `yaml:"max_capital_per_trade_pct"`
}

type RiskConfig struct {
	MaxPortfolioRiskPct float64 `yaml:"max_portfolio_risk_pct"`
	MaxDailyLossPct     float64 `yaml:"max_daily_loss_pct"`
	MaxWeeklyLossPct    float64 `yaml:"max_weekly_loss_pct"`
	CooldownMinutes     int     `yaml:"cooldown_minutes"`
}

type NotifyConfig struct {
	TelegramBotToken string `yaml:"telegram_bot_token"`
	TelegramChatID   int64  `yaml:"telegram_chat_id"`
}

func MidPrice(bid, ask float64) float64 {
	if bid == 0 || ask == 0 {
		return 0
	}
	return (bid + ask) / 2
}

func DTE(expiry time.Time) int {
	return int(time.Until(expiry).Hours() / 24)
}
