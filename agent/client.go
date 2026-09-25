package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	URL        string
	HTTPClient *http.Client
}

func NewClient(url string) *Client {
	return &Client{
		URL: url,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Register(hostname, os, arch, version string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"hostname": hostname,
		"os":       os,
		"arch":     arch,
		"version":  version,
	})

	resp, err := c.HTTPClient.Post(c.URL+"/api/agent/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.ID, nil
}

func (c *Client) Heartbeat(id, version string) error {
	body, _ := json.Marshal(map[string]string{
		"id":      id,
		"version": version,
	})
	resp, err := c.HTTPClient.Post(c.URL+"/api/agent/heartbeat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) PollTasks(id string) ([]Task, error) {
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/api/agent/tasks?agent_id=%s", c.URL, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tasks []Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c *Client) SendResult(taskID, status, result string) error {
	body, _ := json.Marshal(map[string]string{
		"status": status,
		"result": result,
	})
	resp, err := c.HTTPClient.Post(fmt.Sprintf("%s/api/agent/tasks/%s/result", c.URL, taskID), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

type Task struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Params string `json:"params"`
}
