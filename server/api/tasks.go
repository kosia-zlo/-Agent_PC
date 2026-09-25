package api

import (
	"encoding/json"
	"net/http"
	"myutil-server/models"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func (s *Server) handleAgentTasks(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "missing agent_id")
		return
	}

	tasks, err := s.DB.GetQueuedTasks(agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) handleTaskResult(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Status string `json:"status"`
		Result string `json:"result"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.DB.UpdateTaskResult(id, req.Status, req.Result); err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	s.DB.LogAudit("agent", "task_result", id, "status: "+req.Status)

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentID string `json:"agent_id"`
		Type    string `json:"type"`
		Params  string `json:"params"`
		Actor   string `json:"actor"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validation
	validTypes := map[string]bool{
		"search":          true, 
		"wipe_file":       true, 
		"wipe_disk":       true,
		"execute_binary":  true,
	}
	if !validTypes[req.Type] {
		writeError(w, http.StatusBadRequest, "invalid task type")
		return
	}

	if (req.Type == "wipe_file" || req.Type == "wipe_disk") && req.Actor == "" {
		writeError(w, http.StatusBadRequest, "actor is required for wipe operations")
		return
	}

	task := &models.Task{
		ID:        uuid.New().String(),
		AgentID:   req.AgentID,
		Type:      req.Type,
		Params:    req.Params,
		Status:    "queued",
		CreatedBy: req.Actor,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.DB.CreateTask(task); err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	s.DB.LogAudit(req.Actor, "task_create", task.ID, "type: "+req.Type)

	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if limit > 500 {
		limit = 500
	}

	tasks, err := s.DB.ListTasks(agentID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) handleListAudit(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 200
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if limit > 1000 {
		limit = 1000
	}

	entries, err := s.DB.ListAudit(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, entries)
}
