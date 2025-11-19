package game

import (
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type Room struct {
	clients    map[*User]bool
	broadcast  chan []byte
	register   chan *User
	unregister chan *User
	mu         sync.RWMutex
}

func NewRoom() *Room {
	return &Room{
		clients:    make(map[*User]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *User),
		unregister: make(chan *User),
	}
}

func (r *Room) Run() {
	for {
		r.handleConnection()
	}
}

func (r *Room) handleConnection() {
	select {
	case user := <-r.register:
		r.handleRegister(user)
	case user := <-r.unregister:
		r.handleUnregister(user)
	case message := <-r.broadcast:
		r.broadcastMessage(message)
	}
}

func (r *Room) handleRegister(user *User) {
	r.mu.Lock()
	r.clients[user] = true
	r.mu.Unlock()
	log.Printf("User %s joined the room.", user.Name)
}

func (r *Room) handleUnregister(user *User) {
	r.mu.Lock()
	if _, ok := r.clients[user]; ok {
		delete(r.clients, user)
		log.Printf("User %s left the room.", user.Name)
	}
	r.mu.Unlock()
}

func (r *Room) broadcastMessage(message []byte) {
	r.mu.RLock()
	clients := make([]*User, 0, len(r.clients))
	for client := range r.clients {
		clients = append(clients, client)
	}
	r.mu.RUnlock()

	for _, client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Error broadcasting to %s: %v. Closing connection.", client.Name, err)
			client.Close()
			r.unregister <- client
		}
	}
}
