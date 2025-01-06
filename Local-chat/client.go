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

var publicKey *[32]byte

var privateKey *[32]byte

var contactBook = make(map[string]*[32]byte)

// var publicKeyChan = make(chan *[32]byte, 1)
// var publicKeyChan = make(chan struct{})
var publicKeyChannels = make(map[string]chan struct{})

func RunClient(host, port, username string) {
	GenerateKeyPair()
	clientName = username
	address := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Can not connect to server. Error: %v", err)
	}
	defer conn.Close()

	log.Println("Connected to server: ", address)

	//send client info to server
	clientInfo := ClientInfo{username, *publicKey}
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
			setContact(parsedClient)
			fmt.Printf("[%s]New client connected: %s(PUBLIC KEY: %v)\n", timeString, parsedClient.Username, parsedClient.PublicKey)

		case MessageTypes[1]: //client_offline
			contentToByte, _ := json.Marshal(parsedMessage.Content)
			var parsedClient ClientInfo
			_ = json.Unmarshal(contentToByte, &parsedClient)
			delete(contactBook, parsedClient.Username)
			fmt.Printf("[%s]Client disconnected: %s(PUBLIC KEY: %v)\n", timeString, parsedClient.Username, parsedClient.PublicKey)
		case MessageTypes[2]: //send_broadcast
			fmt.Printf("[%s][<<Public from %s]: %s\n", timeString, parsedMessage.Sender, parsedMessage.Content)
		case MessageTypes[3]: //success_broadcast
			fmt.Printf("[%s][>>Public]: %s\n", timeString, parsedMessage.Content)

		case MessageTypes[4]: //send_private
			contentToByte, _ := json.Marshal(parsedMessage.Content)
			var encryptedMessage []byte
			_ = json.Unmarshal(contentToByte, &encryptedMessage)
			go DecryptContent(encryptedMessage, &parsedMessage.Nonce, parsedMessage.Sender, timeString, conn)

		case MessageTypes[5]: //success_private
			contentToByte, _ := json.Marshal(parsedMessage.Content)
			var encryptedMessage []byte
			_ = json.Unmarshal(contentToByte, &encryptedMessage)
			recipientPublicKey := contactBook[parsedMessage.Recipient]
			decrypted, ok := box.Open(nil, encryptedMessage, &parsedMessage.Nonce, recipientPublicKey, privateKey)
			if !ok {
				continue
			}
			fmt.Printf("[%s][>>Private to %s]: %s\n", timeString, parsedMessage.Recipient, decrypted)
		case MessageTypes[6]: //error
			fmt.Printf("[%s][ERROR]: %s\n", timeString, parsedMessage.Content)
		case MessageTypes[7]: //public_key_of_client
			contentToByte, _ := json.Marshal(parsedMessage.Content)
			var parsedClient ClientInfo
			_ = json.Unmarshal(contentToByte, &parsedClient)
			if parsedClient.PublicKey == [32]byte{} {
				publicKeyChannels[parsedClient.Username] <- struct{}{}
				continue
			}
			setContact(parsedClient)
			publicKeyChannels[parsedClient.Username] <- struct{}{}
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
				if len(parts) >= 2 && len(strings.TrimSpace(parts[1])) > 0 && parts[0][1:] != clientName {
					message.MessageType = MessageTypes[4]
					message.Recipient = parts[0][1:]
					encryptedConntent, nonce := EncryptContent(parts[1], message.Recipient, conn)
					message.Content = encryptedConntent
					message.Nonce = nonce
					request, err = CreateRequest(Methods[1], message)
					if encryptedConntent == nil {
						continue
					}
				} else {
					log.Print("Invalid message. Make sure your message is not empty or you do not send message to yourself")
				}
			}
			if err != nil {
				log.Printf("Incorrect format of message. Error: %v", err)
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

func setContact(newContact ClientInfo) {
	key := newContact.PublicKey
	contactBook[newContact.Username] = &key
}

func getPublicKey(username string, conn net.Conn) {

	params := RequestKey{
		Target: username,
		Sender: clientName,
	}
	request, err := CreateRequest(Methods[3], params) //get_public_key_of_client
	if err != nil {
		log.Printf("Can not get public key from server. Error: %v", err)
		return
	}
	_, err = conn.Write(append(request, '\n'))

	if err != nil {
		log.Printf("Can not send request: %v", err)
		return
	}
}
func GenerateKeyPair() {
	public, private, err := box.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	privateKey = private
	publicKey = public
}

func EncryptContent(message string, recipient string, conn net.Conn) ([]byte, [24]byte) {
	var recipientPublicKey *[32]byte
	key, ok := contactBook[recipient]
	if !ok {
		getPublicKey(recipient, conn)
		publicKeyChannels[recipient] = make(chan struct{})
		<-publicKeyChannels[recipient]
		recipientPublicKey, ok = contactBook[recipient]
		if !ok {
			return nil, [24]byte{}
		}
	} else {
		recipientPublicKey = key
	}
	var nonce [24]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		panic(err)
	}
	encrypted := box.Seal(nil, []byte(message), &nonce, recipientPublicKey, (*[32]byte)(privateKey))
	return encrypted, nonce
}

func DecryptContent(encrypted []byte, nonce *[24]byte, sender string, timeString string, conn net.Conn) {

	senderPublicKey, ok := contactBook[sender]
	if !ok {
		getPublicKey(sender, conn)
		publicKeyChannels[sender] = make(chan struct{})
		<-publicKeyChannels[sender]
		senderPublicKey, ok = contactBook[sender]
		if !ok {
			return
		}
	}
	decrypted, ok := box.Open(nil, encrypted, nonce, senderPublicKey, privateKey)
	if !ok {
		fmt.Printf("[%s]not legitimate message. ALERT SCAMMER\n", timeString)
		return
	}
	fmt.Printf("[%s][<<Private from %s]: %s\n", timeString, sender, decrypted)
}
