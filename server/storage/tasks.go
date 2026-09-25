package storage

import (
	"database/sql"
	"myutil-server/models"
	"time"
)

func (db *DB) CreateTask(t *models.Task) error {
	_, err := db.Conn.Exec(
		"INSERT INTO tasks (id, agent_id, type, params, status, created_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		t.ID, t.AgentID, t.Type, t.Params, t.Status, t.CreatedBy, t.CreatedAt,
	)
	return err
}

func (db *DB) GetQueuedTasks(agentID string) ([]models.Task, error) {
	rows, err := db.Conn.Query(
		"SELECT id, agent_id, type, params, status, created_by, created_at FROM tasks WHERE agent_id = ? AND status = 'queued'",
		agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.AgentID, &t.Type, &t.Params, &t.Status, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (db *DB) UpdateTaskResult(id, status, result string) error {
	now := time.Now().UTC()
	var query string
	var args []interface{}

	switch status {
	case "running":
		query = "UPDATE tasks SET status = ?, started_at = ? WHERE id = ?"
		args = []interface{}{status, now, id}
	case "done", "failed":
		query = "UPDATE tasks SET status = ?, result = ?, finished_at = ? WHERE id = ?"
		args = []interface{}{status, result, now, id}
	default:
		return sql.ErrNoRows
	}

	_, err := db.Conn.Exec(query, args...)
	return err
}

func (db *DB) ListTasks(agentID string, limit int) ([]models.Task, error) {
	query := "SELECT id, agent_id, type, params, status, result, created_by, created_at, started_at, finished_at FROM tasks"
	var args []interface{}
	if agentID != "" {
		query += " WHERE agent_id = ?"
		args = append(args, agentID)
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.Conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.AgentID, &t.Type, &t.Params, &t.Status, &t.Result, &t.CreatedBy, &t.CreatedAt, &t.StartedAt, &t.FinishedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (db *DB) ListAudit(limit int) ([]models.AuditEntry, error) {
	rows, err := db.Conn.Query("SELECT id, timestamp, actor, action, target, details FROM audit ORDER BY timestamp DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.AuditEntry
	for rows.Next() {
		var e models.AuditEntry
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.Actor, &e.Action, &e.Target, &e.Details); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
