package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Erro requesting quote:", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal("Error decoding response:", err)
	}

	bid := result["bid"]
	err = os.WriteFile("cotacao.txt", []byte("Dólar: "+bid), 0644)
	if err != nil {
		log.Fatal("Error writing file", err)
	}
}
