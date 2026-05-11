package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

func TestNewStrategyEngine(t *testing.T) {
	e := NewStrategyEngine()
	if e == nil {
		t.Fatal("engine is nil")
	}
}

func TestEngineRunContextCancel(t *testing.T) {
	e := NewStrategyEngine()
	in := make(chan core.Signal, 1)
	out := make(chan core.TradeDecision, 1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		e.Run(ctx, in, out)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not exit after context cancellation")
	}
}

func TestEngineRunProducesDecision(t *testing.T) {
	e := NewStrategyEngine()
	in := make(chan core.Signal, 1)
	out := make(chan core.TradeDecision, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go e.Run(ctx, in, out)

	in <- core.Signal{
		IVRank:       20,
		Conviction:   0.80,
		Direction:    core.DirectionBullish,
		SpotPrice:    22500,
		ExpectedMove: 400,
	}

	select {
	case decision := <-out:
		if decision.Lots <= 0 {
			t.Errorf("expected positive lots, got %d", decision.Lots)
		}
		if decision.MaxLoss <= 0 {
			t.Errorf("expected positive max loss, got %.0f", decision.MaxLoss)
		}
		t.Logf("decision: strategy=%d lots=%d maxLoss=%.0f ev=%.2f",
			decision.Strategy, decision.Lots, decision.MaxLoss, decision.ExpectedEV)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for trade decision")
	}
}

func TestEngineRunSkipsNoMatch(t *testing.T) {
	e := NewStrategyEngine()
	in := make(chan core.Signal, 1)
	out := make(chan core.TradeDecision, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go e.Run(ctx, in, out)

	in <- core.Signal{
		IVRank:       60,
		Conviction:   0.80,
		Direction:    core.DirectionBullish,
		SpotPrice:    22500,
		ExpectedMove: 400,
	}

	select {
	case <-out:
		t.Error("expected no decision for high IVR")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestComputeDTE(t *testing.T) {
	tests := []struct {
		name     string
		em       float64
		expected int
	}{
		{name: "normal", em: 400, expected: 8},
		{name: "low", em: 100, expected: 10},
		{name: "high", em: 1500, expected: 21},
		{name: "zero", em: 0, expected: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := core.Signal{ExpectedMove: tt.em}
			got := computeDTE(sig)
			if got != tt.expected {
				t.Errorf("computeDTE(%v) = %v, want %v", tt.em, got, tt.expected)
			}
		})
	}
}
