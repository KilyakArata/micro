package main

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"net/http"
	"sync"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var (
	users = make(map[string]User)
	mu    sync.Mutex
)

func main() {
	http.HandleFunc("/users", handleUsers)
	http.HandleFunc("/users/", handleUserByID)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		registerUser(w, r)
	case "GET":
		getUsers(w)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/users/"):]

	switch r.Method {
	case "PUT":
		updateUser(w, r, id)
	case "DELETE":
		deleteUser(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	users[user.ID] = user
	mu.Unlock()

	// Отправляем сообщение в Kafka
	sendKafkaMessage(user)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func updateUser(w http.ResponseWriter, r *http.Request, id string) {
	var updatedUser User
	err := json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	user, exists := users[id]
	if !exists {
		mu.Unlock()
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user.Name = updatedUser.Name
	user.Email = updatedUser.Email
	users[id] = user
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func deleteUser(w http.ResponseWriter, r *http.Request, id string) {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := users[id]; exists {
		delete(users, id)
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
	}
}

func getUsers(w http.ResponseWriter) {
	mu.Lock()
	defer mu.Unlock()

	json.NewEncoder(w).Encode(users)
}

func sendKafkaMessage(user User) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}
	defer producer.Close()

	message, err := json.Marshal(user)
	if err != nil {
		log.Fatalf("Error marshaling user: %v", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: "user_registered",
		Value: sarama.StringEncoder(message),
	}

	_, _, err = producer.SendMessage(msg)
	if err != nil {
		log.Fatalf("Error sending Kafka message: %v", err)
	}

	fmt.Println("Message sent to Kafka:", string(message))
}
