package store

import (
	"fmt"
	"time"
)

func (db *DB) SaveVIX(date time.Time, vix float64) error {
	_, err := db.conn.Exec(
		`INSERT OR REPLACE INTO vix_history (date, vix) VALUES (?, ?)`,
		date.Format("2006-01-02"), vix,
	)
	if err != nil {
		return fmt.Errorf("save vix: %w", err)
	}
	return nil
}

func (db *DB) GetVIXRange() (low52, high52 float64, err error) {
	query := `SELECT MIN(vix), MAX(vix) FROM vix_history
		WHERE date >= date('now', '-365 days')`
	err = db.conn.QueryRow(query).Scan(&low52, &high52)
	if err != nil {
		return 0, 0, fmt.Errorf("get vix range: %w", err)
	}
	return low52, high52, nil
}

func (db *DB) GetRecentIVHistory(days int) ([]float64, error) {
	query := `SELECT vix FROM vix_history
		WHERE date >= date('now', ?)
		ORDER BY date ASC`
	rows, err := db.conn.Query(query, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, fmt.Errorf("get iv history: %w", err)
	}
	defer rows.Close()

	var history []float64
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scan iv: %w", err)
		}
		history = append(history, v)
	}
	return history, rows.Err()
}
