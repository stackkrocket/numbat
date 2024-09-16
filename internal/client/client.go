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
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Channel to handle termination signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel to communicate between goroutines
	doneChan := make(chan struct{})

	// Goroutine to handle server responses
	go func() {
		for {
			select {
			case <-doneChan:
				return // Exit the loop when doneChan is closed
			default:
				message, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					fmt.Println("Error reading from server or server disconnected:", err)
					return
				}
				fmt.Print("Server response: ", message)
			}
		}
	}()

	// Goroutine to send client input to server
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text() + "\n"
			conn.Write([]byte(text))

			// If client types "end", break the loop to close the connection
			if strings.TrimSpace(scanner.Text()) == "end" {
				break
			}
		}
		close(doneChan) // Signal the reading loop to stop
	}()

	// Goroutine to handle termination signal
	go func() {
		<-signalChan
		fmt.Println("\nClient terminating, sending disconnect signal to server.")
		conn.Write([]byte("Client disconnected\n"))
		close(doneChan) // Signal the reading loop to stop
	}()

	// Block until doneChan is closed
	<-doneChan
	fmt.Println("Client has exited.")
}
