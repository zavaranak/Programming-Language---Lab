package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Client struct {
	conn       net.Conn
	clientInfo ClientInfo
}

// Global variables
var (
	clients         = make(map[net.Addr]*Client) //stored clients map: {key:address => value:Client}
	clientsMutex    sync.RWMutex
	listener        net.Listener
	closeServerChan chan struct{}
)

const serverName string = "Server"

func RunServer(host, port string) {
	//create channel
	closeServerChan = make(chan struct{})
	defer func() {
		close(closeServerChan)
		log.Println("Server is shutting down...")
	}()

	// Generate server listener tcp
	address := fmt.Sprintf("%s:%s", host, port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Can not start server listener:  %v", err)
	}
	defer ln.Close() //close server listener

	listener = ln
	log.Println("Server is running at ", address)
	go func() {
		for {
			select {
			case <-closeServerChan:
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					log.Printf("An error was occupied while accepting connection from client: %v", err)
					continue
				}
				go handleConnection(conn)
			}
		}
	}()

	//Wait for signal "closeServerChan" while thread "go func()" is still running
	<-closeServerChan
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	data, err := reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			log.Printf("An error was occupied while reading information of client: %v", err)
		}
		return
	}

	connAddr, err := handleFirstRequest(conn, data)

	if err != nil {
		if err != io.EOF {
			log.Printf("An error was occupied while handling information of client: %v", err)
		}
		return
	}

	clientInfo := clients[connAddr].clientInfo

	defer func() {
		clientsMutex.Lock()
		delete(clients, conn.RemoteAddr())
		clientsMutex.Unlock()
		log.Printf("Client %s (%s) disconnected", clientInfo.Username, conn.RemoteAddr())
		message := Message{
			Sender:      serverName,
			Content:     clientInfo,
			MessageType: MessageTypes[1],
			Timestamp:   time.Now().Unix(),
		}
		broadcastMessage(message)
	}()
	// Listen to message from current connection (client) and pass it to handleMessage goroutine
	for {
		jsonRpcRequest, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("An error was occupied while reading request from %s: %v", clientInfo.Username, err)
			}
			return
		}
		jsonRpcRequest = strings.TrimSpace(jsonRpcRequest)
		log.Printf("New request from %s\n", clientInfo.Username)
		go handleMessage(conn, jsonRpcRequest)
	}
}

func handleFirstRequest(conn net.Conn, data string) (net.Addr, error) {
	request, err := RequestParser([]byte(data))
	if err != nil {
		return nil, err
	}
	params := request.Params
	paramsToByte, _ := json.Marshal(params)
	var clientInfo ClientInfo
	_ = json.Unmarshal(paramsToByte, &clientInfo)
	client := &Client{
		conn:       conn,
		clientInfo: clientInfo,
	}
	clientsMutex.Lock()
	clients[conn.RemoteAddr()] = client
	clientsMutex.Unlock()

	message := Message{
		Sender:      serverName,
		Content:     client.clientInfo,
		MessageType: MessageTypes[0],
		Timestamp:   time.Now().Unix(),
	}

	_ = broadcastMessage(message)
	log.Printf("Client %s (%s) connected", clientInfo.Username, conn.RemoteAddr())

	return conn.RemoteAddr(), err
}

func handleMessage(conn net.Conn, jsonRpcRequest string) {
	request, err := RequestParser([]byte(jsonRpcRequest))
	if err != nil {
		log.Printf("Error occured while parsing request from client: %v", err)
		return
	}
	var message Message
	var result string
	message.MessageType = MessageTypes[2]
	params := request.Params
	//Methods  "broadcast_message", "send_private_message", "provide_user_info", "get_public_key_of_client"
	switch request.Method {
	case Methods[0]: //broadcast_message
		paramsToByte, _ := json.Marshal(params)
		_ = json.Unmarshal(paramsToByte, &message)
		message.Timestamp = time.Now().Unix()
		result = broadcastMessage(message)

	case Methods[1]: //send_private_message)
		paramsToByte, _ := json.Marshal(params)
		_ = json.Unmarshal(paramsToByte, &message)
		message.Timestamp = time.Now().Unix()
		client := getClient(message.Recipient)
		if client == nil {
			result = Results[2]
			break
		}
		result = sendMessageToClient(client.conn, message)

	case Methods[3]: //get_public_key_of_client
		clientsMutex.RLock()

		clientsMutex.RUnlock()
		client := getClient(message.Recipient)
		if client == nil {
			result = Results[2]
			break
		}
		result = sendMessageToClient(client.conn, message)
	}

	//Response
	switch result {
	case Results[0]:
		message.MessageType = MessageTypes[3]
		sendMessageToClient(conn, message)
	case Results[1]:
		message.MessageType = MessageTypes[5]
		sendMessageToClient(conn, message)
	case Results[2]:
		var errorMessage Message
		errorMessage.MessageType = MessageTypes[4]
		errorMessage.Content = ("Unable to send message")
		errorMessage.Sender = serverName
		sendMessageToClient(conn, errorMessage)
	case Results[3]:
		var errorMessage Message
		errorMessage.MessageType = MessageTypes[4]
		errorMessage.Content = "User " + message.Recipient + " not found"
		errorMessage.Sender = serverName
		errorMessage.Timestamp = time.Now().Unix()
		sendMessageToClient(conn, errorMessage)
	}

}

func getClient(username string) *Client {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()
	for _, client := range clients {
		if client.clientInfo.Username == username {
			return client
		}
	}
	return nil
}
func sendMessageToClient(client net.Conn, message Message) string {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()
	marshaledMessage, err := json.Marshal(message)
	if err != nil {
		return Results[2]
	}
	_, err = client.Write(append(marshaledMessage, '\n'))
	if err != nil {
		return Results[2]
	}
	return Results[1]
}
func broadcastMessage(message Message) string {
	marshaledMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error while marshaling message, error: %v\n", err)
		return Results[2]
	}
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()
	for _, client := range clients {
		if client.clientInfo.Username != message.Sender {
			_, err = client.conn.Write(append(marshaledMessage, '\n'))
			if err != nil {
				log.Printf("Error occured while broadcasting message")
				return Results[2]
			}
		}
	}
	return Results[0]
}
func RequestParser(data []byte) (Request, error) {
	var request Request
	err := json.Unmarshal([]byte(data), &request)
	return request, err
}
