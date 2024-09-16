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
	clients      = make(map[net.Conn]string) // Track clients with a unique identifier
	clientsMutex sync.Mutex                  // Mutex to protect concurrent access to clients map
)

func main() {
	// Listen on TCP port 8080
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Println("Server is running on port 8080...")

	// Handle incoming client connections
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}

			// Assign a name/ID to the client (e.g., use the remote address)
			clientID := conn.RemoteAddr().String()

			// Add client to the list of connected clients
			clientsMutex.Lock()
			clients[conn] = clientID
			clientsMutex.Unlock()

			fmt.Printf("Client %s connected\n", clientID)
			printConnectedClients()

			// Start a goroutine to handle communication with the client
			go handleConnection(conn)
		}
	}()

	// Server input: to allow the server to send messages to clients
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Send a broadcast message or target a client (clientID):")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text()) // Trim spaces

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
				if id == input {
					targetConn = conn
					break
				}
			}
			clientsMutex.Unlock()

			if targetConn != nil {
				fmt.Println("Enter message for", input, ":")
				scanner.Scan()
				message := scanner.Text() + "\n"
				sendMessageToClient(targetConn, message)
			} else {
				fmt.Println("Invalid client ID or client not found.")
			}
		}
	}
}

// Function to handle individual client connections
func handleConnection(conn net.Conn) {
	defer func() {
		// Remove client from the list of connected clients upon disconnection
		clientsMutex.Lock()
		clientID := clients[conn]
		delete(clients, conn)
		clientsMutex.Unlock()

		fmt.Printf("Client %s disconnected\n", clientID)
		printConnectedClients()

		conn.Close()
	}()

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from client or client disconnected:", err)
			return
		}

		// Lock mutex to safely retrieve the client's ID
		clientsMutex.Lock()
		clientID := clients[conn]
		clientsMutex.Unlock()

		// Print the client's message along with their ID
		fmt.Printf("Message from client %s: %s", clientID, message)

		// If the client sends "end", close the connection for this client
		if strings.TrimSpace(message) == "end" {
			fmt.Printf("Client %s has requested to disconnect.\n", clientID)
			return
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
		fmt.Println("Error sending message to client:", err)
		return
	}
	writer.Flush() // Ensure the message is sent immediately
}

// Function to print the list of connected clients
func printConnectedClients() {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	fmt.Println("Connected clients:")
	for _, clientID := range clients {
		fmt.Println(" -", clientID)
	}
}
