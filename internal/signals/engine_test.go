package signals

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

type mockStore struct {
	low52   float64
	high52  float64
	history []float64
	err     error
}

func (m *mockStore) GetVIXRange() (float64, float64, error) {
	return m.low52, m.high52, m.err
}

func (m *mockStore) GetRecentIVHistory(days int) ([]float64, error) {
	return m.history, m.err
}

func TestEngine_Process(t *testing.T) {
	store := &mockStore{
		low52:   12,
		high52:  30,
		history: []float64{14, 15, 16, 15, 14, 13, 14, 15, 16, 15, 14, 13, 14, 15, 16, 15, 14, 13, 14, 15},
	}
	engine := NewEngine(store)

	expiry := time.Now().AddDate(0, 0, 14)
	snap := core.MarketSnapshot{
		Timestamp: time.Now(),
		SpotPrice: 22500,
		IndiaVIX:  14.5,
		ATMStrike: 22500,
		Chain: []core.OptionData{
			{Strike: 22500, Expiry: expiry, OptionType: core.CE, OI: 100000, IV: 0.13, Bid: 180, Ask: 190},
			{Strike: 22500, Expiry: expiry, OptionType: core.PE, OI: 120000, IV: 0.14, Bid: 175, Ask: 185},
		},
		Expiries: []time.Time{expiry},
	}

	signal := engine.Process(snap)

	if signal.IVRank <= 0 || signal.IVRank >= 100 {
		t.Errorf("IVRank out of range: %v", signal.IVRank)
	}
	if signal.ExpectedMove <= 0 {
		t.Errorf("ExpectedMove should be positive: %v", signal.ExpectedMove)
	}
	if signal.Timestamp.IsZero() {
		t.Error("Timestamp is zero")
	}
}

func TestEngine_Process_EmptyStore(t *testing.T) {
	store := &mockStore{err: errors.New("no data")}
	engine := NewEngine(store)

	snap := core.MarketSnapshot{
		Timestamp: time.Now(),
		SpotPrice: 22500,
		IndiaVIX:  14.5,
		Expiries:  []time.Time{time.Now().AddDate(0, 0, 14)},
	}

	signal := engine.Process(snap)

	if signal.IVRank != 50 {
		t.Errorf("IVRank with no store = %v, want 50", signal.IVRank)
	}
	if signal.IVZScore != 0 {
		t.Errorf("IVZScore with no store = %v, want 0", signal.IVZScore)
	}
}

func TestEngine_Process_ZeroVIX(t *testing.T) {
	store := &mockStore{low52: 12, high52: 30}
	engine := NewEngine(store)

	snap := core.MarketSnapshot{
		SpotPrice: 22500,
		IndiaVIX:  0,
	}

	signal := engine.Process(snap)

	if signal.ExpectedMove != 0 {
		t.Errorf("ExpectedMove with zero VIX should be 0, got %v", signal.ExpectedMove)
	}
}

func TestClassifyRegime(t *testing.T) {
	tests := []struct {
		ivr      float64
		expected core.Regime
	}{
		{ivr: 20, expected: core.RegimeBuyIV},
		{ivr: 34, expected: core.RegimeBuyIV},
		{ivr: 35, expected: core.RegimeNeutral},
		{ivr: 45, expected: core.RegimeNeutral},
		{ivr: 55, expected: core.RegimeNeutral},
		{ivr: 56, expected: core.RegimeSellIV},
		{ivr: 80, expected: core.RegimeSellIV},
	}

	for _, tt := range tests {
		got := classifyRegime(tt.ivr)
		if got != tt.expected {
			t.Errorf("classifyRegime(%v) = %v, want %v", tt.ivr, got, tt.expected)
		}
	}
}

func TestCalcPCR(t *testing.T) {
	expiry := time.Now().AddDate(0, 0, 14)
	chain := []core.OptionData{
		{Strike: 22500, Expiry: expiry, OptionType: core.CE, OI: 100000},
		{Strike: 22500, Expiry: expiry, OptionType: core.PE, OI: 150000},
	}

	pcr := calcPCR(chain)
	if math.Abs(pcr-1.5) > 0.01 {
		t.Errorf("calcPCR() = %v, want 1.5", pcr)
	}
}

func TestCalcPCR_ZeroCEOI(t *testing.T) {
	pcr := calcPCR(nil)
	if pcr != 1.0 {
		t.Errorf("calcPCR(nil) = %v, want 1.0", pcr)
	}
}
