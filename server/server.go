package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
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

func initDB() *sql.DB {
	db, err := sql.Open("sqlite", "./quotes.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS quotes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        value TEXT NOT NULL
    );`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	return db
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

func logExchangeRate(db *sql.DB, ctx context.Context, bid string) error {
	_, err := db.ExecContext(ctx, "INSERT INTO quotes (value) VALUES (?)", bid)
	return err
}

func quoteHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		http.NotFound(w, r)
		return
	}

	fetchCtx, fetchCancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer fetchCancel()

	bid, err := fetchExchangeRate(fetchCtx)
	if err != nil {
		log.Println("Timeout or error fetching exchange rate:", err)
		http.Error(w, "Failed to fetch exchange rate: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	logCtx, logCancel := context.WithTimeout(r.Context(), 10*time.Millisecond)
	defer logCancel()

	if err := logExchangeRate(db, logCtx, bid); err != nil {
		log.Println("Timeout or error logging exchange rate:", err)
		http.Error(w, "Failed to log exchange rate: "+err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]string{"bid": bid})
}

func main() {
	db := initDB()
	defer db.Close()

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		quoteHandler(db, w, r)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
