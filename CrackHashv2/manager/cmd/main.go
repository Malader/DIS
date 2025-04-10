package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"CrackHash/manager/internal/config"
	"CrackHash/manager/internal/handlers"
	"CrackHash/manager/internal/queue"
	"CrackHash/manager/internal/service"
	"CrackHash/manager/internal/store"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	mongoStore, err := store.NewMongoRequestStore(cfg)
	if err != nil {
		log.Fatalf("Ошибка инициализации MongoStore: %v", err)
	}

	rabbitClient, err := queue.NewRabbitClient(cfg)
	if err != nil {
		log.Printf("Не удалось подключиться к RabbitMQ при старте: %v", err)
	}

	mgrService := service.NewManagerService(mongoStore, rabbitClient, cfg.ResponseTimeout)

	go func() {
		for {
			time.Sleep(5 * time.Second)
			if rabbitClient != nil && rabbitClient.IsConnected() {
				err := mgrService.RetryPendingTasks(context.Background())
				if err != nil {
					log.Printf("Ошибка при повторной отправке зависших задач: %v", err)
				}
			}
		}
	}()

	if rabbitClient != nil && rabbitClient.IsConnected() {
		respCh, err := rabbitClient.StartConsumeResponses()
		if err != nil {
			log.Printf("Не удалось подписаться на worker_responses: %v", err)
		} else {
			go func() {
				for workerResp := range respCh {
					state, ok := mongoStore.Get(workerResp.RequestId)
					if ok {
						if state.Data == nil {
							state.Data = []string{}
						}
						state.Data = append(state.Data, workerResp.Answers.Words...)
						if len(state.Data) > 0 {
							state.Status = service.StatusReady
							if state.Timer != nil {
								state.Timer.Stop()
							}
						}
						mongoStore.Update(workerResp.RequestId, state)
						if state.Status != "IN_PROGRESS" {
							mongoStore.MarkPending(workerResp.RequestId, false)
						}
					}
					rabbitClient.AckMessage(workerResp)

				}
			}()
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/hash/crack", handlers.CrackHandler(ctx, mgrService))
	mux.HandleFunc("/api/hash/status", handlers.StatusHandler(ctx, mongoStore))
	mux.HandleFunc("/internal/api/manager/hash/crack/request", handlers.WorkerResponseHandler(ctx, mongoStore))

	srv := &http.Server{
		Addr:         ":" + cfg.ManagerPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Менеджер запускается на порту %s", cfg.ManagerPort)
	log.Fatal(srv.ListenAndServe())
}
