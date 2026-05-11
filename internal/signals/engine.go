package signals

import (
	"context"
	"time"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

type Store interface {
	GetVIXRange() (low52, high52 float64, err error)
	GetRecentIVHistory(days int) ([]float64, error)
}

type Engine struct {
	store Store
}

func NewEngine(store Store) *Engine {
	return &Engine{store: store}
}

func (e *Engine) Run(ctx context.Context, in <-chan core.MarketSnapshot, out chan<- core.Signal) {
	for {
		select {
		case snap := <-in:
			signal := e.Process(snap)
			select {
			case out <- signal:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (e *Engine) Process(snap core.MarketSnapshot) core.Signal {
	signal := core.Signal{
		Timestamp: time.Now(),
		SpotPrice: snap.SpotPrice,
	}

	signal.IVRank = e.calcIVRank(snap.IndiaVIX)
	signal.IVZScore = e.calcIVZScore(snap.Chain)
	signal.ExpectedMove = calcExpectedMove(snap)
	signal.MaxPain = MaxPain(snap.Chain)
	signal.Skew = calcSkew(snap.Chain, snap.ATMStrike)
	signal.Regime = classifyRegime(signal.IVRank)

	conv := convictionFromSnapshot(snap)
	signal.Conviction = conv.Score
	signal.Direction = conv.Direction

	return signal
}

func (e *Engine) calcIVRank(currentVIX float64) float64 {
	if currentVIX <= 0 {
		return 50
	}
	low52, high52, err := e.store.GetVIXRange()
	if err != nil || high52-low52 <= 0 {
		return 50
	}
	return CalcIVRank(currentVIX, low52, high52)
}

func (e *Engine) calcIVZScore(chain []core.OptionData) float64 {
	var avgIV float64
	count := 0
	for _, opt := range chain {
		if opt.IV > 0 {
			avgIV += opt.IV
			count++
		}
	}
	if count == 0 {
		return 0
	}
	avgIV /= float64(count)

	history, err := e.store.GetRecentIVHistory(20)
	if err != nil || len(history) < 2 {
		return 0
	}

	return CalcIVZScore(avgIV, history)
}

func calcExpectedMove(snap core.MarketSnapshot) float64 {
	if snap.IndiaVIX <= 0 || len(snap.Expiries) == 0 {
		return 0
	}
	dte := core.DTE(snap.Expiries[0])
	if dte <= 0 {
		dte = 10
	}
	iv := snap.IndiaVIX / 100
	return ExpectedMove(snap.SpotPrice, iv, dte)
}

func classifyRegime(ivr float64) core.Regime {
	switch {
	case ivr < 35:
		return core.RegimeBuyIV
	case ivr > 55:
		return core.RegimeSellIV
	default:
		return core.RegimeNeutral
	}
}

type convictionResult struct {
	Score     float64
	Direction core.Direction
}

func convictionFromSnapshot(snap core.MarketSnapshot) convictionResult {
	d := DirectionalSignals{
		SpotPrice:   snap.SpotPrice,
		ATMStrike:   snap.ATMStrike,
		EMAPosition: extractEMASignal(snap.Chain),
		PCR:         calcPCR(snap.Chain),
	}

	score := ConvictionScore(d)
	dir := ResolveDirection(d)

	return convictionResult{Score: score, Direction: dir}
}

func extractEMASignal(chain []core.OptionData) int {
	return 0
}

func calcPCR(chain []core.OptionData) float64 {
	var ceOI, peOI int64
	for _, opt := range chain {
		if opt.OI == 0 {
			continue
		}
		if opt.OptionType == core.CE {
			ceOI += opt.OI
		} else {
			peOI += opt.OI
		}
	}
	if ceOI == 0 {
		return 1.0
	}
	return float64(peOI) / float64(ceOI)
}
