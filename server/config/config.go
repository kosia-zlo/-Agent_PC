package config

import (
	"os"
)

type Config struct {
	ListenAddr  string
	DBPath      string
	UpdateDir   string
	UpdatePubkey string
}

func Load() *Config {
	return &Config{
		ListenAddr:   getEnv("LISTEN_ADDR", "127.0.0.1:8080"),
		DBPath:       getEnv("DB_PATH", "server.db"),
		UpdateDir:    getEnv("UPDATE_DIR", "./updates"),
		UpdatePubkey: getEnv("UPDATE_PUBKEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
