package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

const Version = "v0.1.0"

func main() {
	cfg := LoadConfig()
	client := NewClient(cfg.ServerURL)

	hostname, _ := os.Hostname()
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// 1. Registration
	id := cfg.AgentID
	if id == "" {
		var err error
		id, err = client.Register(hostname, osName, arch, Version)
		if err != nil {
			log.Fatalf("registration failed: %v", err)
		}
	}
	fmt.Printf("Agent started. ID: %s, Server: %s\n", id, cfg.ServerURL)

	// Heartbeat ticker
	heartbeatTicker := time.NewTicker(30 * time.Second)
	// Poll ticker
	pollTicker := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-heartbeatTicker.C:
			if err := client.Heartbeat(id, Version); err != nil {
				log.Printf("heartbeat error: %v", err)
			}
		case <-pollTicker.C:
			tasks, err := client.PollTasks(id)
			if err != nil {
				log.Printf("polling error: %v", err)
				continue
			}

			for _, t := range tasks {
				log.Printf("Executing task %s (%s)...", t.ID, t.Type)
				
				// Mark as running
				client.SendResult(t.ID, "running", "")

				result, err := executeTask(t)
				status := "done"
				if err != nil {
					log.Printf("Task %s failed: %v", t.ID, err)
					status = "failed"
					result = fmt.Sprintf(`{"error":"%s"}`, err.Error())
				}

				if err := client.SendResult(t.ID, status, result); err != nil {
					log.Printf("error sending result: %v", err)
				}
			}
		}
	}
}
