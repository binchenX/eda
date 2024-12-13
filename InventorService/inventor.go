package main

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

type Order struct {
	OrderID string   `json:"orderId"`
	UserID  string   `json:"userId"`
	Items   []string `json:"items"`
}

type InventoryEvent struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
}

var inventory = map[string]int{
	"item1": 10,
	"item2": 5,
	"item3": 0,
}

func main() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("OrderEventsTopic", 0, sarama.OffsetOldest)
	if err != nil {
		log.Fatal(err)
	}
	defer partitionConsumer.Close()

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	for msg := range partitionConsumer.Messages() {
		log.Printf("Received order event: %s", string(msg.Value))

		var order Order
		err := json.Unmarshal(msg.Value, &order)
		if err != nil {
			log.Printf("Error unmarshalling order event: %v", err)
			continue
		}

		status := "OK"
		for _, item := range order.Items {
			if inventory[item] <= 0 {
				status = "FAILED"
				break
			}
		}

		inventoryEvent := InventoryEvent{
			OrderID: order.OrderID,
			Status:  status,
		}

		inventoryEventBytes, err := json.Marshal(inventoryEvent)
		if err != nil {
			log.Printf("Error marshalling inventory event: %v", err)
			continue
		}

		msg := &sarama.ProducerMessage{
			Topic: "InventoryEventsTopic",
			Value: sarama.StringEncoder(inventoryEventBytes),
		}

		_, _, err = producer.SendMessage(msg)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Inventory event published")
	}
}
