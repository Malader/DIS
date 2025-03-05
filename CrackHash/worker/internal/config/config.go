package config

import "os"

type Config struct {
	WorkerPort string
	ManagerURL string
}

func LoadConfig() (Config, error) {
	cfg := Config{
		WorkerPort: "8081",
		ManagerURL: "http://manager:8080/internal/api/manager/hash/crack/request",
	}
	if port := os.Getenv("WORKER_PORT"); port != "" {
		cfg.WorkerPort = port
	}
	if url := os.Getenv("MANAGER_URL"); url != "" {
		cfg.ManagerURL = url
	}
	return cfg, nil
}
