package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserStore struct {
	sync.Mutex
	users map[int]User
}

var store = UserStore{
	users: make(map[int]User),
}

// Функция для регистрации нового пользователя (POST /users)
func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	store.Lock()
	defer store.Unlock()

	user.ID = rand.Intn(1000) // Генерируем случайный ID
	store.users[user.ID] = user

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Функция для обновления данных пользователя (PUT /users/{id})
func updateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/users/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	store.Lock()
	defer store.Unlock()

	if _, ok := store.users[id]; !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user.ID = id // Сохраняем ID
	store.users[id] = user

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Функция для удаления пользователя (DELETE /users/{id})
func deleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/users/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	store.Lock()
	defer store.Unlock()

	if _, ok := store.users[id]; !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	delete(store.users, id)
	w.WriteHeader(http.StatusNoContent) // Успешное удаление без контента
}

// Функция для получения списка всех пользователей (GET /users)
func getUsers(w http.ResponseWriter, r *http.Request) {
	store.Lock()
	defer store.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(store.users)
}

func main() {
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createUser(w, r)
		case http.MethodGet:
			getUsers(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			updateUser(w, r)
		case http.MethodDelete:
			deleteUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
