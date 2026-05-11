package strategy

import (
	"context"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

type Engine struct {
	lastSpot   float64
	lastDTE    int
	capital    float64
}

func NewStrategyEngine() *Engine {
	return &Engine{
		lastDTE: 10,
		capital: 500000,
	}
}

func (e *Engine) Run(ctx context.Context, in <-chan core.Signal, out chan<- core.TradeDecision) {

	for {
		select {
		case sig := <-in:
			if sig.SpotPrice > 0 {
				e.lastSpot = sig.SpotPrice
			}
			if sig.ExpectedMove > 0 {
				e.lastDTE = computeDTE(sig)
			}

			sel, ok := Select(sig, e.lastDTE)
			if !ok {
				continue
			}

			decision, ok := BuildTradeDecision(sig, e.lastSpot, sel, e.capital, nil)
			if !ok {
				continue
			}

			select {
			case out <- decision:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func computeDTE(sig core.Signal) int {
	dte := int(sig.ExpectedMove / 50)
	if dte < 5 {
		dte = 10
	}
	if dte > 21 {
		dte = 21
	}
	return dte
}


