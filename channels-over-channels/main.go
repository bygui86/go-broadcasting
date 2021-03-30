package main

import (
	"fmt"
	"time"
)

func main() {
	// Create the server channel
	serverChan := make(chan chan string, 4)

	// Start the server in its own goroutine
	go server(serverChan)

	// Create the clients channels
	client1Chan := make(chan string, 4)
	client2Chan := make(chan string, 4)
	client3Chan := make(chan string, 4)

	// Push clients channels to server channel to connect the clients
	serverChan <- client1Chan
	serverChan <- client2Chan
	serverChan <- client3Chan

	// time.Sleep(2 * time.Second)

	// Start the clients in their own goroutine, otherwise we would have a deadlock
	go client(1, client1Chan)
	go client(2, client2Chan)
	go client(3, client3Chan)

	// Use a dirty hack to avoid everything ending too early
	fmt.Println("waiting 3 sec before ending main go-routine")
	time.Sleep(3 * time.Second)
}

// The server
// 1. listens endlessly for new clients
// 2. whenever a new client connects
// 		2.1. adds all of them to a list
// 		2.2. sends a text stating the number of total clients to all clients,
func server(serverChan chan chan string) {
	var clients []chan string
	for {
		select {
		case client := <-serverChan:
			clients = append(clients, client)

			// Broadcast a message to all clients
			for _, c := range clients {
				c <- fmt.Sprintf("%d client(s) connected.", len(clients))
			}
		}
	}
}

// The client just listens endlessly to its assigned channel
func client(clientId int, clientChan chan string) {
	fmt.Println(fmt.Sprintf("Client %d started", clientId))
	for {
		text := <-clientChan
		fmt.Println(fmt.Sprintf("Client %d received a message: %s", clientId, text))
	}
}
