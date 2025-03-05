package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"CrackHash/manager/internal/handlers"
	"CrackHash/manager/internal/service"
	"CrackHash/manager/internal/store"
	"CrackHash/manager/internal/types"
	"github.com/stretchr/testify/require"
)

func TestManagerIntegration_Success(t *testing.T) {
	reqStore := store.NewRequestStore()
	stubWorker := stubWorkerClient{}
	mgrSvc := service.NewManagerService(reqStore, stubWorker, 5*time.Second)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/hash/crack", handlers.CrackHandler(context.Background(), mgrSvc))
	mux.HandleFunc("/api/hash/status", handlers.StatusHandler(context.Background(), reqStore))
	mux.HandleFunc("/internal/api/manager/hash/crack/request", handlers.WorkerResponseHandler(context.Background(), reqStore))

	server := httptest.NewServer(mux)
	defer server.Close()

	reqBody := types.CrackRequest{
		Hash:      "dummy",
		MaxLength: 4,
	}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	postResp, err := http.Post(server.URL+"/api/hash/crack", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, postResp.StatusCode)
	defer postResp.Body.Close()

	var createResp types.RequestResponse
	err = json.NewDecoder(postResp.Body).Decode(&createResp)
	require.NoError(t, err)
	require.NotEmpty(t, createResp.RequestID)

	xmlResp := types.CrackHashWorkerResponse{
		RequestId:  createResp.RequestID,
		PartNumber: 0,
	}
	xmlResp.Answers.Words = []string{"dummy"}
	xmlBody, err := xml.MarshalIndent(xmlResp, "", "  ")
	require.NoError(t, err)
	xmlBody = append([]byte(xml.Header), xmlBody...)

	patchReq, err := http.NewRequest(http.MethodPatch, server.URL+"/internal/api/manager/hash/crack/request", bytes.NewReader(xmlBody))
	require.NoError(t, err)
	patchReq.Header.Set("Content-Type", "application/xml")

	patchResp, err := http.DefaultClient.Do(patchReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, patchResp.StatusCode)
	defer patchResp.Body.Close()
	
	getResp, err := http.Get(server.URL + "/api/hash/status?requestId=" + createResp.RequestID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, getResp.StatusCode)
	defer getResp.Body.Close()

	var statusResp types.StatusResponse
	err = json.NewDecoder(getResp.Body).Decode(&statusResp)
	require.NoError(t, err)
	require.Equal(t, "READY", statusResp.Status)
	require.Equal(t, []string{"dummy"}, statusResp.Data)
}

type stubWorkerClient struct{}

func (s stubWorkerClient) SendTask(ctx context.Context, task types.CrackHashManagerRequest) error {
	return nil
}
