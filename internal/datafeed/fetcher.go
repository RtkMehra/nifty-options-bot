package datafeed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

type FetcherConfig struct {
	PythonCmd       string
	ScriptPath      string
	IntervalSeconds int
}

type Fetcher struct {
	pythonCmd  string
	scriptPath string
	interval   time.Duration
}

func NewFetcher(cfg FetcherConfig) *Fetcher {
	if cfg.PythonCmd == "" {
		cfg.PythonCmd = "python3"
	}
	return &Fetcher{
		pythonCmd:  cfg.PythonCmd,
		scriptPath: cfg.ScriptPath,
		interval:   time.Duration(cfg.IntervalSeconds) * time.Second,
	}
}

func (f *Fetcher) Run(ctx context.Context, out chan<- core.MarketSnapshot) {
	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			snap, err := f.fetch()
			if err != nil {
				continue
			}
			select {
			case out <- snap:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

type pythonScraperOutput struct {
	Spot     float64            `json:"spot"`
	VIX      float64            `json:"vix"`
	Expiries []string           `json:"expiries"`
	Expiry   string             `json:"expiry"`
	Chain    []pythonOptionItem `json:"chain"`
	Error    string             `json:"error,omitempty"`
}

type pythonOptionItem struct {
	Strike   int     `json:"strike"`
	Expiry   string  `json:"expiry"`
	Type     string  `json:"type"`
	LTP      float64 `json:"ltp"`
	IV       float64 `json:"iv"`
	OI       int64   `json:"oi"`
	Volume   int64   `json:"volume"`
	Bid      float64 `json:"bid"`
	Ask      float64 `json:"ask"`
	BidQty   int64   `json:"bid_qty"`
	AskQty   int64   `json:"ask_qty"`
}

func (f *Fetcher) fetch() (core.MarketSnapshot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, f.pythonCmd, f.scriptPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return core.MarketSnapshot{}, fmt.Errorf("scraper failed: %w\nstderr: %s", err, stderr.String())
	}

	var scraperResp pythonScraperOutput
	if err := json.Unmarshal(stdout.Bytes(), &scraperResp); err != nil {
		return core.MarketSnapshot{}, fmt.Errorf("decode scraper output: %w", err)
	}

	if scraperResp.Error != "" {
		return core.MarketSnapshot{}, fmt.Errorf("scraper error: %s", scraperResp.Error)
	}

	var expiry time.Time
	if scraperResp.Expiry != "" {
		expiry, _ = time.Parse("2-Jan-2006", scraperResp.Expiry)
	}

	snap := core.MarketSnapshot{
		Timestamp: time.Now(),
		SpotPrice: scraperResp.Spot,
		IndiaVIX:  scraperResp.VIX,
	}

	for _, dateStr := range scraperResp.Expiries {
		e, err := time.Parse("2-Jan-2006", dateStr)
		if err == nil {
			snap.Expiries = append(snap.Expiries, e)
		}
	}

	strikes := make(map[int]bool)
	for _, item := range scraperResp.Chain {
		var optExpiry time.Time
		if item.Expiry != "" {
			optExpiry, _ = time.Parse("2-Jan-2006", item.Expiry)
		}
		if optExpiry.IsZero() {
			optExpiry = expiry
		}

		snap.Chain = append(snap.Chain, core.OptionData{
			Strike:     item.Strike,
			Expiry:     optExpiry,
			OptionType: core.OptionType(item.Type),
			LTP:        item.LTP,
			IV:         item.IV,
			OI:         item.OI,
			Volume:     item.Volume,
			Bid:        item.Bid,
			Ask:        item.Ask,
			MidPrice:   core.MidPrice(item.Bid, item.Ask),
		})
		strikes[item.Strike] = true
	}

	snap.ATMStrike = findATMStrike(snap.SpotPrice, strikes)

	return snap, nil
}

func findATMStrike(spot float64, strikes map[int]bool) int {
	nearest := 0
	minDiff := 1e9
	for s := range strikes {
		diff := abs(float64(s) - spot)
		if diff < minDiff {
			minDiff = diff
			nearest = s
		}
	}
	return nearest
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
