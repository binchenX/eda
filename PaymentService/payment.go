package main

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

type ShippingEvent struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
}

type PaymentEvent struct {
	PaymentID string `json:"paymentId"`
	OrderID   string `json:"orderId"`
	Status    string `json:"status"`
}

func main() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("ShippingEventsTopic", 0, sarama.OffsetOldest)
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
		log.Printf("Received shipping event: %s", string(msg.Value))

		var shippingEvent ShippingEvent
		err := json.Unmarshal(msg.Value, &shippingEvent)
		if err != nil {
			log.Printf("Error unmarshalling shipping event: %v", err)
			continue
		}

		if shippingEvent.Status == "Shipped" {
			// Generate UUID for paymentId
			paymentID := uuid.New().String()

			// Start payment transaction
			paymentEvent := PaymentEvent{
				PaymentID: paymentID,
				OrderID:   shippingEvent.OrderID,
				Status:    "Paid",
			}

			paymentEventBytes, err := json.Marshal(paymentEvent)
			if err != nil {
				log.Printf("Error marshalling payment event: %v", err)
				continue
			}

			msg := &sarama.ProducerMessage{
				Topic: "PaymentEventsTopic",
				Value: sarama.StringEncoder(paymentEventBytes),
			}

			_, _, err = producer.SendMessage(msg)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Payment event published")
		} else {
			log.Printf("Shipping status for order %s is not Shipped, skipping payment", shippingEvent.OrderID)
		}
	}
}
