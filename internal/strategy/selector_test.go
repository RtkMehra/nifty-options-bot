package strategy

import (
	"testing"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

func TestSelect_LongCall(t *testing.T) {
	sig := core.Signal{IVRank: 20, Conviction: 0.80, Direction: core.DirectionBullish}
	sel, ok := Select(sig, 12)
	if !ok {
		t.Fatal("expected selection for LongCall")
	}
	if sel.Strategy != core.StrategyLongCall {
		t.Errorf("expected LongCall, got %v", sel.Strategy)
	}
}

func TestSelect_LongPut(t *testing.T) {
	sig := core.Signal{IVRank: 20, Conviction: 0.80, Direction: core.DirectionBearish}
	sel, ok := Select(sig, 12)
	if !ok {
		t.Fatal("expected selection for LongPut")
	}
	if sel.Strategy != core.StrategyLongPut {
		t.Errorf("expected LongPut, got %v", sel.Strategy)
	}
}

func TestSelect_BullCallSpread(t *testing.T) {
	sig := core.Signal{IVRank: 40, Conviction: 0.70, Direction: core.DirectionBullish}
	sel, ok := Select(sig, 15)
	if !ok {
		t.Fatal("expected selection for BullCallSpread")
	}
	if sel.Strategy != core.StrategyBullCallSpread {
		t.Errorf("expected BullCallSpread, got %v", sel.Strategy)
	}
}

func TestSelect_BearPutSpread(t *testing.T) {
	sig := core.Signal{IVRank: 40, Conviction: 0.70, Direction: core.DirectionBearish}
	sel, ok := Select(sig, 15)
	if !ok {
		t.Fatal("expected selection for BearPutSpread")
	}
	if sel.Strategy != core.StrategyBearPutSpread {
		t.Errorf("expected BearPutSpread, got %v", sel.Strategy)
	}
}

func TestSelect_Straddle(t *testing.T) {
	sig := core.Signal{IVRank: 15, Conviction: 0.60, Direction: core.DirectionBullish}
	sel, ok := Select(sig, 7)
	if !ok {
		t.Fatal("expected selection for LongStraddle")
	}
	if sel.Strategy != core.StrategyLongStraddle {
		t.Errorf("expected LongStraddle, got %v", sel.Strategy)
	}
}

func TestSelect_OTMCall(t *testing.T) {
	sig := core.Signal{IVRank: 10, Conviction: 0.82, Direction: core.DirectionBullish}
	sel, ok := Select(sig, 10)
	if !ok {
		t.Fatal("expected selection for OTMCall")
	}
	if sel.Strategy != core.StrategyOTMCall {
		t.Errorf("expected OTMCall, got %v", sel.Strategy)
	}
}

func TestSelect_NoMatchHighIVR(t *testing.T) {
	sig := core.Signal{IVRank: 60, Conviction: 0.80, Direction: core.DirectionBullish}
	_, ok := Select(sig, 10)
	if ok {
		t.Error("expected no selection for high IVR")
	}
}

func TestSelect_NoMatchLowConviction(t *testing.T) {
	sig := core.Signal{IVRank: 20, Conviction: 0.40, Direction: core.DirectionBullish}
	_, ok := Select(sig, 3)
	if ok {
		t.Error("expected no selection for low conviction + bad DTE")
	}
}

func TestSelect_NoMatchBadDTE(t *testing.T) {
	sig := core.Signal{IVRank: 20, Conviction: 0.80, Direction: core.DirectionBullish}
	_, ok := Select(sig, 25)
	if ok {
		t.Error("expected no selection for DTE outside range")
	}
}

func TestSelect_NoMatchWrongDirection(t *testing.T) {
	sig := core.Signal{IVRank: 20, Conviction: 0.80, Direction: core.DirectionNeutral}
	_, ok := Select(sig, 10)
	if ok {
		t.Error("expected no selection for neutral direction")
	}
}

func TestSelect_PicksHighestConviction(t *testing.T) {
	sig := core.Signal{IVRank: 10, Conviction: 0.82, Direction: core.DirectionBullish}
	sel, ok := Select(sig, 10)
	if !ok {
		t.Fatal("expected a selection")
	}
	if sel.Strategy != core.StrategyOTMCall {
		t.Errorf("expected OTMCall (highest conviction match), got %v", sel.Strategy)
	}
}

func TestDirectionFromStrategy(t *testing.T) {
	tests := []struct {
		strategy core.StrategyType
		expected core.Direction
	}{
		{core.StrategyLongCall, core.DirectionBullish},
		{core.StrategyLongPut, core.DirectionBearish},
		{core.StrategyBullCallSpread, core.DirectionBullish},
		{core.StrategyBearPutSpread, core.DirectionBearish},
		{core.StrategyLongStraddle, core.DirectionNeutral},
		{core.StrategyOTMCall, core.DirectionBullish},
		{core.StrategyOTMPut, core.DirectionBearish},
	}

	for _, tt := range tests {
		sel := Selection{Strategy: tt.strategy}
		got := sel.DirectionFromStrategy()
		if got != tt.expected {
			t.Errorf("%v direction = %v, want %v", tt.strategy, got, tt.expected)
		}
	}
}

func TestIsSpread(t *testing.T) {
	sel1 := Selection{Strategy: core.StrategyBullCallSpread}
	if !sel1.IsSpread() {
		t.Error("BullCallSpread should be a spread")
	}
	sel2 := Selection{Strategy: core.StrategyLongCall}
	if sel2.IsSpread() {
		t.Error("LongCall should not be a spread")
	}
}

func TestCalcSpreadWidth(t *testing.T) {
	width := CalcSpreadWidth(50, core.StrategyBullCallSpread)
	if width != 150 {
		t.Errorf("spread width = %v, want 150", width)
	}

	width = CalcSpreadWidth(50, core.StrategyLongCall)
	if width != 0 {
		t.Errorf("non-spread width = %v, want 0", width)
	}
}

func TestSelectStrikes_LongCall(t *testing.T) {
	sig := core.Signal{}
	sel := Selection{Strategy: core.StrategyLongCall}
	legs := SelectStrikes(sig, 22500, sel)

	if len(legs) != 1 {
		t.Fatalf("expected 1 leg, got %d", len(legs))
	}
	if legs[0].OptionType != core.CE {
		t.Errorf("expected CE, got %v", legs[0].OptionType)
	}
	if legs[0].Action != "BUY" {
		t.Errorf("expected BUY, got %v", legs[0].Action)
	}
}

func TestSelectStrikes_LongPut(t *testing.T) {
	sig := core.Signal{}
	sel := Selection{Strategy: core.StrategyLongPut}
	legs := SelectStrikes(sig, 22500, sel)

	if len(legs) != 1 {
		t.Fatalf("expected 1 leg, got %d", len(legs))
	}
	if legs[0].OptionType != core.PE {
		t.Errorf("expected PE, got %v", legs[0].OptionType)
	}
}

func TestSelectStrikes_BullCallSpread(t *testing.T) {
	sig := core.Signal{}
	sel := Selection{Strategy: core.StrategyBullCallSpread}
	legs := SelectStrikes(sig, 22500, sel)

	if len(legs) != 2 {
		t.Fatalf("expected 2 legs, got %d", len(legs))
	}
	if legs[0].Action != "BUY" || legs[1].Action != "SELL" {
		t.Errorf("expected BUY then SELL, got %v %v", legs[0].Action, legs[1].Action)
	}
	if legs[1].Strike <= legs[0].Strike {
		t.Errorf("sell strike %v should be above buy strike %v", legs[1].Strike, legs[0].Strike)
	}
}

func TestSelectStrikes_BearPutSpread(t *testing.T) {
	sig := core.Signal{}
	sel := Selection{Strategy: core.StrategyBearPutSpread}
	legs := SelectStrikes(sig, 22500, sel)

	if len(legs) != 2 {
		t.Fatalf("expected 2 legs, got %d", len(legs))
	}
	if legs[1].Strike >= legs[0].Strike {
		t.Errorf("sell strike %v should be below buy strike %v", legs[1].Strike, legs[0].Strike)
	}
}

func TestSelectStrikes_Straddle(t *testing.T) {
	sig := core.Signal{}
	sel := Selection{Strategy: core.StrategyLongStraddle}
	legs := SelectStrikes(sig, 22500, sel)

	if len(legs) != 2 {
		t.Fatalf("expected 2 legs, got %d", len(legs))
	}
	if legs[0].Strike != legs[1].Strike {
		t.Errorf("straddle strikes should match: %v vs %v", legs[0].Strike, legs[1].Strike)
	}
}

func TestSelectStrikes_OTMCall(t *testing.T) {
	sig := core.Signal{}
	sel := Selection{Strategy: core.StrategyOTMCall}
	legs := SelectStrikes(sig, 22500, sel)

	if len(legs) != 1 {
		t.Fatalf("expected 1 leg, got %d", len(legs))
	}
	if legs[0].Strike <= 22500 {
		t.Errorf("OTM call strike %v should be above spot 22500", legs[0].Strike)
	}
}

func TestSelectStrikes_OTMPut(t *testing.T) {
	sig := core.Signal{}
	sel := Selection{Strategy: core.StrategyOTMPut}
	legs := SelectStrikes(sig, 22500, sel)

	if len(legs) != 1 {
		t.Fatalf("expected 1 leg, got %d", len(legs))
	}
	if legs[0].Strike >= 22500 {
		t.Errorf("OTM put strike %v should be below spot 22500", legs[0].Strike)
	}
}

func TestNearestStrike(t *testing.T) {
	tests := []struct {
		spot float64
		want float64
	}{
		{22499, 22500},
		{22500, 22500},
		{22501, 22500},
		{22524, 22500},
		{22549, 22550},
		{22751, 22750},
	}

	for _, tt := range tests {
		got := nearestStrike(tt.spot, 50)
		if got != tt.want {
			t.Errorf("nearestStrike(%v) = %v, want %v", tt.spot, got, tt.want)
		}
	}
}
