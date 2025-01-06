package main

import (
	"flag"
	"fmt"
	"os"
)

// Type of Request and Response
type Request struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

//	type Params struct {
//		ClientInfo ClientInfo `json:"clientInfo"`
//		Message    Message    `json:"message"`
//	}
type Message struct {
	Sender      string      `json:"sender"`
	Recipient   string      `json:"recipient,omitempty"`
	Content     interface{} `json:"content"`
	MessageType string      `json:"type"`
	Timestamp   int64       `json:"timestamp"`
	Nonce       [24]byte    `json:"nonce"`
}

type ClientInfo struct {
	Username  string   `json:"username"`
	PublicKey [32]byte `json:"publicKey"`
}
type RequestKey struct {
	Target string `json:"target"`
	Sender string `json:"sender"`
}

var Methods = [4]string{"broadcast_message", "send_private_message", "provide_user_info", "get_public_key_of_client"}
var MessageTypes = [8]string{"client_online", "client_offline", "send_broadcast", "success_broadcast",
	"send_private", "success_private", "error", "public_key_of_client"}
var Results = [4]string{"success_send_broadcast", "success_send_private", "failure", "user_not_found"}

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
		RunServer(*host, *port)
	} else if *mode == "client" {
		if *username == "" {
			fmt.Println("Please provide your username.")
			os.Exit(1)
		}
		RunClient(*host, *port, *username)
	} else {
		fmt.Println("Incorrect mode:", *mode)
		os.Exit(1)
	}
}
