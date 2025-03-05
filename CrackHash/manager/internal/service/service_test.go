package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"CrackHash/manager/internal/service"
	"CrackHash/manager/internal/store"

	"CrackHash/manager/internal/service/mocks"
)

func TestManagerService_CreateTask_Success(t *testing.T) {
	type fields struct {
		prepareWorkerClient func(*mocks.MockWorkerClient)
		responseTimeout     time.Duration
		store               store.RequestStore
	}
	type args struct {
		hash      string
		maxLength int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Success: CreateTask returns non-empty ID",
			fields: fields{
				responseTimeout: 5 * time.Second,
				store:           store.NewRequestStore(),
				prepareWorkerClient: func(mwc *mocks.MockWorkerClient) {
					mwc.EXPECT().
						SendTask(gomock.Any(), gomock.Any()).
						Return(nil)
				},
			},
			args: args{
				hash:      "testhash",
				maxLength: 4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWorkerClient := mocks.NewMockWorkerClient(ctrl)
			tt.fields.prepareWorkerClient(mockWorkerClient)

			svc := service.NewManagerService(tt.fields.store, mockWorkerClient, tt.fields.responseTimeout)
			id, err := svc.CreateTask(context.Background(), tt.args.hash, tt.args.maxLength)
			require.NoError(t, err)
			require.NotEmpty(t, id)
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestManagerService_CreateTask_Failure(t *testing.T) {
	type fields struct {
		prepareWorkerClient func(*mocks.MockWorkerClient)
		responseTimeout     time.Duration
		store               store.RequestStore
	}
	type args struct {
		hash      string
		maxLength int
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		expectedErr string
	}{
		{
			name: "Failure: SendTask returns error",
			fields: fields{
				responseTimeout: 5 * time.Second,
				store:           store.NewRequestStore(),
				prepareWorkerClient: func(mwc *mocks.MockWorkerClient) {
					mwc.EXPECT().
						SendTask(gomock.Any(), gomock.Any()).
						Return(errors.New("send error"))
				},
			},
			args: args{
				hash:      "testhash",
				maxLength: 4,
			},
			expectedErr: "send error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWorkerClient := mocks.NewMockWorkerClient(ctrl)
			tt.fields.prepareWorkerClient(mockWorkerClient)

			svc := service.NewManagerService(tt.fields.store, mockWorkerClient, tt.fields.responseTimeout)
			id, err := svc.CreateTask(context.Background(), tt.args.hash, tt.args.maxLength)
			require.NoError(t, err)
			require.NotEmpty(t, id)
			time.Sleep(100 * time.Millisecond)
		})
	}
}
