package config

import "os"

type Config struct {
	WorkerPort       string
	ManagerURL       string
	RabbitURI        string
	TaskExchange     string
	TaskQueueName    string
	ResponseExchange string
	ResponseQueue    string
}

func LoadConfig() (Config, error) {
	cfg := Config{
		WorkerPort:       "8081",
		ManagerURL:       "http://manager:8080/internal/api/manager/hash/crack/request",
		RabbitURI:        "amqp://guest:guest@rabbitmq:5672/",
		TaskExchange:     "tasks_direct",
		TaskQueueName:    "task_queue",
		ResponseExchange: "responses_direct",
		ResponseQueue:    "worker_responses",
	}

	if port := os.Getenv("WORKER_PORT"); port != "" {
		cfg.WorkerPort = port
	}
	if url := os.Getenv("MANAGER_URL"); url != "" {
		cfg.ManagerURL = url
	}
	if rURI := os.Getenv("RABBIT_URI"); rURI != "" {
		cfg.RabbitURI = rURI
	}
	return cfg, nil
}
