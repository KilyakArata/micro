package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

var (
	store = UserStore{
		users: make(map[int]User),
	}

	// Метрики Prometheus
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	// Регистрация метрик в Prometheus
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

func measureRequestDuration(method, endpoint string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	}
}

// Функция для регистрации нового пользователя (POST /users)
func createUser(w http.ResponseWriter, r *http.Request) {
	httpRequestsTotal.WithLabelValues(r.Method, "/users").Inc()

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	store.Lock()
	defer store.Unlock()

	user.ID = rand.Intn(1000) // Генерация случайного ID
	store.users[user.ID] = user

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Другие обработчики (updateUser, deleteUser, getUsers) можно реализовать аналогично...

func main() {
	// Добавление обработчика для метрик
	http.Handle("/metrics", promhttp.Handler())

	// Обработчики с измерением времени
	http.HandleFunc("/users", measureRequestDuration("POST", "/users", createUser))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
