package main

import (
	"log"

	"github.com/IBM/sarama"
)

func main() {
	config := sarama.NewConfig()
	admin, err := sarama.NewClusterAdmin([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal(err)
	}
	defer admin.Close()

	topics := map[string]sarama.TopicDetail{
		// test topics only
		"test_topic": {
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		// topics for Order Example
		"OrderEventsTopic": {
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		"InventoryEventsTopic": {
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		"ShippingEventsTopic": {
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		"PaymentEventsTopic": {
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	for topicName, topicDetail := range topics {
		err = createIfNotExist(admin, topicName, topicDetail)
		if err != nil {
			log.Printf("Error creating topic %s: %v\n", topicName, err)
		} else {
			log.Printf("Topic %s created successfully or already exists\n", topicName)
		}
	}
}

func createIfNotExist(admin sarama.ClusterAdmin, topicName string, topicDetail sarama.TopicDetail) error {
	// Check if the topic already exists
	topics, err := admin.ListTopics()
	if err != nil {
		return err
	}

	if _, exists := topics[topicName]; exists {
		log.Printf("Topic %s already exists\n", topicName)
		return nil
	}

	// Create the topic if it does not exist
	err = admin.CreateTopic(topicName, &topicDetail, false)
	if err != nil {
		return err
	}

	return nil
}
