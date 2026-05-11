package store

import (
	"os"
	"testing"
	"time"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	tmp := os.TempDir()
	db, err := NewDB(tmp + "/nifty_test_" + t.Name() + ".db")
	if err != nil {
		t.Fatalf("NewDB() err = %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(tmp + "/nifty_test_" + t.Name() + ".db")
	})
	return db
}

func TestNewDB(t *testing.T) {
	db := newTestDB(t)
	if db == nil {
		t.Fatal("db is nil")
	}
}

func TestSaveAndGetVIXRange(t *testing.T) {
	db := newTestDB(t)

	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	if err := db.SaveVIX(yesterday, 12.5); err != nil {
		t.Fatalf("SaveVIX() err = %v", err)
	}
	if err := db.SaveVIX(today, 25.0); err != nil {
		t.Fatalf("SaveVIX() err = %v", err)
	}

	low, high, err := db.GetVIXRange()
	if err != nil {
		t.Fatalf("GetVIXRange() err = %v", err)
	}

	if low != 12.5 {
		t.Errorf("low = %v, want 12.5", low)
	}
	if high != 25.0 {
		t.Errorf("high = %v, want 25.0", high)
	}
}

func TestGetRecentIVHistory(t *testing.T) {
	db := newTestDB(t)
	today := time.Now()

	for i := 0; i < 10; i++ {
		day := today.AddDate(0, 0, -i)
		vix := 15.0 + float64(i)
		if err := db.SaveVIX(day, vix); err != nil {
			t.Fatalf("SaveVIX() err = %v", err)
		}
	}

	history, err := db.GetRecentIVHistory(30)
	if err != nil {
		t.Fatalf("GetRecentIVHistory() err = %v", err)
	}

	if len(history) != 10 {
		t.Errorf("got %d records, want 10", len(history))
		t.Logf("history: %v", history)
	}
}

func TestLogAndGetOpenTrades(t *testing.T) {
	db := newTestDB(t)

	trade := TradeRecord{
		Timestamp: time.Now(),
		Strategy:  "LongCall",
		Symbol:    "NIFTY25JAN22500CE",
		Action:    "BUY",
		Quantity:  75,
		Price:     185.0,
		Premium:   13875.0,
		Reason:    "IVR=25 conviction=0.75",
		Status:    "open",
	}

	if err := db.LogTrade(trade); err != nil {
		t.Fatalf("LogTrade() err = %v", err)
	}

	trades, err := db.GetOpenTrades()
	if err != nil {
		t.Fatalf("GetOpenTrades() err = %v", err)
	}

	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}

	if trades[0].Symbol != "NIFTY25JAN22500CE" {
		t.Errorf("symbol = %v, want NIFTY25JAN22500CE", trades[0].Symbol)
	}
}

func TestMultipleTrades(t *testing.T) {
	db := newTestDB(t)

	trades := []TradeRecord{
		{Timestamp: time.Now(), Strategy: "LongCall", Symbol: "A", Action: "BUY", Quantity: 75, Price: 100, Premium: 7500, Status: "open"},
		{Timestamp: time.Now(), Strategy: "LongPut", Symbol: "B", Action: "BUY", Quantity: 75, Price: 90, Premium: 6750, Status: "open"},
	}

	for _, tr := range trades {
		if err := db.LogTrade(tr); err != nil {
			t.Fatalf("LogTrade() err = %v", err)
		}
	}

	open, err := db.GetOpenTrades()
	if err != nil {
		t.Fatalf("GetOpenTrades() err = %v", err)
	}

	if len(open) != 2 {
		t.Errorf("got %d open trades, want 2", len(open))
	}
}
