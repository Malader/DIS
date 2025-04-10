package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Скрипт для шифрования символов
func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Введите строку: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Ошибка чтения строки: %v\n", err)
		return
	}
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println("Пустая строка, завершение.")
		return
	}

	hashBytes := md5.Sum([]byte(input))
	hashStr := hex.EncodeToString(hashBytes[:])

	fmt.Print("Введите maxLength (оставьте пустым для использования длины строки): ")
	maxLengthStr, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Ошибка чтения maxLength: %v\n", err)
		return
	}
	maxLengthStr = strings.TrimSpace(maxLengthStr)

	var maxLength int
	if maxLengthStr == "" {
		maxLength = len(input)
	} else {
		maxLength, err = strconv.Atoi(maxLengthStr)
		if err != nil {
			fmt.Printf("Ошибка преобразования maxLength: %v. Будет использована длина строки.\n", err)
			maxLength = len(input)
		}
	}

	curlCmd := fmt.Sprintf(`curl -X POST -H "Content-Type: application/json" -d "{\"hash\":\"%s\", \"maxLength\":%d}" http://localhost:8080/api/hash/crack`, hashStr, maxLength)

	fmt.Println("\nСформированная команда:")
	fmt.Println(curlCmd)
}
