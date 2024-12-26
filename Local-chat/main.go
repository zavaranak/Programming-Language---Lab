package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

// Structure for storing client (connections to server)
type Client struct {
	conn     net.Conn
	username string
}

// Global variables
var (
	clients         = make(map[net.Addr]*Client) //stored clients map: {key:address => value:Client}
	clientsMutex    sync.RWMutex
	listener        net.Listener
	closeServerChan chan struct{}
)

const serverName string = "Server"

// Main block
func main() {
	mode := flag.String("mode", "", "execution mode (server or client)")
	host := flag.String("host", "127.0.0.1", "host/domain")
	port := flag.String("port", "8080", "port")
	username := flag.String("username", "", "username")
	flag.Parse()

	if *mode == "" {
		fmt.Println("Please choose execution mode: 'server' or 'client'.")
		os.Exit(1)
	}

	if *mode == "server" {
		runServer(*host, *port)
	} else if *mode == "client" {
		if *username == "" {
			fmt.Println("Please provide your username.")
			os.Exit(1)
		}
		runClient(*host, *port, *username)
	} else {
		fmt.Println("Incorrect mode:", *mode)
		os.Exit(1)
	}
}

func runServer(host, port string) {
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
				conn, err := ln.Accept()
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

	//Read username
	reader := bufio.NewReader(conn)
	username, err := reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			log.Printf("An error was occupied while reading connection from client: %v", err)
		}
		return
	}
	username = strings.TrimSpace(username)

	client := &Client{
		conn:     conn,
		username: username,
	}
	clientsMutex.Lock()
	clients[conn.RemoteAddr()] = client
	clientsMutex.Unlock()

	defer func() {
		clientsMutex.Lock()
		delete(clients, conn.RemoteAddr())
		clientsMutex.Unlock()
		log.Printf("Client %s (%s) disconnected", username, conn.RemoteAddr())

		broadcastMessage(serverName, fmt.Sprintf("Client %s disconnected", username))
	}()
	broadcastMessage(serverName, fmt.Sprintf("Client %s connected", username))
	log.Printf("Client %s (%s) connected", username, conn.RemoteAddr())

	//Listen to message from current connection (client) and pass it to handleMessage goroutine
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("An error was occupied while reading message from %s: %v", username, err)
			}
			return
		}
		message = strings.TrimSpace(message)
		log.Printf("Message from %s: %s", username, message)
		go handleMessage(username, message)
	}
}

func handleMessage(sender string, message string) {
	//handle private message
	if strings.HasPrefix(message, "@") {
		parts := strings.SplitN(message, " ", 2)
		if len(parts) < 2 {
			sendMessageToClient(sender, "Message does not match the format ( @ussername message )", sender, true)
			return
		}

		recipient := strings.TrimPrefix(parts[0], "@")
		actualMessage := parts[1]
		sendMessageToClient(sender, actualMessage, recipient, true)
	} else {
		//hanle broadcast message
		broadcastMessage(sender, message)
	}
}

func broadcastMessage(sender, message string) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	//broadcast message to all user
	for _, client := range clients {
		if client.username != sender {
			sendMessageToClient(sender, message, client.username, false)
		}
	}
}

func sendMessageToClient(sender, message, recipient string, private bool) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	//Send message to recipent
	for _, client := range clients {
		if client.username == recipient {
			var message_sender string
			if sender != serverName && private {
				message_sender = "[PRIVATE]" + sender
			} else {
				message_sender = sender
			}
			_, err := client.conn.Write([]byte(fmt.Sprintf("%s: %s\n", message_sender, message)))
			if err != nil {
				log.Printf("An error was occupied while sending messages to %s: %v", recipient, err)
			}
			return
		}
	}

	//Response to sender if recipient not found
	log.Printf("User '%s' not found.", recipient)
	for _, client := range clients {
		if client.username == sender {
			var serverResponse string = "User " + recipient + " not found"
			_, err := client.conn.Write([]byte(fmt.Sprintf("%s: %s\n", serverName, serverResponse)))
			if err != nil {
				log.Printf("An error was occupied while sending messages to %s: %v", sender, err)
			}
			return
		}
	}
}

func runClient(host, port, username string) {
	address := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Can not connect to server. Error: %v", err)
	}
	defer conn.Close()

	log.Println("Connected to server: ", address)
	_, err = conn.Write([]byte(username + "\n"))
	if err != nil {
		log.Fatalf("Can not write send client's username. Error: %v", err)
	}

	go receiveMessages(conn)
	sendMessage(conn, username)
}

func receiveMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Can not read message from server. Error: %v", err)
			}
			return
		}
		if len(strings.TrimSpace(message)) > 1 {
			log.Printf("%s", message)
		}
	}
}

func sendMessage(conn net.Conn, username string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		message, _ := reader.ReadString('\n')
		_, err := conn.Write([]byte(message))
		if err != nil {
			log.Printf("Can not send message. Error: %v", err)
			return
		}
		if len(strings.TrimSpace(message)) > 1 {
			if !strings.HasPrefix(message, "@") {
				log.Printf("You: %s", message)
			} else {
				parts := strings.SplitN(message, " ", 2)
				if len(parts) >= 2 {
					target := parts[0]
					log.Printf("[PRIVATE]You to %s: %s", target[1:], parts[1]) // Log recipient and message
				}
			}
		}
	}
}
