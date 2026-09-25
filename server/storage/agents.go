package storage

import (
	"database/sql"
	"myutil-server/models"
	"time"
)

func (db *DB) UpsertAgent(a *models.Agent) error {
	_, err := db.Conn.Exec(
		`INSERT INTO agents (id, hostname, os, arch, version, ip, ring, status, last_seen, cert_fingerprint)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			hostname=excluded.hostname,
			os=excluded.os,
			arch=excluded.arch,
			version=excluded.version,
			ip=excluded.ip,
			status=excluded.status,
			last_seen=excluded.last_seen`,
		a.ID, a.Hostname, a.OS, a.Arch, a.Version, a.IP, a.Ring, a.Status, a.LastSeen, a.CertFingerprint,
	)
	return err
}

func (db *DB) UpdateAgentSeen(id, version string) error {
	_, err := db.Conn.Exec(
		"UPDATE agents SET last_seen = ?, status = 'online', version = ? WHERE id = ?",
		time.Now().UTC(), version, id,
	)
	return err
}

func (db *DB) ListAgents() ([]models.Agent, error) {
	rows, err := db.Conn.Query("SELECT id, hostname, os, arch, version, ip, ring, status, last_seen, cert_fingerprint FROM agents")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []models.Agent
	for rows.Next() {
		var a models.Agent
		if err := rows.Scan(&a.ID, &a.Hostname, &a.OS, &a.Arch, &a.Version, &a.IP, &a.Ring, &a.Status, &a.LastSeen, &a.CertFingerprint); err != nil {
			return nil, err
		}
		agents = append(agents, a)
	}
	return agents, nil
}

func (db *DB) GetAgent(id string) (*models.Agent, error) {
	var a models.Agent
	err := db.Conn.QueryRow(
		"SELECT id, hostname, os, arch, version, ip, ring, status, last_seen, cert_fingerprint FROM agents WHERE id = ?",
		id,
	).Scan(&a.ID, &a.Hostname, &a.OS, &a.Arch, &a.Version, &a.IP, &a.Ring, &a.Status, &a.LastSeen, &a.CertFingerprint)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (db *DB) MarkAgentsOffline() error {
	_, err := db.Conn.Exec(
		"UPDATE agents SET status = 'offline' WHERE last_seen < datetime('now', '-90 seconds')",
	)
	return err
}
