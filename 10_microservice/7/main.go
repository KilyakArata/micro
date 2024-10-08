package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

var db *sql.DB

// User структура для пользователя
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Функция для загрузки конфигурации из .env
func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла")
	}
}

// Подключение к базе данных
func connectDB() {
	var err error
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Ошибка пинга БД: %v", err)
	}
}

// Создание пользователя
func createUser(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Имя пользователя не может быть пустым", http.StatusBadRequest)
		return
	}

	var id int
	err := db.QueryRow("INSERT INTO users (name) VALUES ($1) RETURNING id", name).Scan(&id)
	if err != nil {
		http.Error(w, "Ошибка создания пользователя", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Пользователь создан с ID: %d\n", id)
}

// Удаление пользователя
func deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Ошибка удаления пользователя", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Пользователь с ID: %s удален\n", id)
}

func main() {
	loadEnv()
	connectDB()

	r := mux.NewRouter()
	r.HandleFunc("/users", createUser).Methods("POST")
	r.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")

	serverPort := os.Getenv("SERVER_PORT")
	log.Printf("Сервер запущен на порту %s", serverPort)
	http.ListenAndServe(":"+serverPort, r)
}
