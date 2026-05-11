package strategy

import (
	"math"
	"testing"
	"time"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

func TestCalcEV_Positive(t *testing.T) {
	sig := core.Signal{IVRank: 20, Conviction: 0.75}
	sel := Selection{Strategy: core.StrategyLongCall}

	ev := CalcEV(sig, sel)
	if !ev.IsPositive {
		t.Error("expected positive EV for LongCall with low IVR")
	}
	if ev.EV <= 0 {
		t.Errorf("EV should be > 0, got %v", ev.EV)
	}
}

func TestCalcEV_DebitSpread(t *testing.T) {
	sig := core.Signal{IVRank: 40, Conviction: 0.70}
	sel := Selection{Strategy: core.StrategyBullCallSpread}

	ev := CalcEV(sig, sel)
	if !ev.IsPositive {
		t.Error("expected positive EV for debit spread")
	}
}

func TestCalcEV_Straddle(t *testing.T) {
	sig := core.Signal{IVRank: 15, Conviction: 0.70}
	sel := Selection{Strategy: core.StrategyLongStraddle}

	ev := CalcEV(sig, sel)
	if !ev.IsPositive {
		t.Error("expected positive EV for straddle with low IVR")
	}
}

func TestCalcEV_OTM(t *testing.T) {
	sig := core.Signal{IVRank: 10, Conviction: 0.85}
	sel := Selection{Strategy: core.StrategyOTMCall}

	ev := CalcEV(sig, sel)
	if !ev.IsPositive {
		t.Error("expected positive EV for OTM with high conviction")
	}
}

func TestCalcEV_HighIVR(t *testing.T) {
	sig := core.Signal{IVRank: 45, Conviction: 0.75}
	sel := Selection{Strategy: core.StrategyLongCall}

	ev := CalcEV(sig, sel)
	expectedWR := 0.45 * 0.9
	expectedEV := expectedWR*2.0 - (1-expectedWR)*0.5
	if math.Abs(ev.EV-expectedEV) > 0.01 {
		t.Errorf("EV for high IVR = %v, want %v", ev.EV, expectedEV)
	}
}

func TestCalcEV_HighConviction(t *testing.T) {
	sig := core.Signal{IVRank: 20, Conviction: 0.90}
	sel := Selection{Strategy: core.StrategyLongCall}

	ev := CalcEV(sig, sel)
	if ev.EV <= 0 {
		t.Errorf("EV should be positive for high conviction, got %v", ev.EV)
	}
	if ev.AvgWinMult <= 2.0 {
		t.Errorf("AvgWinMult should be boosted above base 2.0, got %v", ev.AvgWinMult)
	}
}

func TestBuildTradeDecision(t *testing.T) {
	sig := core.Signal{IVRank: 20, Conviction: 0.80, Direction: core.DirectionBullish, ExpectedMove: 400}
	sel, ok := Select(sig, 10)
	if !ok {
		t.Fatal("Select() failed")
	}

	chain := []core.OptionData{
		{Strike: 22500, Expiry: expiry, OptionType: core.CE, IV: 0.13, Bid: 180, Ask: 190},
		{Strike: 22500, Expiry: expiry, OptionType: core.PE, IV: 0.14, Bid: 175, Ask: 185},
	}

	decision, ok := BuildTradeDecision(sig, 22500, sel, 100000, chain)
	if ok {
		t.Logf("decision: strategy=%v lots=%d maxLoss=%.0f ev=%.2f", decision.Strategy, decision.Lots, decision.MaxLoss, decision.ExpectedEV)
	} else {
		t.Log("no trade decision (expected when premiums from chain are unavailable)")
	}
}

func TestBuildTradeDecision_NegativeEV(t *testing.T) {
	sig := core.Signal{IVRank: 50, Conviction: 0.40, Direction: core.DirectionBullish}
	sel := Selection{Strategy: core.StrategyLongCall}

	_, ok := BuildTradeDecision(sig, 22500, sel, 100000, nil)
	if ok {
		t.Error("expected no decision for negative EV")
	}
}

func TestAdjustWinRateForIVR(t *testing.T) {
	tests := []struct {
		base float64
		ivr  float64
	}{
		{0.45, 15},
		{0.45, 25},
		{0.45, 35},
		{0.45, 45},
	}

	for _, tt := range tests {
		got := adjustWinRateForIVR(tt.base, tt.ivr)
		if got <= 0 || got > 0.65 {
			t.Errorf("adjustWinRateForIVR(%v, %v) = %v out of range", tt.base, tt.ivr, got)
		}
	}
}

func TestAdjustWinMultForConviction(t *testing.T) {
	tests := []struct {
		base       float64
		conviction float64
	}{
		{2.0, 0.90},
		{2.0, 0.80},
		{2.0, 0.60},
	}

	for _, tt := range tests {
		got := adjustWinMultForConviction(tt.base, tt.conviction)
		if got < tt.base {
			t.Errorf("adjustWinMultForConviction(%v, %v) = %v < base", tt.base, tt.conviction, got)
		}
	}
}

var expiry = time.Now().AddDate(0, 0, 14)
