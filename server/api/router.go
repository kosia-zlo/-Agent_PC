package api

import (
	"encoding/json"
	"log"
	"net/http"
	"myutil-server/config"
	"myutil-server/storage"
)

type Server struct {
	DB     *storage.DB
	Config *config.Config
}

func NewServer(db *storage.DB, cfg *config.Config) *Server {
	return &Server{DB: db, Config: cfg}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// Agent endpoints
	mux.HandleFunc("POST /api/agent/register", s.handleRegister)
	mux.HandleFunc("POST /api/agent/heartbeat", s.handleHeartbeat)
	mux.HandleFunc("GET /api/agent/tasks", s.handleAgentTasks)
	mux.HandleFunc("POST /api/agent/tasks/{id}/result", s.handleTaskResult)
	mux.HandleFunc("GET /api/agent/update", s.handleUpdate)

	// Operator endpoints
	mux.HandleFunc("GET /api/agents", s.handleListAgents)
	mux.HandleFunc("GET /api/agents/{id}", s.handleGetAgent)
	mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	mux.HandleFunc("GET /api/tasks", s.handleListTasks)
	mux.HandleFunc("GET /api/audit", s.handleListAudit)

	// Web UI
	mux.HandleFunc("/", ServeWeb)

	return LoggingMiddleware(mux)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if v != nil {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			log.Printf("JSON encode error: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
