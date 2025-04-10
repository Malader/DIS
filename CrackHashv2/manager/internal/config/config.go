package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	ManagerPort        string
	WorkerURLs         []string
	ResponseTimeout    time.Duration
	MongoURI           string
	MongoDatabase      string
	RabbitURI          string
	TaskExchange       string
	TaskQueueName      string
	ResponseExchange   string
	ResponseQueueName  string
	ReplicationTimeout time.Duration
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		ManagerPort: "8080",
		WorkerURLs: []string{
			"http://worker1:8081/internal/api/worker/hash/crack/task",
			"http://worker2:8081/internal/api/worker/hash/crack/task",
			"http://worker3:8081/internal/api/worker/hash/crack/task",
		},
		ResponseTimeout:    3 * time.Minute,
		MongoURI:           "mongodb://mongo1:27017,mongo2:27017,mongo3:27017/?replicaSet=rs0",
		MongoDatabase:      "crackhash",
		RabbitURI:          "amqp://guest:guest@rabbitmq:5672/",
		TaskExchange:       "tasks_direct",
		TaskQueueName:      "task_queue",
		ResponseExchange:   "responses_direct",
		ResponseQueueName:  "worker_responses",
		ReplicationTimeout: 2 * time.Second,
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
	if mongoURI := os.Getenv("MONGO_URI"); mongoURI != "" {
		cfg.MongoURI = mongoURI
	}
	if db := os.Getenv("MONGO_DB"); db != "" {
		cfg.MongoDatabase = db
	}
	if rabbitURI := os.Getenv("RABBIT_URI"); rabbitURI != "" {
		cfg.RabbitURI = rabbitURI
	}
	return cfg, nil
}
