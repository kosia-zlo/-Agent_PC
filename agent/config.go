package main

import (
	"flag"
	"os"
)

type Config struct {
	ServerURL string
	AgentID   string
}

func LoadConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.ServerURL, "server", "http://127.0.0.1:8080", "Server URL")
	flag.StringVar(&cfg.AgentID, "id", "", "Static Agent ID (optional)")
	flag.Parse()

	if envUrl := os.Getenv("SERVER_URL"); envUrl != "" {
		cfg.ServerURL = envUrl
	}
	return cfg
}
