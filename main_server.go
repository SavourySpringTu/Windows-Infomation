package main

import (
	proto "main/proto"
	"sync"
)

type server struct {
	proto.UnimplementedAgentClientServiceServer
	Mu      sync.Mutex
	Clients map[string]ClientStream
}

type ClientStream struct {
	Auth   string
	Name   string
	Stream proto.AgentClientService_StreamMessageClient
	Done   chan struct{}
}

func NewServer() *server {
	return &server{
		Clients: make(map[string]ClientStream),
	}
}

func (s *server) StreamMessage(stream proto.AgentClientService_StreamMessageServer) error {
	for {

	}
}

func AuthenticationServer(stream proto.UnimplementedAgentClientServiceServer) {

}

func main() {

}
