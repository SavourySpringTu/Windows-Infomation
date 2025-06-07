package main

import (
	"AgentClient/appsInfo"
	"AgentClient/biosInfo"
	"AgentClient/connectionsInfo"
	"AgentClient/dnsCachedInfo"
	"AgentClient/filesInfo"
	"AgentClient/hardWareInfo"
	"AgentClient/kernelModuleInfo"
	"AgentClient/networkInfo"
	"AgentClient/processesInfo"
	"AgentClient/proto"
	"AgentClient/registryInfo"
	"AgentClient/sysInfo"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
	"strconv"
	"sync"
	"time"
)

const (
	PORTSERVER = "59051"
)

var (
	AUTHENTICATION_KEY = ""
)

// Connect to server, init client, open stream and receive request from server, send response to server
func RunClient() {
	for {
		stream, conn, errCN := ConnectAndInit()
		if errCN != nil {
			fmt.Println("Connect and init error:", errCN)
		} else {
			errAuth := AuthenticateClient(stream)
			if errAuth != nil {
				fmt.Println("Authentication: ", errAuth)
				continue
			}
			Process(stream)
		}

		AUTHENTICATION_KEY = ""
		stream.CloseSend()
		conn.Close()
		time.Sleep(3 * time.Second)
	}
	return
}

// Init client, connection, open stream
func ConnectAndInit() (proto.AgentService_StreamMessageClient, *grpc.ClientConn, error) {
	fmt.Println("Finding server...........")

	// Connect to server
	conn, errConn := grpc.Dial("localhost:"+PORTSERVER, grpc.WithInsecure(), grpc.WithBlock())
	if errConn != nil {
		return nil, nil, errConn
	}

	// Init client
	client := proto.NewAgentServiceClient(conn)

	// Open stream
	stream, errStream := client.StreamMessage(context.Background())
	if errStream != nil {
		return nil, nil, errStream
	}
	return stream, conn, nil
}

func Process(stream proto.AgentService_StreamMessageClient) {
	wg := sync.WaitGroup{}
	// Authentication success, receive request
	for {
		msg, errRec := stream.Recv()
		if errRec != nil {
			fmt.Println("Receive message error: ", errRec)
			break
		}
		if msg.Auth == AUTHENTICATION_KEY {
			if msg.Type == "close" {
				fmt.Println("Disconnect!")
				fmt.Println("")
				AUTHENTICATION_KEY = ""
				break
			}

			wg.Add(1)
			go func() { // goroutine for each request handle
				result := ProcessCommandMessage(msg)
				errSend := stream.Send(result)
				if errSend != nil {
					fmt.Println("Send response error: ", errSend)
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
	return
}

// Authentication
func AuthenticateClient(stream proto.AgentService_StreamMessageClient) error {
	fmt.Println("Authenticating...........")
	errSend := stream.Send(&proto.CommandMessage{
		Type: "client",
	})
	if errSend != nil {
		return errSend
	}
	mes, errMes := stream.Recv()
	if errMes != nil {
		return errMes
	}
	if mes.Auth == "" {
		return errors.New("Access denied!")
	}
	AUTHENTICATION_KEY = mes.Auth
	fmt.Println("AUTHENTICATION KEY: ", AUTHENTICATION_KEY)
	return nil
}

// Process message from server
func ProcessCommandMessage(msg *proto.CommandMessage) *proto.CommandMessage {
	var result *proto.CommandMessage
	var data any
	var err error
	timeChan := make(chan struct{}, 1)
	go func() {
		switch msg.Type {
		case "-a":
			data, err = appsInfo.GetInfoApp()
		case "-b":
			data, err = biosInfo.GetBiosInfo()
		case "-c":
			data, err = connectionsInfo.GetConnectionsInfo()
		case "-d":
			data, err = dnsCachedInfo.GetDnsCachedInfo()
		case "-f":
			fmt.Println(msg.Parameter)
			data, err = filesInfo.GetInfoFileAndFolder(msg.Parameter)
		case "-n":
			data, err = networkInfo.GetInfoNetWork()
		case "-k":
			data, err = kernelModuleInfo.GetInfoKernelModule()
		case "-i":
			data, err = hardWareInfo.GetHardWareInfo()
		case "-r":
			data, err = registryInfo.GetInfoRegistryByPath(msg.Parameter)
		case "-s":
			data, err = sysInfo.GetSystemInfo()
		case "-p":
			var pid uint64
			var errConvert error
			var all = false
			if msg.Parameter == "" { // if parameter "" get all process
				all = true
			}
			pid, errConvert = strconv.ParseUint(msg.Parameter, 10, 32)
			if errConvert != nil {
				err = errConvert
			}
			data, err = processesInfo.GetInfoProcesses(uint32(pid), all) // get process target
		}
		timeChan <- struct{}{} // process done!
	}()
	select {
	case <-timeChan: // wait process done
		if err != nil { // if error passing err to second parameter
			result, _ = ConvertToYAML("", err.Error(), msg)
		} else { // if success
			result, _ = ConvertToYAML(data, "", msg)
		}
	case <-time.After(15 * time.Second): // if not complted after 15 second then return err time out!
		result, _ = ConvertToYAML(data, "Time out!", msg)
	}
	return result
}

// Parse and print
func ConvertToYAML(data any, err string, msg *proto.CommandMessage) (*proto.CommandMessage, error) {
	fmt.Println("Message id: ", msg.Id)
	fmt.Println("Command   : ", msg.Type, " ", msg.Parameter)
	fmt.Println("------------------------------------------------")
	dataYAML, errMar := yaml.Marshal(data)
	if err != "" {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println(string(dataYAML))
	}
	if errMar != nil {
		return nil, errMar
	}

	result := &proto.CommandMessage{
		Auth:      msg.Auth,
		Id:        msg.Id,
		Type:      msg.Type,
		Parameter: msg.Parameter,
		Error:     err,
		Data:      string(dataYAML),
	}

	return result, nil
}

// Generate id for message
func generateId() string {
	result := make([]byte, 16)
	_, err := rand.Read(result)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(result)
}

func main() {
	RunClient()
	return
}
