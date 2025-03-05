package handlers

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"time"

	"CrackHash/manager/internal/service"
	"CrackHash/manager/internal/store"
	"CrackHash/manager/internal/types"
)

type ManagerService interface {
	CreateTask(ctx context.Context, hash string, maxLength int) (string, error)
}

func CrackHandler(ctx context.Context, svc ManagerService) http.HandlerFunc {
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
		var req types.CrackRequest
		err = json.Unmarshal(body, &req)
		if err != nil {
			http.Error(w, "Некорректный JSON", http.StatusBadRequest)
			return
		}
		requestID, err := svc.CreateTask(ctx, req.Hash, req.MaxLength)
		if err != nil {
			http.Error(w, "Ошибка создания задачи", http.StatusInternalServerError)
			return
		}
		resp := types.RequestResponse{RequestID: requestID}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func StatusHandler(ctx context.Context, store store.RequestStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.URL.Query().Get("requestId")
		if requestID == "" {
			http.Error(w, "requestId не задан", http.StatusBadRequest)
			return
		}
		state, ok := store.Get(requestID)
		if !ok {
			http.Error(w, "Запрос не найден", http.StatusNotFound)
			return
		}
		progress := 0
		if state.Status == "IN_PROGRESS" && !state.StartTime.IsZero() && state.Timeout > 0 {
			elapsed := time.Since(state.StartTime)
			if elapsed >= state.Timeout {
				progress = 100
			} else {
				progress = int(float64(elapsed) / float64(state.Timeout) * 100)
				if progress < 1 && elapsed > 0 {
					progress = 1
				}
			}
		} else if state.Status == "READY" || state.Status == "ERROR" {
			progress = 100
		}
		resp := types.StatusResponse{
			Status:   state.Status,
			Data:     state.Data,
			Progress: progress,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func WorkerResponseHandler(ctx context.Context, store store.RequestStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
			return
		}
		var workerResp types.CrackHashWorkerResponse
		err = xml.Unmarshal(body, &workerResp)
		if err != nil {
			http.Error(w, "Некорректный XML", http.StatusBadRequest)
			return
		}
		state, ok := store.Get(workerResp.RequestId)
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
			store.Update(workerResp.RequestId, state)
		}
		w.WriteHeader(http.StatusOK)
	}
}
