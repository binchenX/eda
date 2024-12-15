package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

type Order struct {
	OrderID string   `json:"orderId"`
	UserID  string   `json:"userId"`
	Items   []string `json:"items"`
}

var producer sarama.SyncProducer

func initialize() {
	var err error
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err = sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal(err)
	}

	InitDb()
}

func sendOrderCreateEvent(order Order) error {
	orderEvent, err := json.Marshal(order)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "OrderEventsTopic",
		Value: sarama.StringEncoder(orderEvent),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Order event is stored in topic(%s)/partition(%d)/offset(%d)\n", "OrderEventsTopic", partition, offset)
	return nil
}

func orderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order.OrderID = uuid.New().String()

	// save order to database
	err = SaveOrder(order)
	if err != nil {
		http.Error(w, "Error saving order to database", http.StatusInternalServerError)
		return
	}

	// todo: save order status to database
	orderStatusMap.Store(order.OrderID, "Order Created")

	// send order create event
	err = sendOrderCreateEvent(order)
	if err != nil {
		http.Error(w, "Error sending order create event", http.StatusInternalServerError)
		return
	}

	// response with order ID
	response := map[string]string{
		"orderId": order.OrderID,
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(responseBytes)
}
