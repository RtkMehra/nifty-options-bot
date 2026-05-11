package execution

import (
	"context"
	"fmt"
	"time"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

type FillResult struct {
	Filled    bool
	OrderID   string
	FillPrice float64
}

type ExecutionEngine struct {
	dhan   *DhanClient
	dryRun bool
}

func NewExecutionEngine(dhan *DhanClient, dryRun bool) *ExecutionEngine {
	return &ExecutionEngine{
		dhan:   dhan,
		dryRun: dryRun,
	}
}

func (e *ExecutionEngine) Run(ctx context.Context, in <-chan core.TradeDecision, out chan<- core.RiskAction) {
	for {
		select {
		case decision := <-in:
			result := e.execute(decision)
			if !result.Filled {
				out <- core.RiskAction{
					Type:   core.RiskAlert,
					Reason: fmt.Sprintf("order failed: %s", decision.Reason),
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (e *ExecutionEngine) execute(decision core.TradeDecision) FillResult {
	if e.dryRun {
		return FillResult{
			Filled:    true,
			OrderID:   "dry-run",
			FillPrice: decision.Legs[0].Price,
		}
	}

	var lastFill FillResult
	for _, leg := range decision.Legs {
		result := e.smartFill(leg)
		if !result.Filled {
			return FillResult{Filled: false}
		}
		lastFill = result
	}

	return lastFill
}

func (e *ExecutionEngine) smartFill(leg core.Leg) FillResult {
	attempt := 0
	price := leg.Price

	for attempt < 3 {
		orderReq := OrderRequest{
			TransactionType: leg.Action,
			ExchangeSegment: "NSE_FNO",
			ProductType:     "INTRADAY",
			OrderType:       "LIMIT",
			Validity:        "DAY",
			TradingSymbol:   leg.TradingSymbol,
			SecurityID:      leg.SecurityID,
			Quantity:        leg.Quantity,
			Price:           price,
		}

		orderResp, err := e.dhan.PlaceOrder(orderReq)
		if err != nil {
			return FillResult{Filled: false}
		}

		status, err := e.waitForFill(orderResp.OrderID)
		if err == nil && status == "TRADED" {
			return FillResult{
				Filled:    true,
				OrderID:   orderResp.OrderID,
				FillPrice: price,
			}
		}

		if err := e.dhan.CancelOrder(orderResp.OrderID); err != nil {
			return FillResult{Filled: false}
		}

		price += 0.5
		attempt++
	}

	return FillResult{Filled: false}
}

func (e *ExecutionEngine) waitForFill(orderID string) (string, error) {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		status, err := e.dhan.GetOrderStatus(orderID)
		if err != nil {
			return "", err
		}
		if status == "TRADED" || status == "REJECTED" || status == "CANCELLED" {
			return status, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return "", fmt.Errorf("timeout waiting for fill")
}
