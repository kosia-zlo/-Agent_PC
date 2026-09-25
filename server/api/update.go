package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"myutil-server/models"
)

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	version := r.URL.Query().Get("version")
	platform := r.URL.Query().Get("platform") // os/arch

	if version == "" || platform == "" {
		writeError(w, http.StatusBadRequest, "missing version or platform")
		return
	}

	latestPath := filepath.Join(s.Config.UpdateDir, "latest.txt")
	content, err := os.ReadFile(latestPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update server error")
		return
	}
	latestVersion := strings.TrimSpace(string(content))

	if version == latestVersion {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	parts := strings.Split(platform, "/")
	if len(parts) != 2 {
		writeError(w, http.StatusBadRequest, "invalid platform format")
		return
	}
	osName, arch := parts[0], parts[1]

	manifestName := fmt.Sprintf("%s-%s-%s.json", latestVersion, osName, arch)
	manifestPath := filepath.Join(s.Config.UpdateDir, manifestName)
	
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		writeError(w, http.StatusNotFound, "no update for this platform")
		return
	}

	var info models.UpdateInfo
	if err := json.Unmarshal(manifestData, &info); err != nil {
		writeError(w, http.StatusInternalServerError, "manifest error")
		return
	}

	binaryPath := filepath.Join(s.Config.UpdateDir, info.URL)
	if _, err := os.Stat(binaryPath); err == nil {
		b, _ := os.ReadFile(binaryPath)
		hash := sha256.Sum256(b)
		info.SHA256 = hex.EncodeToString(hash[:])
	}

	writeJSON(w, http.StatusOK, info)
}
