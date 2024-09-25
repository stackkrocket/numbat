package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	clients      = make(map[net.Conn]string)    // Track clients with their remote address
	clientsMutex sync.Mutex                     // Protect concurrent access to clients map
	pendingReqs  = make(map[string]net.Conn)    // Track pending conversation requests
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Println("Server is running on port 8080...")

	// Goroutine to accept new client connections
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}

			clientID := conn.RemoteAddr().String()
			clientsMutex.Lock()
			clients[conn] = clientID
			clientsMutex.Unlock()

			fmt.Printf("Client %s connected\n", clientID)
			printConnectedClients()

			// Notify all clients about the new connection
			broadcastClientList()

			// Handle each client connection in a new goroutine
			go handleConnection(conn)
		}
	}()

	// Allow server to send messages or target specific clients
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Send a broadcast message or target a client (clientID):")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "broadcast" {
			fmt.Println("Enter broadcast message:")
			scanner.Scan()
			message := scanner.Text() + "\n"
			broadcastMessage(message)
		} else {
			// Target specific client by clientID
			clientsMutex.Lock()
			var targetConn net.Conn
			for conn, id := range clients {
				if strings.HasPrefix(id, input) { // Allow partial match (IP only)
					targetConn = conn
					break
				}
			}
			clientsMutex.Unlock()

			if targetConn != nil {
				fmt.Printf("Enter message for %s: ", input)
				scanner.Scan()
				message := scanner.Text() + "\n"
				sendMessageToClient(targetConn, message)
			} else {
				fmt.Println("Invalid client ID or client not found.")
			}
		}
	}
}

// Function to handle communication with individual clients
func handleConnection(conn net.Conn) {
	defer func() {
		clientsMutex.Lock()
		clientID := clients[conn]
		delete(clients, conn)
		clientsMutex.Unlock()

		fmt.Printf("Client %s disconnected\n", clientID)
		printConnectedClients()

		// Notify all clients about the disconnection
		broadcastClientList()

		conn.Close()
	}()

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading from client: %v\n", err)
			return
		}

		clientsMutex.Lock()
		clientID := clients[conn]
		clientsMutex.Unlock()

		if strings.HasPrefix(message, "start:") {
			targetID := strings.TrimSpace(strings.Split(message, ":")[1])

			// Find target client
			clientsMutex.Lock()
			var targetConn net.Conn
			for conn, id := range clients {
				if strings.HasPrefix(id, targetID) { // Allow partial match (IP only)
					targetConn = conn
					break
				}
			}
			clientsMutex.Unlock()

			if targetConn != nil {
				sendMessageToClient(targetConn, "Conversation request from "+clientID+"\n")
				pendingReqs[clientID] = targetConn
			} else {
				sendMessageToClient(conn, "Client "+targetID+" not found\n")
			}

		} else if strings.TrimSpace(message) == "ok" && pendingReqs[clientID] != nil {
			targetConn := pendingReqs[clientID]
			sendMessageToClient(targetConn, "Client "+clientID+" accepted the conversation\n")
			delete(pendingReqs, clientID)

		} else if strings.TrimSpace(message) == "no" && pendingReqs[clientID] != nil {
			targetConn := pendingReqs[clientID]
			sendMessageToClient(targetConn, "Client "+clientID+" rejected the conversation\n")
			delete(pendingReqs, clientID)

		} else {
			fmt.Printf("Message from client %s: %s", clientID, message)

			if strings.TrimSpace(message) == "end" {
				fmt.Printf("Client %s requested to disconnect.\n", clientID)
				return
			}
		}
	}
}

// Function to broadcast a message to all connected clients
func broadcastMessage(message string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for conn := range clients {
		sendMessageToClient(conn, message)
	}
}

// Function to send a message to a specific client
func sendMessageToClient(conn net.Conn, message string) {
	writer := bufio.NewWriter(conn)
	_, err := writer.WriteString(message)
	if err != nil {
		fmt.Printf("Error sending message to client: %v\n", err)
		return
	}
	writer.Flush()
}

// Function to print all connected clients
func printConnectedClients() {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	fmt.Println("Connected clients:")
	for _, clientID := range clients {
		fmt.Println(" -", clientID)
	}
}

// Function to broadcast the list of connected clients to all clients
func broadcastClientList() {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	clientList := "Connected clients:\n"
	for _, clientID := range clients {
		clientList += " - " + clientID + "\n"
	}

	for conn := range clients {
		sendMessageToClient(conn, clientList)
	}
}
