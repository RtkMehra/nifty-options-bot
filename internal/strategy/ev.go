package strategy

import (
	"math"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

type EVResult struct {
	EV          float64
	WinRate     float64
	AvgWinMult  float64
	AvgLossMult float64
	IsPositive  bool
}

func CalcEV(sig core.Signal, sel Selection) EVResult {
	var winRate, avgWinMult, avgLossMult float64

	switch sel.Strategy {
	case core.StrategyLongCall, core.StrategyLongPut:
		winRate = 0.45
		avgWinMult = 2.0
		avgLossMult = 0.5
	case core.StrategyBullCallSpread, core.StrategyBearPutSpread:
		winRate = 0.50
		avgWinMult = 1.8
		avgLossMult = 0.4
	case core.StrategyLongStraddle:
		winRate = 0.35
		avgWinMult = 2.5
		avgLossMult = 0.6
	case core.StrategyOTMCall, core.StrategyOTMPut:
		winRate = 0.35
		avgWinMult = 4.0
		avgLossMult = 0.6
	}

	winRate = adjustWinRateForIVR(winRate, sig.IVRank)
	avgWinMult = adjustWinMultForConviction(avgWinMult, sig.Conviction)

	ev := (winRate*avgWinMult - (1-winRate)*avgLossMult)

	return EVResult{
		EV:          ev,
		WinRate:     winRate,
		AvgWinMult:  avgWinMult,
		AvgLossMult: avgLossMult,
		IsPositive:  ev > 0,
	}
}

func adjustWinRateForIVR(base float64, ivr float64) float64 {
	if ivr < 20 {
		return math.Min(base*1.2, 0.65)
	}
	if ivr < 30 {
		return math.Min(base*1.1, 0.60)
	}
	if ivr > 40 {
		return base * 0.9
	}
	return base
}

func adjustWinMultForConviction(base float64, conviction float64) float64 {
	if conviction > 0.85 {
		return base * 1.3
	}
	if conviction > 0.75 {
		return base * 1.1
	}
	return base
}

func BuildTradeDecision(sig core.Signal, spot float64, sel Selection, capital float64, chain []core.OptionData) (core.TradeDecision, bool) {
	ev := CalcEV(sig, sel)
	if !ev.IsPositive {
		return core.TradeDecision{}, false
	}

	legs := SelectStrikes(sig, spot, sel)
	if len(legs) == 0 {
		return core.TradeDecision{}, false
	}

	legs = enrichLegPrices(legs, chain)

	premium := totalPremium(legs)
	positionAmt := PositionSize(capital, sel, ev.EV)
	lotSize := CalcLots(positionAmt, premium)

	if lotSize == 0 {
		return core.TradeDecision{}, false
	}

	maxLoss := CalcMaxLoss(sel, legs, lotSize)
	profitTarget := maxLoss * 2.0
	stopLoss := maxLoss * 0.5

	return core.TradeDecision{
		Strategy:     sel.Strategy,
		Legs:         legs,
		Lots:         lotSize,
		MaxLoss:      maxLoss,
		ProfitTarget: profitTarget,
		StopLoss:     stopLoss,
		ExpectedEV:   ev.EV,
		Reason:       buildReason(sig, sel, ev),
	}, true
}

func enrichLegPrices(legs []core.Leg, chain []core.OptionData) []core.Leg {
	if len(chain) == 0 {
		for i := range legs {
			if legs[i].Price == 0 {
				legs[i].Price = 100
			}
		}
		return legs
	}
	for i, leg := range legs {
		for _, opt := range chain {
			if opt.Strike == leg.Strike && opt.OptionType == leg.OptionType {
				legs[i].Price = opt.MidPrice
				legs[i].TradingSymbol = buildSymbol(leg)
				break
			}
		}
	}
	return legs
}

func buildSymbol(leg core.Leg) string {
	return ""
}

func buildReason(sig core.Signal, sel Selection, ev EVResult) string {
	return ""
}
