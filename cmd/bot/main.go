package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
	"github.com/ritikmehra/nifty-options-bot/internal/datafeed"
	"github.com/ritikmehra/nifty-options-bot/internal/execution"
	"github.com/ritikmehra/nifty-options-bot/internal/signals"
	"github.com/ritikmehra/nifty-options-bot/internal/store"
	"github.com/ritikmehra/nifty-options-bot/internal/strategy"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := store.NewDB("nifty_bot.db")
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	snapshots := make(chan core.MarketSnapshot, 10)
	signalsCh := make(chan core.Signal, 5)
	decisions := make(chan core.TradeDecision, 5)
	riskActions := make(chan core.RiskAction, 10)

	fetcher := datafeed.NewFetcher(datafeed.FetcherConfig{
		IntervalSeconds: 5,
		ScriptPath:      filepath.Join("scripts", "optionscraper.py"),
		PythonCmd:       "python3",
	})
	go fetcher.Run(ctx, snapshots)

	engine := signals.NewEngine(db)
	go engine.Run(ctx, snapshots, signalsCh)

	strategyEngine := strategy.NewStrategyEngine()
	go strategyEngine.Run(ctx, signalsCh, decisions)

	execEngine := execution.NewExecutionEngine(nil, true)
	go execEngine.Run(ctx, decisions, riskActions)

	go func() {
		for snap := range snapshots {
			log.Printf("SNAP spot=%.2f atm=%d chain=%d",
				snap.SpotPrice, snap.ATMStrike, len(snap.Chain))
		}
	}()

	go func() {
		for sig := range signalsCh {
			log.Printf("SIGNAL ivr=%.1f z=%.2f em=%.0f reg=%d dir=%d conf=%.2f mp=%d skew=%.1f",
				sig.IVRank, sig.IVZScore, sig.ExpectedMove,
				sig.Regime, sig.Direction, sig.Conviction,
				sig.MaxPain, sig.Skew)
		}
	}()

	go func() {
		for d := range decisions {
			log.Printf("TRADE strategy=%d lots=%d maxLoss=%.0f ev=%.2f reason=%s",
				d.Strategy, d.Lots, d.MaxLoss, d.ExpectedEV, d.Reason)
		}
	}()

	go func() {
		for a := range riskActions {
			log.Printf("RISK type=%d pos=%s reason=%s", a.Type, a.PositionID, a.Reason)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
	log.Println("shutting down...")
	cancel()
	time.Sleep(2 * time.Second)
}
