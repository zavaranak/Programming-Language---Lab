package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/nacl/box"
	// "crypto/ed25519"
)

var clientName string

var publicKey []byte

// var privateKey []byte

var userList = make(map[string]*[]byte)

func RunClient(host, port, username string) {
	clientName = username
	address := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Can not connect to server. Error: %v", err)
	}
	defer conn.Close()
	log.Println("Connected to server: ", address)
	publicKey = []byte("byte")
	clientInfo := ClientInfo{username, publicKey}
	request, err := CreateRequest(Methods[2], clientInfo)
	if err != nil {
		log.Fatalf("Can not create first request to server. Error: %v", err)
	}
	_, err = conn.Write(append(request, '\n'))
	if err != nil {
		log.Fatalf("Can not write send first request to server. Error: %v", err)
	}

	go receiveMessages(conn)
	sendMessage(conn)
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

		var parsedMessage Message
		_ = json.Unmarshal([]byte(message), &parsedMessage)

		timeString := time.Unix(parsedMessage.Timestamp, 0).Format("15:04")

		switch parsedMessage.MessageType {
		case MessageTypes[0]: //client_online
			contentToByte, _ := json.Marshal(parsedMessage.Content)
			var parsedClient ClientInfo
			_ = json.Unmarshal(contentToByte, &parsedClient)

			userList[parsedClient.Username] = &parsedClient.PublicKey
			fmt.Printf("[%s]New client connected: %s(PUBLIC KEY: %v)\n", timeString, parsedClient.Username, parsedClient.PublicKey)

		case MessageTypes[1]: //client_offline
			contentToByte, _ := json.Marshal(parsedMessage.Content)
			var parsedClient ClientInfo
			_ = json.Unmarshal(contentToByte, &parsedClient)
			delete(userList, parsedClient.Username)
			fmt.Printf("[%s]Client disconnected: %s(PUBLIC KEY: %v)\n", timeString, parsedClient.Username, parsedClient.PublicKey)
		case MessageTypes[2]: //send_broadcast
			fmt.Printf("[%s][<<Public from %s]: %s", timeString, parsedMessage.Sender, parsedMessage.Content)
		case MessageTypes[3]: //success_broadcast
			fmt.Printf("[%s][>>Public]: %s", timeString, parsedMessage.Content)

		case MessageTypes[4]: //send_private
			fmt.Printf("[%s][<<Private from %s]: %s", timeString, parsedMessage.Sender, parsedMessage.Content)

		case MessageTypes[5]: //success_private
			fmt.Printf("[%s][>>Private to %s]: %s", timeString, parsedMessage.Recipient, parsedMessage.Content)

		case MessageTypes[6]: //error
			fmt.Printf("[%s][ERROR]: %s", timeString, parsedMessage.Content)
		case MessageTypes[7]: //public_key_of_client
			contentToByte, _ := json.Marshal(parsedMessage.Content)
			var parsedClient ClientInfo
			_ = json.Unmarshal(contentToByte, &parsedClient)
			userList[parsedClient.Username] = &parsedClient.PublicKey
		}

	}
}

func sendMessage(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		var request []byte
		data, _ := reader.ReadString('\n')

		if len(strings.TrimSpace(data)) >= 1 {
			var err error
			var message Message
			message.Sender = clientName

			if !strings.HasPrefix(data, "@") {
				message.MessageType = MessageTypes[2]
				message.Content = data
				request, err = CreateRequest(Methods[0], message)
			} else {
				parts := strings.SplitN(data, " ", 2)
				if len(parts) >= 2 {
					message.MessageType = MessageTypes[4]
					message.Recipient = parts[0][1:]
					message.Content = parts[1]
					request, err = CreateRequest(Methods[1], message)
				}
			}
			if err != nil {
				log.Printf("Can not create JSON-rpc request to server. Error: %v", err)
				return
			}
		}
		_, err := conn.Write(append(request, '\n'))
		if err != nil {
			log.Printf("Can not send message. Error: %v", err)
			return
		}
	}
}
func CreateRequest(method string, params interface{}) ([]byte, error) {

	// var test []byte = []byte("test")
	var request Request
	request.Method = method
	request.Params = params
	marshaledRequest, err := json.Marshal(request)
	return marshaledRequest, err
}

//	func getPublicKey(username string, client net.Conn) {
//		params := make(map[string]interface{})
//		request, err := CreateRequest(Methods[3], params)
//		if err != nil {
//			log.Printf("Can not get public key from server. Error: %v", err)
//		}
//		return
//	}
func GenerateKeyPair() (*[32]byte, *[32]byte) {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	return publicKey, privateKey
}

func EncryptContent(message string, recipientPublicKey *[32]byte, senderPrivateKey *[32]byte) ([]byte, *[24]byte) {
	var nonce [24]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		panic(err)
	}

	encrypted := box.Seal(nil, []byte(message), &nonce, recipientPublicKey, senderPrivateKey)
	return encrypted, &nonce
}

func DecryptContetnt(encrypted []byte, nonce *[24]byte, senderPublicKey *[32]byte, recipientPrivateKey *[32]byte) (string, error) {
	decrypted, ok := box.Open(nil, encrypted, nonce, senderPublicKey, recipientPrivateKey)
	if !ok {
		return "", fmt.Errorf("decryption failed")
	}
	return string(decrypted), nil
}
