package main

import (
	"AgentServer/proto"
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	PORT = "59051"
)

type ClientStream struct {
	Auth   string
	Name   string
	Stream proto.AgentService_StreamMessageServer
	Rm     chan struct{}
}

type server struct {
	proto.UnimplementedAgentServiceServer
	Mu       sync.Mutex
	Client   map[string]*ClientStream
	Electron *ClientStream
	Count    int
}

func NewServer() *server {
	return &server{
		Client: make(map[string]*ClientStream),
		Electron: &ClientStream{
			Auth: "",
		},
	}
}

var mapCommand = map[string]int{
	"-a": 0,
	"-b": 0,
	"-c": 0,
	"-d": 0,
	"-f": 1,
	"-n": 0,
	"-k": 0,
	"-i": 0,
	"-r": 1,
	"-s": 0,
	"-p": 2,
}

var GRPC *grpc.Server

// Stream for a client connect, authen client and receive message from client
func (s *server) StreamMessage(stream proto.AgentService_StreamMessageServer) error {

	client, errAuth := ServerAuthenticateClient(s, stream)
	if errAuth != nil || client == nil {
		fmt.Println("Authenticate client fail!")
		return errAuth
	}
	fmt.Println(client.Name, "connect success!")

	// Authen success
	go func() {
		if client.Auth == s.Electron.Auth { // is electron
			ReceiveElectron(s, client)
		} else {
			if s.Electron.Auth != "" {
				LoadClient(s) // send client load message to electron every time there is a new stream
			}
			ReceiveClient(s, client)
		}
	}()

	<-client.Rm //wait server request close stream

	if s.Electron.Auth == client.Auth {
		CloseStream(client, s)
		s.Electron.Auth = ""
	} else { // if the closed stream is not stream of electron
		CloseStream(client, s)
		LoadClient(s) // send message load client to electron every time close stream
	}
	fmt.Println("Disconnect ", client.Name)
	return nil
}

func ReceiveElectron(s *server, client *ClientStream) {
	for {
		mess, errMess := client.Stream.Recv()
		if errMess != nil {
			fmt.Println("Receive message from", client.Name, ": ", errMess)
			client.Rm <- struct{}{}
			return
		}
		if mess.Auth == client.Auth {
			if mess.Type == "loadclient" {
				errLoad := LoadClient(s)
				if errLoad != nil {
					fmt.Println("Load client: ", errLoad)
					return
				}
			} else {
				cmd := "s " + mess.Data + " " + mess.Type + " " + mess.Parameter
				errProcess := ProcessInputCommand(cmd, s)
				if errProcess != nil {
					fmt.Println("Process command:", errProcess)
					SendMessageError(client, mess, errProcess.Error())
				}
			}
		}
	}
}

func LoadClient(s *server) error {
	if s.Electron.Auth != "" {
		data := ListClient(s)
		dataYAML, errMar := yaml.Marshal(data)

		if errMar != nil {
			s.Electron.Rm <- struct{}{}
			return errMar
		} else {
			errSend := s.Electron.Stream.Send(&proto.CommandMessage{
				Type:  "loadclient",
				Data:  string(dataYAML),
				Error: "",
			})
			if errSend != nil {
				s.Electron.Rm <- struct{}{}
				return errSend
			}
		}
		return nil
	}
	return errors.New("Electron invalid!")
}

func ReceiveClient(s *server, client *ClientStream) {
	for {
		mess, errMess := client.Stream.Recv()
		if errMess != nil {
			fmt.Println("Receive message from", client.Name, ": ", errMess)
			client.Rm <- struct{}{}
			return
		}
		// Check auth and print message
		if mess.Auth == client.Auth {
			fmt.Println("------------------- " + client.Name + " -------------------------")
			if mess.Error == "" {
				yaml, _ := yaml.Marshal(mess.Data)
				fmt.Println(string(yaml))
			} else {
				yaml, _ := yaml.Marshal(mess.Error)
				fmt.Println(string(yaml))
			}

			// Send message to electron
			if s.Electron.Auth != "" {
				id := generateIdServer()
				mess.Auth = s.Electron.Auth
				mess.Id = id
				errSend := s.Electron.Stream.Send(mess)
				if errSend != nil {
					fmt.Println("Send Message to Electron:", errSend)
				}
			} else {
				fmt.Println("Authentication of Electrion invalid!")
			}
		}
	}
}

// Authen client and add connect client  to map
func ServerAuthenticateClient(s *server, stream proto.AgentService_StreamMessageServer) (*ClientStream, error) {
	s.Mu.Lock()
	mess, errMess := stream.Recv()
	if errMess != nil {
		return nil, errMess
	}

	id := generateIdServer()
	auth := "client-" + id // auth = `client-<key>`

	name := "client" + strconv.Itoa(s.Count) // name = `client + Coutn int server`
	// Send authen key to client
	errSend := stream.Send(&proto.CommandMessage{
		Auth:  auth,
		Id:    id,
		Type:  mess.Type,
		Data:  "Authentication success!",
		Error: "",
	})
	if errSend != nil {
		return nil, errSend
	}

	rm := make(chan struct{}, 1)
	clientStream := &ClientStream{ // Init new ClientStream
		Auth:   auth,
		Name:   name,
		Stream: stream,
		Rm:     rm,
	}

	s.Count++
	if mess.Type == "electron" { // if client is Electron
		s.Electron = clientStream
	} else {
		s.Client[auth] = clientStream // add clientstream to map
	}
	s.Mu.Unlock()
	return clientStream, nil
}

