package strategy

import (
	"math"
	"testing"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

func TestKellySize_Positive(t *testing.T) {
	k := KellySize(0.45, 2.0, 0.5, 100000)
	if k <= 0 {
		t.Errorf("positive EV trade should have kelly > 0, got %v", k)
	}
	expected := ((0.45*2.0 - 0.55*0.5) / 2.0) / 2.0
	if math.Abs(k-expected) > 0.01 {
		t.Errorf("KellySize() = %v, want %v", k, expected)
	}
}

func TestKellySize_Negative(t *testing.T) {
	k := KellySize(0.30, 1.5, 0.8, 100000)
	if k != 0 {
		t.Errorf("negative EV trade should have kelly = 0, got %v", k)
	}
}

func TestKellySize_ZeroWinMult(t *testing.T) {
	k := KellySize(0.5, 0, 0.5, 100000)
	if k != 0 {
		t.Errorf("zero win mult should return 0, got %v", k)
	}
}

func TestPositionSize(t *testing.T) {
	sel := Selection{MaxCapital: 0.10}
	ps := PositionSize(100000, sel, 0.08)
	if ps > 10000 {
		t.Errorf("position size %v should be capped at 10%% = 10000", ps)
	}
}

func TestCalcLots(t *testing.T) {
	tests := []struct {
		name       string
		allocation float64
		premium    float64
		want       int
	}{
		{name: "exact", allocation: 30000, premium: 180, want: 2},
		{name: "partial", allocation: 25000, premium: 180, want: 1},
		{name: "not enough", allocation: 5000, premium: 180, want: 0},
		{name: "zero premium", allocation: 30000, premium: 0, want: 0},
		{name: "zero allocation", allocation: 0, premium: 180, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcLots(tt.allocation, tt.premium)
			if got != tt.want {
				t.Errorf("CalcLots(%v, %v) = %v, want %v", tt.allocation, tt.premium, got, tt.want)
			}
		})
	}
}

func TestCalcMaxLoss_LongCall(t *testing.T) {
	sel := Selection{Strategy: core.StrategyLongCall}
	legs := []core.Leg{
		{Strike: 22500, OptionType: core.CE, Action: "BUY", Price: 180},
	}
	loss := CalcMaxLoss(sel, legs, 2)
	expected := 180.0 * 2
	if loss != expected {
		t.Errorf("max loss = %v, want %v", loss, expected)
	}
}

func TestCalcMaxLoss_Spread(t *testing.T) {
	sel := Selection{Strategy: core.StrategyBullCallSpread}
	legs := []core.Leg{
		{Strike: 22500, OptionType: core.CE, Action: "BUY", Price: 185},
		{Strike: 22650, OptionType: core.CE, Action: "SELL", Price: 125},
	}
	loss := CalcMaxLoss(sel, legs, 1)
	expected := (185.0 - 125.0)
	if loss != expected {
		t.Errorf("spread max loss = %v, want %v", loss, expected)
	}
}

func TestTotalPremium(t *testing.T) {
	legs := []core.Leg{
		{Action: "BUY", Price: 180},
		{Action: "SELL", Price: 120},
	}
	tp := totalPremium(legs)
	if tp != 180 {
		t.Errorf("totalPremium = %v, want 180", tp)
	}
}

func TestCalcSpreadDebit(t *testing.T) {
	legs := []core.Leg{
		{Strike: 22500, Action: "BUY", Price: 185},
		{Strike: 22650, Action: "SELL", Price: 125},
	}
	debit := calcSpreadDebit(legs)
	expected := 185.0 - 125.0
	if debit != expected {
		t.Errorf("spread debit = %v, want %v", debit, expected)
	}
}

func TestCalcSpreadDebit_Negative(t *testing.T) {
	legs := []core.Leg{
		{Strike: 22500, Action: "SELL", Price: 185},
		{Strike: 22650, Action: "BUY", Price: 125},
	}
	debit := calcSpreadDebit(legs)
	if debit != 0 {
		t.Errorf("negative debit should return 0, got %v", debit)
	}
}
