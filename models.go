package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type Account struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

type Transaction struct {
	ID              int       `json:"id"`
	Value           float64   `json:"value"`
	AccountID       int       `json:"account_id"`
	GroupType       string    `json:"group_type"`
	Account2ID      *int      `json:"account2_id,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
}

var db *sql.DB

func initDB(dataSourceName string) {
	var err error
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Could not ping database: %v", err)
	}

	fmt.Println("Connected to database!")
}
