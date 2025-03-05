package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"CrackHash/manager/internal/types"
)

//go:generate mockgen -destination=./mocks/mock_worker_client.go -package=mocks CrackHash/manager/internal/service WorkerClient

type multiWorkerClient struct {
	workerURLs []string
	client     http.Client
}

func NewMultiWorkerClient(urls []string) WorkerClient {
	return multiWorkerClient{
		workerURLs: urls,
		client:     http.Client{Timeout: 2 * time.Minute},
	}
}

func (m multiWorkerClient) SendTask(ctx context.Context, task types.CrackHashManagerRequest) error {
	var succeeded bool
	var errs []string
	total := len(m.workerURLs)
	for i, url := range m.workerURLs {
		task.PartNumber = i
		task.PartCount = total

		jsonData, err := json.Marshal(task)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := m.client.Do(req)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			succeeded = true
		} else {
			errs = append(errs, "HTTP error: "+http.StatusText(resp.StatusCode))
		}
	}
	if !succeeded {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
