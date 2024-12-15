package main

import (
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDb() {
	var err error
	connStr := "postgres://postgres:postgres@localhost:5432/orders?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Create orders table if it does not exist
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS orders (
        id SERIAL PRIMARY KEY,
        order_id VARCHAR(255) NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        items JSONB NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    `
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func SaveOrder(order Order) error {
	orderItems, err := json.Marshal(order.Items)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO orders (order_id, user_id, items) VALUES ($1, $2, $3)", order.OrderID, order.UserID, orderItems)
	return err
}
