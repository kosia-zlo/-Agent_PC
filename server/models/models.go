package models

import (
	"time"
)

type Agent struct {
	ID              string    `json:"id"`
	Hostname        string    `json:"hostname"`
	OS              string    `json:"os"`
	Arch            string    `json:"arch"`
	Version         string    `json:"version"`
	IP              string    `json:"ip"`
	Ring            string    `json:"ring"`
	Status          string    `json:"status"`
	LastSeen        time.Time `json:"last_seen"`
	CertFingerprint string    `json:"cert_fingerprint,omitempty"`
}

type Task struct {
	ID         string    `json:"id"`
	AgentID    string    `json:"agent_id"`
	Type       string    `json:"type"`
	Params     string    `json:"params"`
	Status     string    `json:"status"`
	Result     string    `json:"result,omitempty"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

type AuditEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Details   string    `json:"details"`
}

type UpdateInfo struct {
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
	URL     string `json:"url"`
}
