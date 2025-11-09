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
		select {
		case user := <-r.register:
			r.mu.Lock()
			r.clients[user] = true
			r.mu.Unlock()
			log.Printf("User %s joined the room.", user.Name)
		case user := <-r.unregister:
			r.mu.Lock()
			if _, ok := r.clients[user]; ok {
				delete(r.clients, user)
				log.Printf("User %s left the room.", user.Name)
			}
			r.mu.Unlock()
		case message := <-r.broadcast:
			r.mu.RLock()
			for client := range r.clients {
				err := client.conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("Error broadcasting to %s: %v. Closing connection.", client.Name, err)
					client.conn.Close()
					delete(r.clients, client)
				}
			}
			r.mu.RUnlock()
		}
	}
}
