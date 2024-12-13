package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/IBM/sarama"
)

type OrderStatus struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
}

type InventoryEvent struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
}

type ShippingEvent struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
}

type PaymentEvent struct {
	PaymentID string `json:"paymentId"`
	OrderID   string `json:"orderId"`
	Status    string `json:"status"`
}

var orderStatusMap sync.Map

func orderStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	orderID := r.URL.Query().Get("orderId")
	if orderID == "" {
		http.Error(w, "Missing orderId parameter", http.StatusBadRequest)
		return
	}

	status, ok := orderStatusMap.Load(orderID)
	if !ok {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	response := OrderStatus{
		OrderID: orderID,
		Status:  status.(string),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBytes)
}

func consumeEvents() {
	config := sarama.NewConfig()
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	topics := []string{"OrderEventsTopic", "InventoryEventsTopic", "ShippingEventsTopic", "PaymentEventsTopic"}
	for _, topic := range topics {
		go consumeTopic(consumer, topic)
	}

	// Keep the main function running
	select {}
}

func consumeTopic(consumer sarama.Consumer, topic string) {
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Fatal(err)
	}
	defer partitionConsumer.Close()

	for msg := range partitionConsumer.Messages() {
		log.Printf("Received message from topic %s: %s", topic, string(msg.Value))
		updateOrderStatus(topic, string(msg.Value))
	}
}

func updateOrderStatus(topic, message string) {
	var orderID, status string
	switch topic {
	case "OrderEventsTopic":
		var order Order
		json.Unmarshal([]byte(message), &order)
		orderID = order.OrderID
		status = "Order Created"
	case "InventoryEventsTopic":
		var inventoryEvent InventoryEvent
		json.Unmarshal([]byte(message), &inventoryEvent)
		orderID = inventoryEvent.OrderID
		if inventoryEvent.Status == "OK" {
			status = "Inventory Confirmed"
		} else {
			status = "Inventory Failed"
		}
	case "ShippingEventsTopic":
		var shippingEvent ShippingEvent
		json.Unmarshal([]byte(message), &shippingEvent)
		orderID = shippingEvent.OrderID
		status = "Shipped"
	case "PaymentEventsTopic":
		var paymentEvent PaymentEvent
		json.Unmarshal([]byte(message), &paymentEvent)
		orderID = paymentEvent.OrderID
		status = "Paid"
	}

	orderStatusMap.Store(orderID, status)
	log.Printf("Order %s status updated to %s", orderID, status)
}
