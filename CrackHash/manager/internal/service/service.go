package service

import (
	"context"
	"errors"
	"time"

	"CrackHash/manager/internal/store"
	"CrackHash/manager/internal/types"

	"github.com/google/uuid"
)

const (
	StatusInProgress = "IN_PROGRESS" // добавить процент выполнения
	StatusReady      = "READY"
	StatusError      = "ERROR"
	MaxQueueSize     = 20
)

type WorkerClient interface {
	SendTask(ctx context.Context, task types.CrackHashManagerRequest) error
}

type ManagerServiceImpl struct {
	store           store.RequestStore
	workerClient    WorkerClient
	responseTimeout time.Duration
}

func NewManagerService(s store.RequestStore, wc WorkerClient, timeout time.Duration) ManagerServiceImpl {
	return ManagerServiceImpl{
		store:           s,
		workerClient:    wc,
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

	state := store.RequestState{
		Status:    StatusInProgress,
		Data:      nil,
		StartTime: time.Now(),
		Timeout:   m.responseTimeout,
		Timer: time.AfterFunc(m.responseTimeout, func() {
			s, ok := m.store.Get(requestID)
			if ok && s.Status == StatusInProgress {
				s.Status = StatusError
				m.store.Update(requestID, s)
			}
		}),
	}
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

	go func() {
		if err := m.workerClient.SendTask(ctx, task); err != nil {
			m.store.Update(requestID, store.RequestState{
				Status: StatusError,
			})
		}
	}()

	return requestID, nil
}
