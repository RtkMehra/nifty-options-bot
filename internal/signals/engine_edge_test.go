package signals

import (
	"context"
	"testing"
	"time"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

func TestEngineRunContextCancel(t *testing.T) {
	store := &mockStore{low52: 12, high52: 30}
	engine := NewEngine(store)

	in := make(chan core.MarketSnapshot, 1)
	out := make(chan core.Signal, 1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		engine.Run(ctx, in, out)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not exit after context cancellation")
	}
}

func TestEngineRunProcessesSnapshot(t *testing.T) {
	store := &mockStore{
		low52:   12,
		high52:  30,
		history: []float64{14, 15, 16},
	}
	engine := NewEngine(store)

	in := make(chan core.MarketSnapshot, 1)
	out := make(chan core.Signal, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go engine.Run(ctx, in, out)

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

	in <- snap

	select {
	case sig := <-out:
		if sig.Timestamp.IsZero() {
			t.Error("signal has zero timestamp")
		}
		if sig.IVRank <= 0 {
			t.Errorf("expected positive IVRank, got %v", sig.IVRank)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for signal output")
	}
}

func TestEngineRunBlocksOnFullOut(t *testing.T) {
	store := &mockStore{low52: 12, high52: 30}
	engine := NewEngine(store)

	in := make(chan core.MarketSnapshot, 1)
	out := make(chan core.Signal, 0)

	ctx, cancel := context.WithCancel(context.Background())

	snap := core.MarketSnapshot{
		Timestamp: time.Now(),
		SpotPrice: 22500,
		IndiaVIX:  14.5,
		Expiries:  []time.Time{time.Now().AddDate(0, 0, 14)},
	}

	done := make(chan struct{})
	go func() {
		engine.Run(ctx, in, out)
		close(done)
	}()

	in <- snap
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not exit after cancel with blocked out channel")
	}
}

func TestExtractEMASignal(t *testing.T) {
	got := extractEMASignal(nil)
	if got != 0 {
		t.Errorf("extractEMASignal() = %v, want 0 (stub)", got)
	}
}

func TestCalcExpectedMoveEmptyExpiries(t *testing.T) {
	snap := core.MarketSnapshot{
		SpotPrice: 22500,
		IndiaVIX:  14.5,
	}
	em := calcExpectedMove(snap)
	if em != 0 {
		t.Errorf("expected move with no expiries = %v, want 0", em)
	}
}

func TestCalcExpectedMoveDefaultDTE(t *testing.T) {
	snap := core.MarketSnapshot{
		SpotPrice: 22500,
		IndiaVIX:  14.5,
		Expiries:  []time.Time{time.Now().Add(-time.Hour)},
	}
	em := calcExpectedMove(snap)
	if em <= 0 {
		t.Errorf("expected move with past expiry should use default DTE, got %v", em)
	}
}

func TestConvictionUnusedFields(t *testing.T) {
	d := DirectionalSignals{
		BOSConfirmed: true,
		OBRetest:     true,
		SpotPrice:    99999,
		ATMStrike:    99999,
	}

	score := ConvictionScore(d)
	if score != 0.55 {
		t.Errorf("BOS+OB score with unused fields = %v, want 0.55", score)
	}
}

func TestIVZScoreNoStddev(t *testing.T) {
	got := CalcIVZScore(15, []float64{15, 15, 15, 15})
	if got != 0 {
		t.Errorf("zero stddev should return 0, got %v", got)
	}
}
