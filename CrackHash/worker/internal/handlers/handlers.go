package handlers

import (
	"CrackHash/worker/internal/config"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"time"

	"CrackHash/worker/internal/types"
)

type WorkerService interface {
	ProcessTask(hash string, maxLength int, alphabet []string, partNumber, partCount int) []string
}

func TaskHandler(ctx context.Context, svc WorkerService) http.HandlerFunc {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации воркера: %v", err)
	}
	managerURL := cfg.ManagerURL

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
			return
		}
		var req types.CrackHashManagerRequest
		err = json.Unmarshal(body, &req)
		if err != nil {
			http.Error(w, "Некорректный JSON", http.StatusBadRequest)
			return
		}

		results := svc.ProcessTask(req.Hash, req.MaxLength, req.Alphabet.Symbols, req.PartNumber, req.PartCount)

		response := types.CrackHashWorkerResponse{
			RequestId:  req.RequestId,
			PartNumber: req.PartNumber,
		}
		response.Answers.Words = results

		xmlData, err := xml.MarshalIndent(response, "", "  ")
		if err != nil {
			log.Printf("Ошибка маршалинга XML: %v", err)
			http.Error(w, "Ошибка обработки", http.StatusInternalServerError)
			return
		}
		xmlData = append([]byte(xml.Header), xmlData...)

		patchReq, err := http.NewRequest(http.MethodPatch, managerURL, bytes.NewReader(xmlData))
		if err != nil {
			log.Printf("Ошибка создания PATCH запроса: %v", err)
			http.Error(w, "Ошибка отправки результата", http.StatusInternalServerError)
			return
		}
		patchReq.Header.Set("Content-Type", "application/xml")
		client := &http.Client{Timeout: 10 * time.Second}
		patchResp, err := client.Do(patchReq)
		if err != nil {
			log.Printf("Ошибка отправки PATCH запроса: %v", err)
			http.Error(w, "Ошибка отправки результата", http.StatusInternalServerError)
			return
		}
		defer patchResp.Body.Close()
		if patchResp.StatusCode != http.StatusOK {
			log.Printf("Менеджер вернул ошибку на PATCH запрос: %s", patchResp.Status)
			http.Error(w, "Менеджер вернул ошибку", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write(xmlData)
	}
}
