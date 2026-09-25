package api

import (
	"encoding/json"
	"net"
	"net/http"
	"myutil-server/models"
	"time"

	"github.com/google/uuid"
)

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       string `json:"id"`
		Hostname string `json:"hostname"`
		OS       string `json:"os"`
		Arch     string `json:"arch"`
		Version  string `json:"version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	id := req.ID
	if id == "" {
		id = uuid.New().String()
	}

	// Get IP from RemoteAddr or X-Forwarded-For
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		} else {
			ip = host
		}
	}

	agent := &models.Agent{
		ID:       id,
		Hostname: req.Hostname,
		OS:       req.OS,
		Arch:     req.Arch,
		Version:  req.Version,
		IP:       ip,
		Ring:     "stable",
		Status:   "online",
		LastSeen: time.Now().UTC(),
	}

	if err := s.DB.UpsertAgent(agent); err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	s.DB.LogAudit("system", "register", agent.ID, "agent registered from "+ip)

	writeJSON(w, http.StatusOK, map[string]any{
		"id":            agent.ID,
		"heartbeat_sec": 30,
		"poll_sec":      10,
	})
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.DB.UpdateAgentSeen(req.ID, req.Version); err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := s.DB.ListAgents()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, agents)
}

func (s *Server) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	agent, err := s.DB.GetAgent(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if agent == nil {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}
	writeJSON(w, http.StatusOK, agent)
}
