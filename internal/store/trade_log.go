package store

import (
	"fmt"
	"time"
)

type TradeRecord struct {
	ID        int64
	Timestamp time.Time
	Strategy  string
	Symbol    string
	Action    string
	Quantity  int
	Price     float64
	Premium   float64
	Reason    string
	Status    string
}

func (db *DB) LogTrade(t TradeRecord) error {
	_, err := db.conn.Exec(
		`INSERT INTO trade_log (timestamp, strategy, symbol, action, quantity, price, premium, reason, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Timestamp.Format(time.RFC3339),
		t.Strategy, t.Symbol, t.Action, t.Quantity,
		t.Price, t.Premium, t.Reason, t.Status,
	)
	if err != nil {
		return fmt.Errorf("log trade: %w", err)
	}
	return nil
}

func (db *DB) GetOpenTrades() ([]TradeRecord, error) {
	rows, err := db.conn.Query(
		`SELECT id, timestamp, strategy, symbol, action, quantity, price, premium, reason, status
		FROM trade_log WHERE status = 'open' ORDER BY timestamp DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("get open trades: %w", err)
	}
	defer rows.Close()

	var trades []TradeRecord
	for rows.Next() {
		var t TradeRecord
		var ts string
		if err := rows.Scan(&t.ID, &ts, &t.Strategy, &t.Symbol,
			&t.Action, &t.Quantity, &t.Price, &t.Premium, &t.Reason, &t.Status); err != nil {
			return nil, fmt.Errorf("scan trade: %w", err)
		}
		t.Timestamp, _ = time.Parse(time.RFC3339, ts)
		trades = append(trades, t)
	}
	return trades, rows.Err()
}
