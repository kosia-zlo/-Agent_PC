package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

const Version = "v0.2.0"

func main() {
	cfg := LoadConfig()
	client := NewClient(cfg.ServerURL)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("get hostname: %v", err)
	}

	id, err := client.Register(cfg.AgentID, hostname, runtime.GOOS, runtime.GOARCH, Version)
	if err != nil {
		log.Fatalf("registration failed: %v", err)
	}
	fmt.Printf("Agent %s (%s) connected to %s\n", id, Version, cfg.ServerURL)

	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()
	pollTicker := time.NewTicker(10 * time.Second)
	defer pollTicker.Stop()

	if err := client.Heartbeat(id, Version); err != nil {
		log.Printf("initial heartbeat error: %v", err)
	}
	if err := pollAndRunTasks(client, id); err != nil {
		log.Printf("initial task poll error: %v", err)
	}

	for {
		select {
		case <-heartbeatTicker.C:
			if err := client.Heartbeat(id, Version); err != nil {
				log.Printf("heartbeat error: %v", err)
			}
		case <-pollTicker.C:
			if err := pollAndRunTasks(client, id); err != nil {
				log.Printf("polling error: %v", err)
			}
		}
	}
}

func pollAndRunTasks(client *Client, id string) error {
	tasks, err := client.PollTasks(id)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		log.Printf("Executing task %s (%s)...", task.ID, task.Type)
		if err := client.SendResult(task.ID, "running", ""); err != nil {
			log.Printf("error marking task %s as running: %v", task.ID, err)
			continue
		}

		result, err := executeTask(task)
		status := "done"
		if err != nil {
			log.Printf("Task %s failed: %v", task.ID, err)
			status = "failed"
			errorResult, marshalErr := json.Marshal(map[string]string{"error": err.Error()})
			if marshalErr != nil {
				return fmt.Errorf("encode task error: %w", marshalErr)
			}
			result = string(errorResult)
		}
		if err := client.SendResult(task.ID, status, result); err != nil {
			log.Printf("error sending result for task %s: %v", task.ID, err)
		}
	}
	return nil
}
