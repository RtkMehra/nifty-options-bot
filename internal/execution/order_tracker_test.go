package execution

import (
	"testing"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

func TestNewOrderTracker(t *testing.T) {
	ot := NewOrderTracker()
	if ot == nil {
		t.Fatal("tracker is nil")
	}
}

func TestAddAndGetOpen(t *testing.T) {
	ot := NewOrderTracker()

	ot.Add("order-1", core.Leg{Strike: 22500, OptionType: core.CE, Action: "BUY", Price: 180}, core.StrategyLongCall)
	ot.Add("order-2", core.Leg{Strike: 22650, OptionType: core.CE, Action: "SELL", Price: 120}, core.StrategyBullCallSpread)

	open := ot.GetOpen()
	if len(open) != 2 {
		t.Fatalf("expected 2 open orders, got %d", len(open))
	}
}

func TestUpdateStatus(t *testing.T) {
	ot := NewOrderTracker()
	ot.Add("order-1", core.Leg{Price: 180}, core.StrategyLongCall)

	ot.Update("order-1", "TRADED")
	open := ot.GetOpen()
	if len(open) != 1 {
		t.Fatalf("expected 1 order after TRADED, got %d", len(open))
	}
	if open[0].Status != "TRADED" {
		t.Errorf("status = %v, want TRADED", open[0].Status)
	}
	if open[0].FillPrice != 180 {
		t.Errorf("fill price = %v, want 180", open[0].FillPrice)
	}
}

func TestGetByStrategy(t *testing.T) {
	ot := NewOrderTracker()
	ot.Add("o1", core.Leg{}, core.StrategyLongCall)
	ot.Add("o2", core.Leg{}, core.StrategyLongPut)
	ot.Add("o3", core.Leg{}, core.StrategyLongCall)

	callOrders := ot.GetByStrategy(core.StrategyLongCall)
	if len(callOrders) != 2 {
		t.Errorf("expected 2 LongCall orders, got %d", len(callOrders))
	}

	putOrders := ot.GetByStrategy(core.StrategyLongPut)
	if len(putOrders) != 1 {
		t.Errorf("expected 1 LongPut order, got %d", len(putOrders))
	}
}

func TestGetOpenEmpty(t *testing.T) {
	ot := NewOrderTracker()
	open := ot.GetOpen()
	if len(open) != 0 {
		t.Errorf("expected 0 open orders, got %d", len(open))
	}
}
