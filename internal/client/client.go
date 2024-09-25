package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		return
	}
	defer conn.Close()

	// Handle termination signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Communicate between goroutines
	doneChan := make(chan struct{})

	// Goroutine to listen for messages from the server
	go func() {
		reader := bufio.NewReader(conn)
		for {
			select {
			case <-doneChan:
				return // Exit when doneChan is closed
			default:
				message, err := reader.ReadString('\n')
				if err != nil {
					fmt.Printf("Error reading from server: %v\n", err)
					return
				}
				fmt.Print("Server: ", message)

				// If it's a conversation request, handle it
				if strings.HasPrefix(message, "Conversation request from") {
					fmt.Print("Accept conversation (yes/no)? ")
					var response string
					fmt.Scanln(&response)
					if strings.ToLower(response) == "yes" {
						conn.Write([]byte("ok\n"))
					} else {
						conn.Write([]byte("no\n"))
					}
				}
			}
		}
	}()

	// Goroutine to send client input to server
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("Enter message or 'start:clientID' to initiate conversation: ")
			scanner.Scan()
			text := scanner.Text() + "\n"
			conn.Write([]byte(text))

			// If the client types "end", break the loop and end connection
			if strings.TrimSpace(text) == "end" {
				break
			}
		}
		close(doneChan) // Signal the reading goroutine to stop
	}()

	// Goroutine to handle termination signals
	go func() {
		<-signalChan
		fmt.Println("\nClient terminating, notifying server...")
		conn.Write([]byte("Client disconnected\n"))
		close(doneChan)
	}()

	// Block until doneChan is closed
	<-doneChan
	fmt.Println("Client has exited")
}