// Console to input command
func Console(s *server, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	reader := bufio.NewReader(os.Stdin)
	for {
		str, errRead := reader.ReadString('\n')
		if errRead != nil {
			fmt.Println("Read line: ", errRead)
			continue
		}
		// Process command
		errCmd := ProcessInputCommand(str, s)
		if errCmd != nil {
			if errCmd.Error() == "shutdown" {
				return
			}
			fmt.Println("Process command: ", errCmd)
		}
	}
	return
}

// Process input command and send request to client target, show list client, register client
func ProcessInputCommand(str string, s *server) error {
	str = strings.TrimSpace(str)
	sliceCommand := strings.Fields(str)

	lenCmd := len(sliceCommand)

	if lenCmd == 0 || lenCmd > 4 {
		return errors.New("Command invalid!")
	}

	if sliceCommand[0] == "ls" && lenCmd == 1 {

		// If the first command is `ls` then show list client
		list := ListClient(s)
		for _, j := range list {
			fmt.Println(j)
		}

	} else if sliceCommand[0] == "rm" && lenCmd == 2 {

		// If the first commad is `rm` then delete client target
		client, exist := FindClientByName(sliceCommand[1], s.Client)
		if exist != nil {
			return errors.New("Can't find client!")
		}
		client.Rm <- struct{}{}

	} else if sliceCommand[0] == "s" { // Else fisrt command is `name client`

		if lenCmd < 3 {
			return errors.New("Missing parameter!")
		}

		// Find Client to send request
		client, exist := FindClientByName(sliceCommand[1], s.Client)
		if exist != nil {
			return errors.New("Can't find client!")
		}
		typeCmd := ""
		param := ""

		_, existType := mapCommand[sliceCommand[2]]
		if existType == false {
			return errors.New("Flag Invalid!")
		}
		typeCmd = sliceCommand[2] // flag

		if mapCommand[sliceCommand[2]] == 1 {
			if lenCmd < 4 {
				return errors.New("Missing parameter!")
			}
			param = sliceCommand[3] // parameter
		} else if mapCommand[sliceCommand[2]] == 2 {
			if lenCmd >= 4 {
				param = sliceCommand[3] // parameter
			}
		} else {
			if lenCmd > 3 {
				return errors.New("This command has no parameters!")
			}
		}

		// generate id message
		id := generateIdServer()

		errSend := client.Stream.Send(&proto.CommandMessage{
			Auth:      client.Auth,
			Id:        id,
			Type:      typeCmd,
			Parameter: param,
			Error:     "",
		})
		if errSend != nil {
			return errSend
		}
	} else if sliceCommand[0] == "help" && lenCmd == 1 {
		HelpCommand()
	} else if sliceCommand[0] == "shutdown" && lenCmd == 1 { // Shut down server
		ShutDownServer(s)
		return errors.New("shutdown")
	} else {
		return errors.New("Command invalid!")
	}
	return nil
}

func HelpCommand() {
	fmt.Println("")
	fmt.Println("ls                           ", "-", " Show list client")
	fmt.Println("s <NameClient> <flag> <path> ", "-", " Send requets to client target, <path> non ``")
	fmt.Println("rm <NameClient>              ", "-", " Remove client target")
	fmt.Println("shutdown                     ", "-", " Shut down server")
	fmt.Println("help                         ", "-", " Show command")
}

// Show list client
func ListClient(s *server) map[string]string {
	var result = make(map[string]string)
	for _, j := range s.Client {
		result[j.Auth] = j.Name
	}
	return result
}

// send message error
func SendMessageError(client *ClientStream, msg *proto.CommandMessage, error string) {
	client.Stream.Send(&proto.CommandMessage{
		Id:    msg.Id,
		Auth:  msg.Auth,
		Error: error,
	})
}

// Find client in map
func FindClientByName(name string, clients map[string]*ClientStream) (*ClientStream, error) {
	for _, j := range clients {
		if j.Name == name {
			return j, nil
		}
	}
	return &ClientStream{}, errors.New("Can't find Client!")
}

// Shut down server
func ShutDownServer(s *server) {
	fmt.Println("Shutting down....................")
	go func() {
		for _, i := range s.Client {
			i.Rm <- struct{}{}
		}
	}()
	GRPC.GracefulStop()
}

// Send message close stream to client
func CloseStream(client *ClientStream, s *server) {
	s.Mu.Lock()
	delete(s.Client, client.Auth)
	s.Mu.Unlock()
}

func main() {
	// Find port
	listen, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		fmt.Println("Listen error: ", err)
		return
	}

	// New server
	grpcServer := grpc.NewServer()
	GRPC = grpcServer
	sv := NewServer()
	proto.RegisterAgentServiceServer(grpcServer, sv)
	fmt.Println("Server started on port: ", PORT)

	var wg sync.WaitGroup

	// Console for server
	go func() {
		Console(sv, &wg)
	}()

	// Listen connect from client
	errServer := grpcServer.Serve(listen)
	if errServer != nil {
		fmt.Println("Server listen: ", errServer)
		return
	}
	wg.Wait()
	return
}

func generateIdServer() string {
	result := make([]byte, 16)
	_, err := rand.Read(result)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(result)
}
