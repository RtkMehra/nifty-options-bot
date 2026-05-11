package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func NewDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migrate db: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS vix_history (
			date TEXT PRIMARY KEY,
			vix REAL NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS trade_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			strategy TEXT NOT NULL,
			symbol TEXT NOT NULL,
			action TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			price REAL NOT NULL,
			premium REAL NOT NULL,
			reason TEXT,
			status TEXT NOT NULL DEFAULT 'open'
		)`,
		`CREATE TABLE IF NOT EXISTS option_chain_snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			spot_price REAL NOT NULL,
			india_vix REAL,
			data JSON NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_vix_history_date ON vix_history(date)`,
		`CREATE INDEX IF NOT EXISTS idx_trade_log_timestamp ON trade_log(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_snapshots_timestamp ON option_chain_snapshots(timestamp)`,
	}

	for _, s := range schema {
		if _, err := db.conn.Exec(s); err != nil {
			return fmt.Errorf("exec schema: %w", err)
		}
	}

	return nil
}

func (db *DB) Conn() *sql.DB {
	return db.conn
}
