package strategy

import (
	"math"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

func SelectStrikes(sig core.Signal, spot float64, sel Selection) []core.Leg {
	switch sel.Strategy {
	case core.StrategyLongCall:
		return longCallStrikes(sig, spot)
	case core.StrategyLongPut:
		return longPutStrikes(sig, spot)
	case core.StrategyBullCallSpread:
		return bullCallSpreadStrikes(sig, spot)
	case core.StrategyBearPutSpread:
		return bearPutSpreadStrikes(sig, spot)
	case core.StrategyLongStraddle:
		return longStraddleStrikes(sig, spot)
	case core.StrategyOTMCall:
		return otmCallStrikes(sig, spot)
	case core.StrategyOTMPut:
		return otmPutStrikes(sig, spot)
	default:
		return nil
	}
}

func longCallStrikes(sig core.Signal, spot float64) []core.Leg {
	step := float64(StrikeStep())
	buyStrike := nearestStrike(spot, step)
	return []core.Leg{
		{Strike: int(buyStrike), OptionType: core.CE, Action: "BUY"},
	}
}

func longPutStrikes(sig core.Signal, spot float64) []core.Leg {
	step := float64(StrikeStep())
	buyStrike := nearestStrike(spot, step)
	return []core.Leg{
		{Strike: int(buyStrike), OptionType: core.PE, Action: "BUY"},
	}
}

func bullCallSpreadStrikes(sig core.Signal, spot float64) []core.Leg {
	step := float64(StrikeStep())
	buyStrike := nearestStrike(spot, step)
	spread := float64(CalcSpreadWidth(StrikeStep(), core.StrategyBullCallSpread))
	sellStrike := buyStrike + spread

	return []core.Leg{
		{Strike: int(buyStrike), OptionType: core.CE, Action: "BUY"},
		{Strike: int(sellStrike), OptionType: core.CE, Action: "SELL"},
	}
}

func bearPutSpreadStrikes(sig core.Signal, spot float64) []core.Leg {
	step := float64(StrikeStep())
	buyStrike := nearestStrike(spot, step)
	spread := float64(CalcSpreadWidth(StrikeStep(), core.StrategyBearPutSpread))
	sellStrike := buyStrike - spread

	return []core.Leg{
		{Strike: int(buyStrike), OptionType: core.PE, Action: "BUY"},
		{Strike: int(sellStrike), OptionType: core.PE, Action: "SELL"},
	}
}

func longStraddleStrikes(sig core.Signal, spot float64) []core.Leg {
	step := float64(StrikeStep())
	strike := nearestStrike(spot, step)

	return []core.Leg{
		{Strike: int(strike), OptionType: core.CE, Action: "BUY"},
		{Strike: int(strike), OptionType: core.PE, Action: "BUY"},
	}
}

func otmCallStrikes(sig core.Signal, spot float64) []core.Leg {
	step := float64(StrikeStep())
	buyStrike := nearestStrike(spot, step) + step

	return []core.Leg{
		{Strike: int(buyStrike), OptionType: core.CE, Action: "BUY"},
	}
}

func otmPutStrikes(sig core.Signal, spot float64) []core.Leg {
	step := float64(StrikeStep())
	buyStrike := nearestStrike(spot, step) - step

	return []core.Leg{
		{Strike: int(buyStrike), OptionType: core.PE, Action: "BUY"},
	}
}

func nearestStrike(spot, step float64) float64 {
	return math.Round(spot/step) * step
}

func deltaBasedStrike(spot, targetDelta, step float64) float64 {
	strikes := []float64{-2, -1, 0, 1, 2}
	best := 0.0
	minDiff := 1.0

	for _, offset := range strikes {
		strike := nearestStrike(spot, step) + offset*step
		approxDelta := 0.5 - 0.1*((strike-spot)/step)
		if approxDelta < 0 {
			approxDelta = 0
		}
		if approxDelta > 1 {
			approxDelta = 1
		}
		diff := math.Abs(approxDelta - targetDelta)
		if diff < minDiff {
			minDiff = diff
			best = strike
		}
	}
	if best == 0 {
		return nearestStrike(spot, step)
	}
	return best
}
