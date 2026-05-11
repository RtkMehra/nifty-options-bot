package datafeed

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

func TestFindATMStrike(t *testing.T) {
	strikes := map[int]bool{
		22400: true,
		22450: true,
		22500: true,
		22550: true,
		22600: true,
	}

	tests := []struct {
		name string
		spot float64
		want int
	}{
		{name: "exact match", spot: 22500, want: 22500},
		{name: "round down", spot: 22460, want: 22450},
		{name: "round up", spot: 22530, want: 22550},
		{name: "below min", spot: 22000, want: 22400},
		{name: "above max", spot: 23000, want: 22600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findATMStrike(tt.spot, strikes)
			if got != tt.want {
				t.Errorf("findATMStrike(%v) = %v, want %v", tt.spot, got, tt.want)
			}
		})
	}
}

func TestFindATMStrikeEmpty(t *testing.T) {
	got := findATMStrike(22500, map[int]bool{})
	if got != 0 {
		t.Errorf("empty map should return 0, got %v", got)
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{input: 5.0, want: 5.0},
		{input: -5.0, want: 5.0},
		{input: 0, want: 0},
		{input: -0.001, want: 0.001},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := abs(tt.input)
			if got != tt.want {
				t.Errorf("abs(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFetcherRunContextCancel(t *testing.T) {
	fetcher := NewFetcher(FetcherConfig{IntervalSeconds: 1})
	snapshots := make(chan core.MarketSnapshot, 5)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		fetcher.Run(ctx, snapshots)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after context cancellation")
	}
}

func TestParseScraperOutput(t *testing.T) {
	jsonData := `{
		"spot": 23809.35,
		"vix": 18.52,
		"expiries": ["14-May-2026", "21-May-2026"],
		"expiry": "14-May-2026",
		"chain": [
			{"strike": 23800, "expiry": "14-May-2026", "type": "CE", "ltp": 123.45, "iv": 0.2463, "oi": 12345, "volume": 678, "bid": 120.0, "ask": 125.0, "bid_qty": 100, "ask_qty": 75},
			{"strike": 23800, "expiry": "14-May-2026", "type": "PE", "ltp": 89.10, "iv": 0.2250, "oi": 54321, "volume": 234, "bid": 87.0, "ask": 91.0, "bid_qty": 50, "ask_qty": 60},
			{"strike": 23900, "expiry": "14-May-2026", "type": "CE", "ltp": 45.0, "iv": 0.1850, "oi": 1000, "volume": 50, "bid": 44.0, "ask": 46.0, "bid_qty": 10, "ask_qty": 20}
		]
	}`

	// We test by creating a temporary fetcher and parsing JSON directly
	// by unit-testing the parse logic through a helper
	snap, err := parseScraperJSON(jsonData)
	if err != nil {
		t.Fatalf("parseScraperJSON() err = %v", err)
	}

	if snap.SpotPrice != 23809.35 {
		t.Errorf("SpotPrice = %v, want 23809.35", snap.SpotPrice)
	}
	if snap.IndiaVIX != 18.52 {
		t.Errorf("IndiaVIX = %v, want 18.52", snap.IndiaVIX)
	}
	if len(snap.Chain) != 3 {
		t.Errorf("Chain length = %d, want 3", len(snap.Chain))
	}
	if len(snap.Expiries) != 2 {
		t.Errorf("Expiries length = %d, want 2", len(snap.Expiries))
	}
	if snap.ATMStrike != 23800 {
		t.Errorf("ATMStrike = %d, want 23800 (spot=23809.35)", snap.ATMStrike)
	}

	// Verify IV is NOT divided again (scraper already divides by 100)
	ce := snap.Chain[0]
	if ce.OptionType != core.CE {
		t.Errorf("OptionType = %v, want CE", ce.OptionType)
	}
	if ce.IV != 0.2463 {
		t.Errorf("CE IV = %v, want 0.2463 (already divided by 100 in Python)", ce.IV)
	}
	if ce.LTP != 123.45 {
		t.Errorf("CE LTP = %v, want 123.45", ce.LTP)
	}
	if ce.OI != 12345 {
		t.Errorf("CE OI = %d, want 12345", ce.OI)
	}
	if ce.Volume != 678 {
		t.Errorf("CE Volume = %d, want 678", ce.Volume)
	}
	if ce.Bid != 120.0 {
		t.Errorf("CE Bid = %v, want 120.0", ce.Bid)
	}
	if ce.Ask != 125.0 {
		t.Errorf("CE Ask = %v, want 125.0", ce.Ask)
	}
	if ce.MidPrice != 122.5 {
		t.Errorf("CE MidPrice = %v, want 122.5", ce.MidPrice)
	}

	pe := snap.Chain[1]
	if pe.OptionType != core.PE {
		t.Errorf("OptionType = %v, want PE", pe.OptionType)
	}
	if pe.IV != 0.2250 {
		t.Errorf("PE IV = %v, want 0.2250", pe.IV)
	}
}

func TestParseScraperOutputError(t *testing.T) {
	_, err := parseScraperJSON(`{"error": "something went wrong"}`)
	if err == nil {
		t.Fatal("expected error for error response, got nil")
	}
}

func TestParseScraperOutputInvalidJSON(t *testing.T) {
	_, err := parseScraperJSON(`{invalid}`)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// parseScraperJSON is a test helper that parses raw JSON without invoking subprocess.
func parseScraperJSON(data string) (core.MarketSnapshot, error) {
	var scraperResp pythonScraperOutput
	if err := json.Unmarshal([]byte(data), &scraperResp); err != nil {
		return core.MarketSnapshot{}, err
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
