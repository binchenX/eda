package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/linkedin/goavro/v2"
)

type Order struct {
	OrderID string   `json:"orderId"`
	UserID  string   `json:"userId"`
	Items   []string `json:"items"`
}

var producer sarama.SyncProducer
var avroCodec *goavro.Codec

func initialize() {
	var err error
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err = sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal(err)
	}

	InitDb()

	// Fetch schema from schema registry
	schema, err := fetchSchemaFromRegistry("http://localhost:8081/subjects/OrderEventTopic-value/versions/latest")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Avro codec
	avroCodec, err = goavro.NewCodec(schema)
	if err != nil {
		log.Fatal(err)
	}
}

func fetchSchemaFromRegistry(url string) (string, error) {
	maxRetries := 10
	retryInterval := time.Second * 5
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("failed to connect to Schema Registry: %v", err)
			log.Printf("Attempt %d/%d: %v. Retrying in %v...", i+1, maxRetries, lastErr, retryInterval)
			time.Sleep(retryInterval)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("failed to read response body: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				return "", fmt.Errorf("failed to parse response JSON: %v", err)
			}

			schema, ok := result["schema"].(string)
			if !ok {
				return "", fmt.Errorf("invalid schema format in response")
			}

			log.Printf("Successfully fetched schema from Schema Registry")
			return schema, nil
		}

		lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		log.Printf("Attempt %d/%d: %v. Retrying in %v...", i+1, maxRetries, lastErr, retryInterval)
		time.Sleep(retryInterval)
	}

	return "", fmt.Errorf("failed to fetch schema after %d attempts: %v", maxRetries, lastErr)
}

func sendOrderCreateEvent(order Order) error {
	// Convert Order struct to Avro binary
	orderMap := map[string]interface{}{
		"orderId": order.OrderID,
		"userId":  order.UserID,
		"items":   order.Items,
	}
	binaryOrder, err := avroCodec.BinaryFromNative(nil, orderMap)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "OrderEventsTopic",
		Value: sarama.ByteEncoder(binaryOrder),
	}

	// dump the Value to see the Avro binary
	log.Printf("Avro binary: %v\n", msg.Value)

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

	err = SaveOrder(order)
	if err != nil {
		http.Error(w, "Error saving order to database", http.StatusInternalServerError)
		return
	}

	err = sendOrderCreateEvent(order)
	if err != nil {
		http.Error(w, "Error sending order create event", http.StatusInternalServerError)
		return
	}

	orderStatusMap.Store(order.OrderID, "Order Created")

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
