package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const MAX_CONNECTIONS = 10

var (
	mutex       sync.Mutex //protect shared resourese like chat history
	chatHistory []string //slice that store chat history
	clients     []*Client //slice to store the connected client
)

//struct represents a connected client, with fields for the client's name and the connection
type Client struct {
	Name       string
	Connection net.Conn
}

func StartTCPServer(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Printf("Error: Can't start the server on port %d: %v\n", port, err)
		return
	}
	defer listener.Close()

	localIP := GetIPv4()
	fmt.Printf("Server is running at IP address: %s and port: %d\n", localIP, port)

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error: Unable to accept the new connection: %v\n", err)
			continue
		}

		mutex.Lock()
		if len(clients) >= MAX_CONNECTIONS {
			mutex.Unlock()
			connection.Write([]byte("Sorry, but we have reached the maximum number of connections\n"))
			connection.Close()
			continue
		}
		mutex.Unlock()

		client := &Client{
			Connection: connection,
		}

		go HandleClient(client)
	}
}
//HandleClient Function:
//This function handles the client connection, including:
//Printing a welcome message and prompting the client for a name.
//Verifying that the chosen name is unique.
//Notifying all other clients about the new client.
//Sending the chat history to the new client.
//Handling the client's messages, adding them to the chat history, and broadcasting them to all other clients.
//Removing the client from the list of connected clients when the connection is closed.


func HandleClient(client *Client) {
	defer client.Connection.Close()

	// Print welcome message and prompt for name
	PrintLinuxLogo(client.Connection)

	// Read client's name
	// Loop until a unique name is provided
	for {
		client.Connection.Write([]byte("[Enter your name]: "))
		reader := bufio.NewReader(client.Connection)
		name, err := reader.ReadString('\n')
		if err != nil || strings.TrimSpace(name) == "" {
			client.Connection.Write([]byte(fmt.Sprintf("Error: invalid client name: %v \n", err)))
			continue
		}
		name = strings.TrimSpace(name)

		mutex.Lock()
		if verifyName(name) {
			client.Name = strings.TrimSpace(name)
			mutex.Unlock()
			break
		} else {
			mutex.Unlock()
			client.Connection.Write([]byte("Name already taken. Please choose a different name.\n"))
		}
	}
	// Notify all clients that a new client has joined
	mutex.Lock()
	for _, c := range clients {
		if c != client {
			c.Connection.Write([]byte(fmt.Sprintf("\n%s has joined the chat...\n", client.Name)))
		} else {
			c.Connection.Write([]byte(fmt.Sprintf("%s has joined the chat...\n", client.Name)))
		}

		c.Connection.Write([]byte("[" + time.Now().Format("2006-01-02 15:04:05") + "][" + c.Name + "]: "))
	}
	clients = append(clients, client)
	mutex.Unlock()

	SendChatHistory(client)

	client.Connection.Write([]byte("[" + time.Now().Format("2006-01-02 15:04:05") + "][" + client.Name + "]: "))
	scanner := bufio.NewScanner(client.Connection)
	for scanner.Scan() {
		message := scanner.Text()

		if strings.TrimSpace(message) == "" {
			client.Connection.Write([]byte("[" + time.Now().Format("2006-01-02 15:04:05") + "][" + client.Name + "]: "))
			continue 
		}

		messageToSend := fmt.Sprintf("[%s][%s]: %s\n", time.Now().Format("2006-01-02 15:04:05"), client.Name, message)

		mutex.Lock()
		chatHistory = append(chatHistory, messageToSend)
		for _, c := range clients {
			if c != client {
				c.Connection.Write([]byte("\n" + messageToSend))
			}
			c.Connection.Write([]byte("[" + time.Now().Format("2006-01-02 15:04:05") + "][" + c.Name + "]: "))
		}
		mutex.Unlock()
	}

	mutex.Lock()
	for i, c := range clients {
		if c == client {
			clients = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	for _, c := range clients {
		if c != client {
			c.Connection.Write([]byte(fmt.Sprintf("\n%s has left the chat...\n", client.Name)))
		} else {
			c.Connection.Write([]byte(fmt.Sprintf("%s has left the chat...\n", client.Name)))
		}

		c.Connection.Write([]byte("[" + time.Now().Format("2006-01-02 15:04:05") + "][" + c.Name + "]: "))
	}
	mutex.Unlock()
}

//sends the chat history to a specific client
func SendChatHistory(client *Client) {
	mutex.Lock()
	defer mutex.Unlock()

	for _, message := range chatHistory {
		client.Connection.Write([]byte(message))
	}
}

//retrieves the IPv4 address of the server
func GetIPv4() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP
}


//function checks if the given name is unique among the connected clients.
func verifyName(name string) bool {
	for _, client := range clients {
		if client.Name == name {
			return false
		}
	}
	return true
}
