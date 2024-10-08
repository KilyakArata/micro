package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Разрешаем запросы от всех источников
		return true
	},
}

// Хранилище для подключенных клиентов
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

// Структура сообщения
type Message struct {
	MessageType string
	Data        []byte
}

// Обработчик WebSocket соединений
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Обновляем соединение до WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// Регистрируем клиента
	clients[ws] = true
	fmt.Println("Новый клиент подключен!")

	for {
		var msg Message
		// Читаем новое сообщение от клиента
		messageType, messageData, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Ошибка при чтении: %v", err)
			delete(clients, ws)
			break
		}

		// Определяем тип сообщения (текст или бинарные данные)
		switch messageType {
		case websocket.TextMessage:
			msg = Message{MessageType: "text", Data: messageData}
		case websocket.BinaryMessage:
			msg = Message{MessageType: "binary", Data: messageData}
		default:
			log.Println("Неизвестный тип сообщения")
			continue
		}

		// Передаем сообщение в канал для рассылки всем клиентам
		broadcast <- msg
	}
}

// Функция рассылки сообщений всем клиентам
func handleMessages() {
	for {
		// Получаем сообщение из канала
		msg := <-broadcast
		for client := range clients {
			// Отправляем сообщение всем клиентам
			err := client.WriteMessage(websocket.TextMessage, msg.Data)
			if err != nil {
				log.Printf("Ошибка отправки сообщения клиенту: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	// Обработчик WebSocket соединений
	http.HandleFunc("/ws", handleConnections)
	// Запуск асинхронной обработки сообщений
	go handleMessages()

	// Старт сервера
	port := ":8080"
	fmt.Printf("Сервер запущен на порту %s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
