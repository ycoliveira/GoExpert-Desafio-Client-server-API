package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRate struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

func fetchExchangeRate(ctx context.Context) (string, error) {
	var result struct {
		USDBRL ExchangeRate `json:"USDBRL"`
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil
	}

	return result.USDBRL.Bid, nil
}

func logExchangeRate(ctx context.Context, bid string) error {
	db, err := sql.Open("sqlite3", "./quotes.db")
	if err != nil {
		return err
	}

	defer db.Close()

	_, err = db.ExecContext(ctx, "INSERT INTO quotes (value) VALUES (?)", bid)
	return err
}

func quoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		http.NotFound(w, r)
		return
	}

	fetchCtx, fetchCancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer fetchCancel()

	bid, err := fetchExchangeRate(fetchCtx)
	if err != nil {
		http.Error(w, "failed to fetch exchange rate", http.StatusServiceUnavailable)
		return
	}

	logCtx, logCancel := context.WithTimeout(r.Context(), 10*time.Millisecond)
	defer logCancel()

	if err := logExchangeRate(logCtx, bid); err != nil {
		log.Println("failed to log exchange rate:", err)
	}

	json.NewEncoder(w).Encode(map[string]string{"bid": bid})
}

func main() {

	http.HandleFunc("/cotacao", quoteHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

	// req, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	// defer req.Body.Close()
	// if err != nil {
	// 	panic(err)
	// }

	// res, err := io.ReadAll(req.Body)
	// if err != nil {
	// 	panic(err)
	// }

	// var data Cambio
	// err = json.Unmarshal(res, &data)
	// if err != nil {
	// 	fmt.Println("erro")
	// }

	// fmt.Println(data)
}
