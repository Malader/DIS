package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"CrackHash/worker/internal/handlers"
	"CrackHash/worker/internal/service"
	"CrackHash/worker/internal/types"
)

func TestWorkerIntegration_Success(t *testing.T) {
	managerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer managerServer.Close()

	os.Setenv("MANAGER_URL", managerServer.URL)
	defer os.Unsetenv("MANAGER_URL")

	workerSvc := service.NewWorkerService()

	mux := http.NewServeMux()
	mux.HandleFunc("/internal/api/worker/hash/crack/task", handlers.TaskHandler(context.Background(), workerSvc))
	server := httptest.NewServer(mux)
	defer server.Close()

	reqPayload := types.CrackHashManagerRequest{
		RequestId:  "test-request",
		PartNumber: 0,
		PartCount:  1,
		Hash:       "0cc175b9c0f1b6a831c399e269772661", // MD5("a")
		MaxLength:  1,
		Alphabet: types.Alphabet{
			Symbols: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
		},
	}
	jsonData, err := json.Marshal(reqPayload)
	require.NoError(t, err)

	postResp, err := http.Post(server.URL+"/internal/api/worker/hash/crack/task", "application/json", bytes.NewReader(jsonData))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, postResp.StatusCode)
	defer postResp.Body.Close()

	var xmlResp types.CrackHashWorkerResponse
	err = xml.NewDecoder(postResp.Body).Decode(&xmlResp)
	require.NoError(t, err)

	require.Equal(t, "test-request", xmlResp.RequestId)
	require.Equal(t, 0, xmlResp.PartNumber)
	require.Contains(t, xmlResp.Answers.Words, "a")
}
