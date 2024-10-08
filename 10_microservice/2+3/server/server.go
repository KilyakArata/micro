package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"math/rand"
	"microservice/2+3/auth"
	"net"
	"sync"

	"google.golang.org/grpc"
	pb "microservice/2+3/pkg/note_v1"
)

type server struct {
	pb.UnimplementedUserServiceServer
	sync.Mutex
	users map[int32]*pb.User
}

// JWT аутентификация
func (s *server) unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	token := md["authorization"]
	if len(token) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing token")
	}

	// Убираем префикс "Bearer " и валидируем токен
	claims, err := auth.ValidateJWT(token[0][7:])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	// Добавляем имя пользователя в контекст
	newCtx := context.WithValue(ctx, "username", claims.Username)
	return handler(newCtx, req)
}

func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	s.Lock()
	defer s.Unlock()

	user := &pb.User{
		Id:    int32(rand.Intn(1000)),
		Name:  req.Name,
		Email: req.Email,
	}

	s.users[user.Id] = user
	return &pb.UserResponse{User: user}, nil
}

func (s *server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	s.Lock()
	defer s.Unlock()

	user, exists := s.users[req.Id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	user.Name = req.Name
	user.Email = req.Email
	return &pb.UserResponse{User: user}, nil
}

func (s *server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	s.Lock()
	defer s.Unlock()

	_, exists := s.users[req.Id]
	if !exists {
		return &pb.DeleteUserResponse{Success: false}, fmt.Errorf("user not found")
	}

	delete(s.users, req.Id)
	return &pb.DeleteUserResponse{Success: true}, nil
}

func (s *server) GetUsers(ctx context.Context, req *pb.Empty) (*pb.GetUsersResponse, error) {
	s.Lock()
	defer s.Unlock()

	var users []*pb.User
	for _, user := range s.users {
		users = append(users, user)
	}

	return &pb.GetUsersResponse{Users: users}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor((&server{}).unaryInterceptor))
	pb.RegisterUserServiceServer(grpcServer, &server{users: make(map[int32]*pb.User)})

	fmt.Println("gRPC server with JWT authentication listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
