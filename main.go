package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Define our message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

var port = "3000"

var clients = make(map[*websocket.Conn]bool) // connected clients

var broadcast = make(chan Message) // broadcast channel

// Configure the upgrader

var upgrader = websocket.Upgrader{}

func main() {
	fileServer := http.FileServer(http.Dir("./public/"))

	http.Handle("/", fileServer)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages()

	fmt.Println("Application Started At http://127.0.0.1:" + port)

	HandleErrors(http.ListenAndServe(":"+port, nil))
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast

		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)

			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket

	ws, err := upgrader.Upgrade(w, r, nil)

	HandleErrors(err)

	// Register our new client
	clients[ws] = true

	for {
		var msg Message

		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)

		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}

		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}

	// Make sure we close the connection when the function returns
	defer ws.Close()
}

// Handling Errors.
func HandleErrors(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
