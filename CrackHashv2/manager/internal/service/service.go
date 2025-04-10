package service

import (
	"context"
	"errors"
	"log"
	"time"

	"CrackHash/manager/internal/queue"
	"CrackHash/manager/internal/store"
	"CrackHash/manager/internal/types"

	"github.com/google/uuid"
)

const (
	StatusInProgress = "IN_PROGRESS"
	StatusReady      = "READY"
	StatusError      = "ERROR"
	MaxQueueSize     = 100
)

type ManagerService interface {
	CreateTask(ctx context.Context, hash string, maxLength int) (string, error)
	RetryPendingTasks(ctx context.Context) error
}

type ManagerServiceImpl struct {
	store           store.RequestStore
	rabbitClient    queue.TaskQueue
	responseTimeout time.Duration
}

func NewManagerService(
	s store.RequestStore,
	qc queue.TaskQueue,
	timeout time.Duration,
) ManagerServiceImpl {
	return ManagerServiceImpl{
		store:           s,
		rabbitClient:    qc,
		responseTimeout: timeout,
	}
}

func (m ManagerServiceImpl) CreateTask(
	ctx context.Context,
	hash string,
	maxLength int,
) (string, error) {

	if m.store.Count() >= MaxQueueSize {
		return "", errors.New("очередь заполнена, попробуйте позже")
	}

	requestID := uuid.New().String()
	log.Printf("[managerService] Создаём задачу requestID=%s, hash=%s, maxLength=%d",
		requestID, hash, maxLength)

	state := store.RequestState{
		Status:    StatusInProgress,
		Data:      nil,
		StartTime: time.Now(),
		Timeout:   m.responseTimeout,
	}
	state.Timer = time.AfterFunc(m.responseTimeout, func() {
		s, ok := m.store.Get(requestID)
		if ok && s.Status == StatusInProgress {
			s.Status = StatusError
			m.store.Update(requestID, s)
			log.Printf("[managerService] Задача requestID=%s переведена в ERROR (таймаут)", requestID)
			m.store.MarkPending(requestID, false)
		}
	})

	m.store.Set(requestID, state)

	var alphabet []string
	for _, ch := range "abcdefghijklmnopqrstuvwxyz0123456789" {
		alphabet = append(alphabet, string(ch))
	}

	task := types.CrackHashManagerRequest{
		RequestId:  requestID,
		PartNumber: 0,
		PartCount:  1,
		Hash:       hash,
		MaxLength:  maxLength,
		Alphabet: types.Alphabet{
			Symbols: alphabet,
		},
	}

	go func(reqID string, t types.CrackHashManagerRequest) {
		if m.rabbitClient != nil && m.rabbitClient.IsConnected() {
			log.Printf("[managerService] Публикуем задачу requestID=%s в RabbitMQ (hash=%s, maxLength=%d)",
				reqID, t.Hash, t.MaxLength)
			err := m.rabbitClient.PublishTask(t)
			if err != nil {
				log.Printf("[managerService] Ошибка PublishTask для requestID=%s: %v", reqID, err)
				m.store.MarkPending(reqID, true)
			} else {
				m.store.MarkPending(reqID, false)
				log.Printf("[managerService] Успешно отправили requestID=%s в очередь", reqID)
			}
		} else {
			log.Printf("[managerService] RabbitMQ не подключен, помечаем requestID=%s как pending", reqID)
			m.store.MarkPending(reqID, true)
		}
	}(requestID, task)

	return requestID, nil
}

func (m ManagerServiceImpl) RetryPendingTasks(ctx context.Context) error {
	pendingList := m.store.GetPending()
	if len(pendingList) > 0 {
		log.Printf("[managerService] Найдено %d pending-задач, пробуем переотправить...", len(pendingList))
	}
	for _, req := range pendingList {
		if req.Status == StatusInProgress {
			fullAlphabet := []string{
				"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
				"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
				"u", "v", "w", "x", "y", "z",
				"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
			}
			task := types.CrackHashManagerRequest{
				RequestId:  req.ID,
				PartNumber: 0,
				PartCount:  1,
				Hash:       req.Hash,
				MaxLength:  req.MaxLength,
				Alphabet: types.Alphabet{
					Symbols: fullAlphabet,
				},
			}
			log.Printf("[managerService] Переотправляем pending-задачу requestID=%s (hash=%s, maxLength=%d)",
				req.ID, req.Hash, req.MaxLength)
			if err := m.rabbitClient.PublishTask(task); err != nil {
				log.Printf("[managerService] Ошибка при повторной отправке %s: %v", req.ID, err)
			} else {
				m.store.MarkPending(req.ID, false)
				log.Printf("[managerService] Успешно переотправили requestID=%s", req.ID)
			}
		}
	}
	return nil
}
