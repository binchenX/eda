package main

import (
	"log"

	"github.com/IBM/sarama"
)

func main() {
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("test_topic", 0, sarama.OffsetOldest)
	if err != nil {
		log.Fatal(err)
	}
	defer partitionConsumer.Close()

	for msg := range partitionConsumer.Messages() {
		log.Printf("Received message: %s", string(msg.Value))
	}
}
