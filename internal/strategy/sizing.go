package strategy

import (
	"math"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

const LotSize = 75

func KellySize(winRate, winMult, lossMult, capital float64) float64 {
	if winMult <= 0 {
		return 0
	}
	k := (winRate*winMult - (1-winRate)*lossMult) / winMult
	if k < 0 {
		return 0
	}
	return k / 2.0
}

func PositionSize(capital float64, sel Selection, kellyFraction float64) float64 {
	maxCap := capital * sel.MaxCapital
	kellyCap := capital * kellyFraction

	return math.Min(maxCap, kellyCap)
}

func CalcLots(allocation, premium float64) int {
	if premium <= 0 || allocation <= 0 {
		return 0
	}
	costPerLot := premium * LotSize
	lots := int(allocation / costPerLot)
	if lots < 1 {
		return 0
	}
	return lots
}

func CalcMaxLoss(sel Selection, legs []core.Leg, lots int) float64 {
	switch sel.Strategy {
	case core.StrategyLongCall, core.StrategyLongPut, core.StrategyOTMCall, core.StrategyOTMPut:
		return totalPremium(legs) * float64(lots)
	case core.StrategyBullCallSpread:
		return calcSpreadDebit(legs) * float64(lots)
	case core.StrategyBearPutSpread:
		return calcSpreadDebit(legs) * float64(lots)
	case core.StrategyLongStraddle:
		return totalPremium(legs) * float64(lots)
	default:
		return totalPremium(legs) * float64(lots)
	}
}

func MaxLossPerLot(sel Selection, legs []core.Leg) float64 {
	switch sel.Strategy {
	case core.StrategyBullCallSpread:
		return calcSpreadWidthAmount(legs)
	case core.StrategyBearPutSpread:
		return calcSpreadWidthAmount(legs)
	default:
		return totalPremium(legs)
	}
}

func totalPremium(legs []core.Leg) float64 {
	sum := 0.0
	for _, l := range legs {
		if l.Action == "BUY" {
			sum += l.Price
		}
	}
	return sum
}

func calcSpreadDebit(legs []core.Leg) float64 {
	debit := 0.0
	for _, l := range legs {
		if l.Action == "BUY" {
			debit += l.Price
		} else {
			debit -= l.Price
		}
	}
	if debit < 0 {
		return 0
	}
	return debit
}

func calcSpreadWidthAmount(legs []core.Leg) float64 {
	if len(legs) < 2 {
		return 0
	}
	buy := legs[0]
	sell := legs[1]
	diff := buy.Strike - sell.Strike
	if diff < 0 {
		diff = -diff
	}
	return float64(diff)
}
