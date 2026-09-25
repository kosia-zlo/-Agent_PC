package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	ModelPath  string
	EnginePath string
	TargetDir  string
	OutputFile string
}

type AIResponse struct {
	Content string `json:"content"`
}

type FileResult struct {
	Path     string `json:"path"`
	IsSensitive bool   `json:"is_sensitive"`
	Reason      string `json:"reason"`
}

func main() {
	cfg := &Config{}
	flag.StringVar(&cfg.ModelPath, "model", "", "Path to GGUF model")
	flag.StringVar(&cfg.EnginePath, "engine", "llama-server", "Path to llama-server binary")
	flag.StringVar(&cfg.TargetDir, "dir", "", "Directory to scan")
	flag.StringVar(&cfg.OutputFile, "out", "report.json", "Output file")
	flag.Parse()

	if cfg.ModelPath == "" || cfg.TargetDir == "" {
		log.Fatal("Usage: ai-scanner -model <path> -dir <path> [-engine <path>]")
	}

	// 1. Start llama-server
	fmt.Println("[*] Starting AI Engine...")
	cmd := exec.Command(cfg.EnginePath, "-m", cfg.ModelPath, "--port", "8081", "--nobrowser")
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start engine: %v", err)
	}
	defer func() {
		process, _ := os.FindProcess(cmd.Process.Pid)
		process.Kill()
	}()

	// Wait for server to be ready
	for i := 0; i < 10; i++ {
		resp, err := http.Get("http://localhost:8081/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(2 * time.Second)
		if i == 9 {
			log.Fatal("AI Engine failed to start in time")
		}
	}

	// 2. Scan files
	fmt.Printf("[*] Scanning directory: %s\n", cfg.TargetDir)
	var results []FileResult

	err = filepath.WalkDir(cfg.TargetDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		// Only scan text-like files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".txt" && ext != ".conf" && ext != ".log" && ext != ".json" && ext != ".xml" {
			return nil
		}

		content, err := readSnippet(path)
		if err != nil {
			return nil
		}

		res, err := analyzeText(content)
		if err != nil {
			log.Printf("Error analyzing %s: %v", path, err)
			return nil
		}

		if res.IsSensitive {
			fmt.Printf("[!] Sensitive file found: %s\n", path)
			results = append(results, FileResult{
				Path:        path,
				IsSensitive: true,
				Reason:      res.Reason,
			})
		}
		return nil
	})

	if err != nil {
		log.Fatalf("WalkDir error: %v", err)
	}

	// 3. Save report
	report, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile(cfg.OutputFile, report, 0644)
	fmt.Printf("[+] Scan complete. Report saved to %s\n", cfg.OutputFile)
}

func readSnippet(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 1024)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	return string(buf[:n]), nil
}

func analyzeText(text string) (FileResult, error) {
	prompt := fmt.Sprintf(
		"Analyze the following text and determine if it contains sensitive information (passwords, API keys, private data). "+
			"Respond ONLY in JSON format: {\"is_sensitive\": true/false, \"reason\": \"short explanation\"}. "+
			"Text: %s", text,
	)

	payload := map[string]any{
		"prompt": prompt,
		"n_predict": 64,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8081/completion", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return FileResult{}, err
	}
	defer resp.Body.Close()

	var aiResp AIResponse
	json.NewDecoder(resp.Body).Decode(&aiResp)

	// Try to parse JSON from AI content
	var res FileResult
	if err := json.Unmarshal([]byte(aiResp.Content), &res); err != nil {
		// Fallback: simple keyword check if AI failed to return JSON
		if strings.Contains(strings.ToLower(aiResp.Content), "true") {
			return FileResult{IsSensitive: true, Reason: aiResp.Content}, nil
		}
		return FileResult{IsSensitive: false}, nil
	}

	return res, nil
}
