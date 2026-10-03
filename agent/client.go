package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	URL        string
	HTTPClient *http.Client
}

func NewClient(url string) *Client {
	return &Client{
		URL: strings.TrimRight(url, "/"),
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Register(id, hostname, os, arch, version string) (string, error) {
	body, err := json.Marshal(map[string]string{
		"id":       id,
		"hostname": hostname,
		"os":       os,
		"arch":     arch,
		"version":  version,
	})
	if err != nil {
		return "", fmt.Errorf("encode registration: %w", err)
	}

	resp, err := c.HTTPClient.Post(c.URL+"/api/agent/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registration failed: server returned %s", resp.Status)
	}

	var res struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	if res.ID == "" {
		return "", fmt.Errorf("registration failed: server returned an empty agent ID")
	}
	return res.ID, nil
}

func (c *Client) Heartbeat(id, version string) error {
	body, err := json.Marshal(map[string]string{
		"id":      id,
		"version": version,
	})
	if err != nil {
		return fmt.Errorf("encode heartbeat: %w", err)
	}
	resp, err := c.HTTPClient.Post(c.URL+"/api/agent/heartbeat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed: server returned %s", resp.Status)
	}
	return nil
}

func (c *Client) PollTasks(id string) ([]Task, error) {
	resp, err := c.HTTPClient.Get(c.URL + "/api/agent/tasks?agent_id=" + url.QueryEscape(id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("task poll failed: server returned %s", resp.Status)
	}

	var tasks []Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c *Client) SendResult(taskID, status, result string) error {
	body, err := json.Marshal(map[string]string{
		"status": status,
		"result": result,
	})
	if err != nil {
		return fmt.Errorf("encode task result: %w", err)
	}
	resp, err := c.HTTPClient.Post(c.URL+"/api/agent/tasks/"+url.PathEscape(taskID)+"/result", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sending task result failed: server returned %s", resp.Status)
	}
	return nil
}

type Task struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Params string `json:"params"`
}
