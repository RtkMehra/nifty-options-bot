package execution

import (
	"context"
	"testing"
	"time"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

func TestDryRun(t *testing.T) {
	e := NewExecutionEngine(nil, true)
	result := e.execute(core.TradeDecision{
		Strategy: core.StrategyLongCall,
		Legs:     []core.Leg{{Price: 180, Action: "BUY"}},
	})
	if !result.Filled {
		t.Error("dry run should always fill")
	}
	if result.FillPrice != 180 {
		t.Errorf("fill price = %v, want 180", result.FillPrice)
	}
}

func TestRunDryMode(t *testing.T) {
	e := NewExecutionEngine(nil, true)
	in := make(chan core.TradeDecision, 1)
	out := make(chan core.RiskAction, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go e.Run(ctx, in, out)

	in <- core.TradeDecision{
		Strategy: core.StrategyLongCall,
		Legs:     []core.Leg{{Price: 180, Action: "BUY"}},
	}

	select {
	case action := <-out:
		if action.Type != core.RiskAlert {
			t.Errorf("expected risk alert, got %v", action.Type)
		}
	case <-time.After(time.Second):
		t.Log("no risk action generated (expected)")
	}
}

func TestRunContextCancel(t *testing.T) {
	e := NewExecutionEngine(nil, true)
	in := make(chan core.TradeDecision, 1)
	out := make(chan core.RiskAction, 1)

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

func TestNewExecutionEngine(t *testing.T) {
	e := NewExecutionEngine(nil, true)
	if e == nil {
		t.Fatal("engine is nil")
	}
	if !e.dryRun {
		t.Error("expected dryRun = true")
	}
}
