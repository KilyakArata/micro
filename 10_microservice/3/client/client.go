package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func main() {
	// Подключаемся к серверу
	ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("Ошибка подключения:", err)
	}
	defer ws.Close()

	// Отправляем текстовое сообщение
	message := "Привет, сервер!"
	err = ws.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Fatal("Ошибка отправки сообщения:", err)
	}
	fmt.Println("Отправлено текстовое сообщение:", message)

	// Отправляем бинарное сообщение
	binaryData := []byte{0x01, 0x02, 0x03, 0x04}
	err = ws.WriteMessage(websocket.BinaryMessage, binaryData)
	if err != nil {
		log.Fatal("Ошибка отправки бинарного сообщения:", err)
	}
	fmt.Println("Отправлено бинарное сообщение:", binaryData)

	// Чтение сообщений от сервера
	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Println("Ошибка чтения сообщения:", err)
				return
			}
			fmt.Printf("Получено сообщение от сервера: %s\n", message)
		}
	}()

	// Держим соединение открытым
	time.Sleep(10 * time.Second)
	fmt.Println("Закрываем соединение")
}
