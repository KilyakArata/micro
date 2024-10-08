package main

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Error creating Kafka consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("user_registered", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Error starting partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	// Обработка сообщений
	go func() {
		for message := range partitionConsumer.Messages() {
			var user User
			err := json.Unmarshal(message.Value, &user)
			if err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				continue
			}

			// Отправляем уведомление (в данном случае просто выводим на экран)
			sendNotification(user)
		}
	}()

	// Завершаем работу при сигнале завершения
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	fmt.Println("Shutting down...")
}

func sendNotification(user User) {
	fmt.Printf("Sending notification to user: %s, email: %s\n", user.Name, user.Email)
}
