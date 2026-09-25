package storage

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

type DB struct {
	Conn *sql.DB
}

func NewDB(path string) (*DB, error) {
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &DB{Conn: db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}

	return s, nil
}

func (db *DB) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			hostname TEXT,
			os TEXT,
			arch TEXT,
			version TEXT,
			ip TEXT,
			ring TEXT DEFAULT 'stable',
			status TEXT,
			last_seen DATETIME,
			cert_fingerprint TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			agent_id TEXT,
			type TEXT,
			params TEXT,
			status TEXT,
			result TEXT,
			created_by TEXT,
			created_at DATETIME,
			started_at DATETIME,
			finished_at DATETIME,
			FOREIGN KEY(agent_id) REFERENCES agents(id)
		);`,
		`CREATE TABLE IF NOT EXISTS audit (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			actor TEXT,
			action TEXT,
			target TEXT,
			details TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks(agent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_ts ON audit(timestamp);`,
	}

	for _, q := range queries {
		if _, err := db.Conn.Exec(q); err != nil {
			return fmt.Errorf("schema init error: %v", err)
		}
	}
	return nil
}

func (db *DB) LogAudit(actor, action, target, details string) {
	_, err := db.Conn.Exec(
		"INSERT INTO audit (actor, action, target, details, timestamp) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)",
		actor, action, target, details,
	)
	if err != nil {
		log.Printf("Audit log error: %v", err)
	}
}
