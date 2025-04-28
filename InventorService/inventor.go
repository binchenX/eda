package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/IBM/sarama"
	"github.com/linkedin/goavro/v2"
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

var OrderEventAvroCodec *goavro.Codec

func init() {
	var err error

	// Fetch schema from schema registry by subject and version
	schema, err := fetchSchemaFromRegistry("http://localhost:8081/subjects/OrderEventTopic-value/versions/latest")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Avro codec for Order events
	OrderEventAvroCodec, err = goavro.NewCodec(schema)
	if err != nil {
		log.Fatal(err)
	}
}

func fetchSchemaFromRegistry(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch schema: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	schema, ok := result["schema"].(string)
	if !ok {
		return "", fmt.Errorf("invalid schema format")
	}

	return schema, nil
}

func main() {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("OrderEventsTopic", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatal(err)
	}
	defer partitionConsumer.Close()

	for msg := range partitionConsumer.Messages() {
		native, _, err := OrderEventAvroCodec.NativeFromBinary(msg.Value)
		if err != nil {
			log.Printf("Error decoding Avro message: %v", err)
			continue
		}

		orderMap := native.(map[string]interface{})
		order := Order{
			OrderID: orderMap["orderId"].(string),
			UserID:  orderMap["userId"].(string),
			Items:   convertToStringSlice(orderMap["items"].([]interface{})),
		}

		log.Printf("Consumed order: %+v\n", order)
		// Process the order (e.g., update inventory)
		status := processOrder(order)

		// Send InventoryEvent
		sendInventoryEvent(order.OrderID, status)
	}
}

func convertToStringSlice(interfaces []interface{}) []string {
	strings := make([]string, len(interfaces))
	for i, v := range interfaces {
		strings[i] = v.(string)
	}
	return strings
}

func processOrder(order Order) string {
	// if any of the items are out of stock, return "failed"
	for _, item := range order.Items {
		if inventory[item] > 0 {
			inventory[item]--
			log.Printf("Item %s inventory decremented. New inventory: %d\n", item, inventory[item])
		} else {
			log.Printf("Item %s is out of stock!\n", item)
			return "FAILED"
		}
	}
	return "OK"
}

func sendInventoryEvent(orderID, status string) {
	event := InventoryEvent{
		OrderID: orderID,
		Status:  status,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshalling inventory event: %v", err)
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: "InventoryEventsTopic",
		Value: sarama.ByteEncoder(eventBytes),
	}

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Printf("Error creating Kafka producer: %v", err)
		return
	}
	defer producer.Close()

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Printf("Error sending inventory event: %v", err)
		return
	}

	log.Printf("Inventory event sent to topic(%s)/partition(%d)/offset(%d)\n", "InventoryEventsTopic", partition, offset)
}
