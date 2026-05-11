package execution

import (
	"sync"
	"time"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

type TrackedOrder struct {
	OrderID     string
	Leg         core.Leg
	Status      string
	PlacedAt    time.Time
	FilledAt    *time.Time
	FillPrice   float64
	Strategy    core.StrategyType
}

type OrderTracker struct {
	mu      sync.RWMutex
	orders  map[string]*TrackedOrder
}

func NewOrderTracker() *OrderTracker {
	return &OrderTracker{
		orders: make(map[string]*TrackedOrder),
	}
}

func (t *OrderTracker) Add(orderID string, leg core.Leg, strategy core.StrategyType) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.orders[orderID] = &TrackedOrder{
		OrderID:  orderID,
		Leg:      leg,
		Status:   "PENDING",
		PlacedAt: time.Now(),
		Strategy: strategy,
	}
}

func (t *OrderTracker) Update(orderID, status string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if o, ok := t.orders[orderID]; ok {
		o.Status = status
		if status == "TRADED" {
			now := time.Now()
			o.FilledAt = &now
			o.FillPrice = o.Leg.Price
		}
	}
}

func (t *OrderTracker) GetOpen() []TrackedOrder {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var open []TrackedOrder
	for _, o := range t.orders {
		if o.Status == "PENDING" || o.Status == "TRADED" {
			open = append(open, *o)
		}
	}
	return open
}

func (t *OrderTracker) GetByStrategy(strategy core.StrategyType) []TrackedOrder {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []TrackedOrder
	for _, o := range t.orders {
		if o.Strategy == strategy {
			result = append(result, *o)
		}
	}
	return result
}
