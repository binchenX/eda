package main

import (
	"log"
	"net/http"
)

func main() {
	initialize()

	go consumeEvents()

	http.HandleFunc("/order", orderHandler)
	http.HandleFunc("/order/status", orderStatusHandler)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
