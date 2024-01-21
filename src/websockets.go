package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var clients = make(map[string]*websocket.Conn) // connected clients
var broadcast = make(chan []byte)              // broadcast channel

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func websocketConnectionHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	clientID := getUniqueID(r)
	clients[clientID] = conn

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(clients, clientID)
			return
		}
		// send the new received msg from the client to the broadcast channel
		broadcast <- message
	}
}

// func sendMessageToClient(clientID string, message []byte) {
// 	conn, ok := clients[clientID]
// 	if ok {
// 		err := conn.WriteMessage(websocket.TextMessage, message)
// 		if err != nil {
// 			log.Println(err)
// 			conn.Close()
// 			delete(clients, clientID)
// 		}
// 	} else {
// 		log.Printf("Client %s not found", clientID)
// 	}
// }

func handleMessages() {
	for {
		// grab the next message from the broadcast channel
		msg := <-broadcast
		// send it out to every client that is currently connected
		for clientID, clientConn := range clients {
			err := clientConn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println(err)
				clientConn.Close()
				delete(clients, clientID)
			}
		}
	}
}

func getUniqueID(r *http.Request) string {
	// if the client sent a Sec-Websocket-Key header, use that as the unique ID
	if (r.Header.Get("Sec-Websocket-Key")) != "" {
		return r.Header.Get("Sec-Websocket-Key")
	}
	// otherwise, generate a new unique ID
	return uuid.New().String()
}

var words = []string{"apple", "banana", "cherry", "date", "elderberry", "fig", "grape", "honeydew", "imbe", "jackfruit", "kiwi", "lemon", "mango", "nectarine", "orange", "papaya", "quince", "raspberry", "strawberry", "tangerine", "ugli", "vanilla", "watermelon", "xigua", "yuzu", "zucchini"}

func broadcastRandomWord() {
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			// log.Println("Broadcasting random word to ws...")
			randomWord := words[rand.Intn(len(words))]
			broadcast <- []byte(randomWord)
		}
	}
}
