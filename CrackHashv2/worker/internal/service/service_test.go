package service_test

import (
	"crypto/md5"
	"encoding/hex"
	"testing"

	"CrackHash/worker/internal/service"
	"github.com/stretchr/testify/require"
)

func TestWorkerService_ProcessTask_Success(t *testing.T) {
	type args struct {
		hash       string
		maxLength  int
		alphabet   []string
		partNumber int
		partCount  int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "MaxLength=1, find 'a'",
			args: args{
				hash:      "0cc175b9c0f1b6a831c399e269772661", // MD5("a")
				maxLength: 1,
				alphabet: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
					"k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
					"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
				partNumber: 0,
				partCount:  1,
			},
			want: []string{"a"},
		},
		{
			name: "MaxLength=2, find 'ab'",
			args: args{
				hash: func() string {
					sum := md5.Sum([]byte("ab"))
					return hex.EncodeToString(sum[:])
				}(),
				maxLength:  2,
				alphabet:   []string{"a", "b", "c"},
				partNumber: 0,
				partCount:  1,
			},
			want: []string{"ab"},
		},
	}

	workerSvc := service.NewWorkerService()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := workerSvc.ProcessTask(tt.args.hash, tt.args.maxLength, tt.args.alphabet, tt.args.partNumber, tt.args.partCount)
			if got == nil {
				got = []string{}
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestWorkerService_ProcessTask_NoMatch(t *testing.T) {
	type args struct {
		hash       string
		maxLength  int
		alphabet   []string
		partNumber int
		partCount  int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "MaxLength=1, no match",
			args: args{
				hash:      "ffffffffffffffffffffffffffffffff",
				maxLength: 1,
				alphabet: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
					"k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
					"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
				partNumber: 0,
				partCount:  1,
			},
			want: []string{},
		},
	}

	workerSvc := service.NewWorkerService()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := workerSvc.ProcessTask(tt.args.hash, tt.args.maxLength, tt.args.alphabet, tt.args.partNumber, tt.args.partCount)
			if got == nil {
				got = []string{}
			}
			require.Equal(t, tt.want, got)
		})
	}
}
