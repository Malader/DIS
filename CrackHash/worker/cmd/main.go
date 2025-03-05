package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"CrackHash/worker/internal/config"
	"CrackHash/worker/internal/handlers"
	"CrackHash/worker/internal/service"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	workerSvc := service.NewWorkerService()

	mux := http.NewServeMux()
	mux.HandleFunc("/internal/api/worker/hash/crack/task", handlers.TaskHandler(ctx, workerSvc))

	srv := &http.Server{
		Addr:         ":" + cfg.WorkerPort,
		Handler:      mux,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
	}

	log.Printf("Worker запускается на порту %s", cfg.WorkerPort)
	log.Fatal(srv.ListenAndServe())
}
