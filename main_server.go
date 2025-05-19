package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	proto "main/proto"
)

// Định nghĩa struct clientStream để quản lý mỗi client
type clientStream struct {
	id     string
	stream proto.MessageCommandService_StreamMessageCommandServer
	done   chan struct{}
}

// server implement interface MessageCommandServiceServer
type server struct {
	proto.UnimplementedMessageCommandServiceServer
	mu      sync.Mutex
	clients map[string]*clientStream
}

// Hàm tạo server mới, khởi tạo map clients
func newServer() *server {
	return &server{
		clients: make(map[string]*clientStream),
	}
}

// Implement method streaming 2 chiều của gRPC service
func (s *server) StreamMessageCommand(stream proto.MessageCommandService_StreamMessageCommandServer) error {
	// Nhận message đầu tiên lấy client ID
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	clientID := msg.Id

	client := &clientStream{
		id:     clientID,
		stream: stream,
		done:   make(chan struct{}),
	}

	// Thêm client vào map với mutex bảo vệ
	s.mu.Lock()
	s.clients[clientID] = client
	s.mu.Unlock()

	log.Printf("Client connected: %s", clientID)

	// Goroutine nhận message từ client liên tục
	go func() {
		defer close(client.done)
		for {
			msg, err = stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("Error receiving from %s: %v", clientID, err)
				return
			}
			log.Printf("Message from %s: %v", clientID, msg)
			// Ví dụ: broadcast message cho client khác
			err = s.sendToClient(msg, "clientID-được-chỉ-định")
			if err != nil {
				log.Printf("Send message failed: %v", err)
			}
		}
	}()

	// Chờ client disconnect
	<-client.done

	// Xóa client khỏi map khi ngắt kết nối
	s.mu.Lock()
	delete(s.clients, clientID)
	s.mu.Unlock()

	log.Printf("Client disconnected: %s", clientID)
	return nil
}

// sendToClient gửi message tới client có ID được chỉ định
func (s *server) sendToClient(msg *proto.MessageCommand, clientID string) error {
	s.mu.Lock()
	client, ok := s.clients[clientID]
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("client %s not found", clientID)
	}

	err := client.stream.Send(msg)
	if err != nil {
		log.Printf("Error sending to %s: %v", clientID, err)
		return err
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterMessageCommandServiceServer(grpcServer, newServer())
	log.Println("Server started on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
