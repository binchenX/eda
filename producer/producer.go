package main

import (
	"log"

	"github.com/IBM/sarama"
)

func main() {
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("Hello World!"),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", "test_topic", partition, offset)
}
