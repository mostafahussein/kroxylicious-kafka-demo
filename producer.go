package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
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

	faker := gofakeit.New(uint64(time.Now().UnixNano()))

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	})
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	for i := 0; i < 10; i++ {
		cc := faker.CreditCard()
		transaction := Transaction{
			CardNumber:     cc.Number,
			CardholderName: faker.Name(),
			Amount:         faker.Price(10, 100),
			Currency:       "USD",
			Location:       faker.City(),
			Timestamp:      time.Now().Format("2006-01-02 15:04:05"),
		}

		payload, _ := json.Marshal(transaction)

		err := producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(transaction.CardNumber),
			Value: payload,
		}, nil)

		if err != nil {
			fmt.Println("Produce error:", err)
			continue
		}
	}

	producer.Flush(5000)
	fmt.Println("[PRODUCER] Finished sending messages")
}

