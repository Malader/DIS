package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"runtime"
	"strings"
	"sync"

	"CrackHash/worker/internal/handlers"
)

type workerServiceImpl struct{}

func NewWorkerService() handlers.WorkerService {
	return workerServiceImpl{}
}

func (w workerServiceImpl) ProcessTask(
	hash string,
	maxLength int,
	alphabet []string,
	partNumber, partCount int,
) []string {

	n := len(alphabet)
	total := 0
	for l := 1; l <= maxLength; l++ {
		total += int(math.Pow(float64(n), float64(l)))
	}
	startIndex := total * partNumber / partCount
	endIndex := total * (partNumber + 1) / partCount

	targetHash := strings.ToLower(hash)
	var results []string

	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}

	rangeSize := endIndex - startIndex
	segSize := rangeSize / numWorkers
	if segSize < 1 {
		segSize = 1
	}

	fmt.Printf("[workerService] Начинаем ProcessTask: hash=%s maxLength=%d partNumber=%d/%d total=%d startIndex=%d endIndex=%d\n",
		hash, maxLength, partNumber, partCount, total, startIndex, endIndex)

	var wg sync.WaitGroup
	resChan := make(chan string, 100)

	for i := 0; i < numWorkers; i++ {
		segStart := startIndex + i*segSize
		segEnd := segStart + segSize
		if i == numWorkers-1 {
			segEnd = endIndex
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			for idx := s; idx < e; idx++ {
				word := indexToWord(idx, maxLength, alphabet)
				if word == "" {
					continue
				}
				hashBytes := md5.Sum([]byte(word))
				hashStr := hex.EncodeToString(hashBytes[:])
				if hashStr == targetHash {
					resChan <- word
				}
			}
		}(segStart, segEnd)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	for w := range resChan {
		results = append(results, w)
	}

	fmt.Printf("[workerService] Завершили ProcessTask: hash=%s, найдено %d слов\n", hash, len(results))
	return results
}

func indexToWord(index int, maxLength int, alphabet []string) string {
	n := len(alphabet)
	sum := 0
	for l := 1; l <= maxLength; l++ {
		count := intPow(n, l)
		if index < sum+count {
			rank := index - sum
			word := ""
			for i := 0; i < l; i++ {
				power := intPow(n, l-i-1)
				pos := rank / power
				word += alphabet[pos]
				rank = rank % power
			}
			return word
		}
		sum += count
	}
	return ""
}

func intPow(a, b int) int {
	result := 1
	for i := 0; i < b; i++ {
		result *= a
	}
	return result
}
