package strategy

import (
	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

type Selection struct {
	Strategy   core.StrategyType
	MaxCapital float64
	MinDTE     int
	MaxDTE     int
}

type StrategyConfig struct {
	PriceATM float64 // nearest atm strike price
}

var strategyMatrix = []struct {
	strategy   core.StrategyType
	ivrMin     float64
	ivrMax     float64
	confMin    float64
	dteMin     int
	dteMax     int
	maxCapPct  float64
}{
	{strategy: core.StrategyOTMCall, ivrMin: 0, ivrMax: 20, confMin: 0.80, dteMin: 7, dteMax: 14, maxCapPct: 0.03},
	{strategy: core.StrategyOTMPut, ivrMin: 0, ivrMax: 20, confMin: 0.80, dteMin: 7, dteMax: 14, maxCapPct: 0.03},
	{strategy: core.StrategyLongStraddle, ivrMin: 0, ivrMax: 25, confMin: 0, dteMin: 5, dteMax: 10, maxCapPct: 0.08},
	{strategy: core.StrategyLongCall, ivrMin: 0, ivrMax: 30, confMin: 0.70, dteMin: 7, dteMax: 14, maxCapPct: 0.10},
	{strategy: core.StrategyLongPut, ivrMin: 0, ivrMax: 30, confMin: 0.70, dteMin: 7, dteMax: 14, maxCapPct: 0.10},
	{strategy: core.StrategyBullCallSpread, ivrMin: 30, ivrMax: 50, confMin: 0.65, dteMin: 10, dteMax: 21, maxCapPct: 0.12},
	{strategy: core.StrategyBearPutSpread, ivrMin: 30, ivrMax: 50, confMin: 0.65, dteMin: 10, dteMax: 21, maxCapPct: 0.12},
}

func Select(sig core.Signal, dte int) (Selection, bool) {
	best := Selection{}
	found := false
	highestConf := 0.0

	for _, row := range strategyMatrix {
		if sig.IVRank < row.ivrMin || sig.IVRank >= row.ivrMax {
			continue
		}
		if dte < row.dteMin || dte > row.dteMax {
			continue
		}
		if sig.Conviction < row.confMin {
			continue
		}

		if row.strategy == core.StrategyLongCall && sig.Direction != core.DirectionBullish {
			continue
		}
		if row.strategy == core.StrategyLongPut && sig.Direction != core.DirectionBearish {
			continue
		}
		if row.strategy == core.StrategyBullCallSpread && sig.Direction != core.DirectionBullish {
			continue
		}
		if row.strategy == core.StrategyBearPutSpread && sig.Direction != core.DirectionBearish {
			continue
		}
		if row.strategy == core.StrategyOTMCall && sig.Direction != core.DirectionBullish {
			continue
		}
		if row.strategy == core.StrategyOTMPut && sig.Direction != core.DirectionBearish {
			continue
		}
		if row.strategy == core.StrategyLongStraddle && sig.Direction == core.DirectionNeutral {
			continue
		}

		if sig.Conviction > highestConf {
			highestConf = sig.Conviction
			best = Selection{
				Strategy:   row.strategy,
				MaxCapital: row.maxCapPct,
				MinDTE:     row.dteMin,
				MaxDTE:     row.dteMax,
			}
			found = true
		}
	}

	return best, found
}

func (s Selection) DirectionFromStrategy() core.Direction {
	switch s.Strategy {
	case core.StrategyLongCall, core.StrategyBullCallSpread, core.StrategyOTMCall:
		return core.DirectionBullish
	case core.StrategyLongPut, core.StrategyBearPutSpread, core.StrategyOTMPut:
		return core.DirectionBearish
	case core.StrategyLongStraddle:
		return core.DirectionNeutral
	default:
		return core.DirectionNeutral
	}
}

func (s Selection) IsSpread() bool {
	return s.Strategy == core.StrategyBullCallSpread || s.Strategy == core.StrategyBearPutSpread
}

func (s Selection) MaxLossPct() float64 {
	switch s.Strategy {
	case core.StrategyLongCall, core.StrategyLongPut, core.StrategyOTMCall, core.StrategyOTMPut:
		return 1.0
	case core.StrategyBullCallSpread, core.StrategyBearPutSpread:
		return 0.6
	case core.StrategyLongStraddle:
		return 1.0
	default:
		return 1.0
	}
}

func CalcSpreadWidth(strikeStep int, strategy core.StrategyType) int {
	switch strategy {
	case core.StrategyBullCallSpread, core.StrategyBearPutSpread:
		return 3 * strikeStep
	default:
		return 0
	}
}

func StrikeStep() int {
	return 50
}
