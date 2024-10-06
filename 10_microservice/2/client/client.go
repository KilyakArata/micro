package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
	"log"
	"microservice/2/auth"
	"time"

	"google.golang.org/grpc"
	pb "microservice/2/pkg/note_v1"
)

func main() {
	// Подключаемся к gRPC серверу
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	// Генерируем JWT токен для аутентификации
	jwtToken, err := auth.GenerateJWT("client_username")
	if err != nil {
		log.Fatalf("Error generating JWT: %v", err)
	}

	// Создаем контекст с метаданными для передачи токена
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Добавляем токен в метаданные запроса
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+jwtToken)

	// Пример создания пользователя с JWT аутентификацией
	createRes, err := client.CreateUser(ctx, &pb.CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	})
	if err != nil {
		log.Fatalf("could not create user: %v", err)
	}
	fmt.Printf("Created User: %v\n", createRes.User)

	// Пример получения списка пользователей
	getRes, err := client.GetUsers(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get users: %v", err)
	}
	fmt.Printf("Users: %v\n", getRes.Users)
}
