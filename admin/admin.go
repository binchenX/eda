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

	topicName := "test_topic"
	topicDetail := sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	err = createIfNotExist(admin, topicName, topicDetail)
	if err != nil {
		log.Printf("Error creating topic: %v\n", err)
	} else {
		log.Printf("Topic %s created successfully or already exists\n", topicName)
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
