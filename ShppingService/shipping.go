package main

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

type InventoryEvent struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
}

type ShippingEvent struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
}

func main() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("InventoryEventsTopic", 0, sarama.OffsetOldest)
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
		log.Printf("Received inventory event: %s", string(msg.Value))

		var inventoryEvent InventoryEvent
		err := json.Unmarshal(msg.Value, &inventoryEvent)
		if err != nil {
			log.Printf("Error unmarshalling inventory event: %v", err)
			continue
		}

		if inventoryEvent.Status == "OK" {
			shippingEvent := ShippingEvent{
				OrderID: inventoryEvent.OrderID,
				Status:  "Shipped",
			}

			shippingEventBytes, err := json.Marshal(shippingEvent)
			if err != nil {
				log.Printf("Error marshalling shipping event: %v", err)
				continue
			}

			msg := &sarama.ProducerMessage{
				Topic: "ShippingEventsTopic",
				Value: sarama.StringEncoder(shippingEventBytes),
			}

			_, _, err = producer.SendMessage(msg)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Shipping event published")
		} else {
			log.Printf("Inventory status for order %s is not OK, skipping shipping", inventoryEvent.OrderID)
		}
	}
}
