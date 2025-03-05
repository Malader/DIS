package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"CrackHash/manager/internal/config"
	"CrackHash/manager/internal/handlers"
	"CrackHash/manager/internal/service"
	"CrackHash/manager/internal/store"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	reqStore := store.NewRequestStore()
	workerClient := service.NewMultiWorkerClient(cfg.WorkerURLs)
	mgrService := service.NewManagerService(reqStore, workerClient, cfg.ResponseTimeout)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/hash/crack", handlers.CrackHandler(ctx, mgrService))
	mux.HandleFunc("/api/hash/status", handlers.StatusHandler(ctx, reqStore))
	mux.HandleFunc("/internal/api/manager/hash/crack/request", handlers.WorkerResponseHandler(ctx, reqStore))

	srv := &http.Server{
		Addr:         ":" + cfg.ManagerPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Менеджер запускается на порту %s", cfg.ManagerPort)
	log.Fatal(srv.ListenAndServe())
}
