package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	ManagerPort     string
	WorkerURLs      []string
	ResponseTimeout time.Duration
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		ManagerPort: "8080",
		WorkerURLs: []string{
			"http://worker1:8081/internal/api/worker/hash/crack/task",
			"http://worker2:8081/internal/api/worker/hash/crack/task",
			"http://worker3:8081/internal/api/worker/hash/crack/task",
		},
		ResponseTimeout: 3 * time.Minute,
	}

	if port := os.Getenv("MANAGER_PORT"); port != "" {
		cfg.ManagerPort = port
	}
	if urls := os.Getenv("WORKER_URLS"); urls != "" {
		cfg.WorkerURLs = strings.Split(urls, ",")
	}
	if timeout := os.Getenv("RESPONSE_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			cfg.ResponseTimeout = d
		}
	}
	return cfg, nil
}
