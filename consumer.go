package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/joho/godotenv"
)

var topic = "credit_card_transactions"

type Transaction struct {
	CardNumber     string  `json:"card_number"`
	CardholderName string  `json:"cardholder_name"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Location       string  `json:"location"`
	Timestamp      string  `json:"timestamp"`
}

func GetKafkaBootstrapServers() string {
	_ = godotenv.Load()

	enableGateway := strings.ToLower(os.Getenv("ENABLE_KAFKA_GATEWAY")) == "true"
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaGatewayHost := os.Getenv("KAFKA_GATEWAY_HOST")

	if enableGateway {
		if kafkaGatewayHost == "" {
			fmt.Println("Error: ENABLE_KAFKA_GATEWAY is true but KAFKA_GATEWAY_HOST is not set")
			os.Exit(1)
		}
		return kafkaGatewayHost
	} else {
		if kafkaHost == "" {
			fmt.Println("Error: KAFKA_HOST is not set in environment")
			os.Exit(1)
		}
		return kafkaHost
	}
}

func main() {
	bootstrapServers := GetKafkaBootstrapServers()

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          "go-cc-consumer",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("[CONSUMER] Listening...")

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

runLoop:
	for {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: exiting\n", sig)
			break runLoop
		default:
			msg, err := consumer.ReadMessage(100 * time.Millisecond)
			if err == nil {
				var tx Transaction
				if err := json.Unmarshal(msg.Value, &tx); err != nil {
					fmt.Printf("[ERROR] Failed to parse JSON: %v\n", err)
				} else {
					fmt.Printf("[CONSUMER] Received: %+v\n", tx)
				}
			} else if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() != kafka.ErrTimedOut {
				fmt.Printf("[ERROR] Kafka error: %v\n", kafkaErr)
			}
		}
	}
}

