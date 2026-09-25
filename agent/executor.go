package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TaskParams struct {
	Path     string   `json:"path"`
	Pattern  string   `json:"pattern"`
	URL      string   `json:"url"`
	SHA256   string   `json:"sha256"`
	Args     []string `json:"args"`
	WorkDir  string   `json:"work_dir"`
}

func executeTask(t Task) (string, error) {
	var params TaskParams
	if err := json.Unmarshal([]byte(t.Params), &params); err != nil {
		return "", fmt.Errorf("invalid params: %v", err)
	}

	switch t.Type {
	case "search":
		return runSearch(params)
	case "wipe_file":
		return runWipeFile(params)
	case "wipe_disk":
		return runWipeDisk(params)
	case "execute_binary":
		return runExecuteBinary(params)
	default:
		return "", fmt.Errorf("unknown task type: %s", t.Type)
	}
}

func runSearch(p TaskParams) (string, error) {
	var found []string
	root := p.Path
	if root == "" {
		root = "/"
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			if p.Pattern != "" && strings.Contains(strings.ToLower(path), strings.ToLower(p.Pattern)) {
				found = append(found, path)
			}
		}
		return nil
	})

	res, _ := json.Marshal(map[string]any{"found_files": found})
	return string(res), err
}

func runWipeFile(p TaskParams) (string, error) {
	if p.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	info, err := os.Stat(p.Path)
	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(p.Path, os.O_WRONLY, 0)
	if err != nil {
		return "", err
	}
	defer f.Close()

	randomData := make([]byte, 4096)
	var written int64
	for written < info.Size() {
		rand.Read(randomData)
		n, _ := f.Write(randomData)
		written += int64(n)
	}
	f.Sync()
	f.Close()

	os.Remove(p.Path)
	return `{"status":"deleted"}`, nil
}

func runWipeDisk(p TaskParams) (string, error) {
	if p.Path == "" {
		return "", fmt.Errorf("disk path is required")
	}
	return fmt.Sprintf(`{"status":"simulated_wipe", "disk":"%s"}`, p.Path), nil
}

func runExecuteBinary(p TaskParams) (string, error) {
	if p.URL == "" {
		return "", fmt.Errorf("binary URL is required")
	}

	// 1. Ensure WorkDir exists
	workDir := p.WorkDir
	if workDir == "" {
		workDir = os.TempDir()
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create workdir: %v", err)
	}

	// 2. Download binary
	fileName := filepath.Base(p.URL)
	destPath := filepath.Join(workDir, fileName)
	
	if err := downloadFile(p.URL, destPath); err != nil {
		return "", fmt.Errorf("download failed: %v", err)
	}

	// 3. Verify SHA256
	if p.SHA256 != "" {
		if err := verifySHA256(destPath, p.SHA256); err != nil {
			os.Remove(destPath)
			return "", fmt.Errorf("integrity check failed: %v", err)
		}
	}

	// 4. Make executable (Unix)
	os.Chmod(destPath, 0755)

	// 5. Execute
	cmd := exec.Command(destPath, p.Args...)
	cmd.Dir = workDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf(`{"status":"error", "output": %q, "err": %q}`, 
			string(output), err.Error()), nil
	}

	return fmt.Sprintf(`{"status":"success", "output": %q}`, string(output)), nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func verifySHA256(path, expectedHex string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actualHex := hex.EncodeToString(h.Sum(nil))
	if actualHex != expectedHex {
		return fmt.Errorf("sha256 mismatch: expected %s, got %s", expectedHex, actualHex)
	}
	return nil
}
