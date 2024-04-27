package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "./server/quotes.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, value FROM quotes")
	if err != nil {
		log.Fatal("Failed to execute query:", err)
	}
	defer rows.Close()

	fmt.Println("ID | Value")
	var id int
	var value string
	for rows.Next() {
		err := rows.Scan(&id, &value)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}
		fmt.Printf("%d | %s\n", id, value)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal("Error fetching rows:", err)
	}
}
