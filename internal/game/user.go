package game

import (
	"errors"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type User struct {
	conn      *websocket.Conn
	ID        string
	Name      string
	game      *Game
	mu        sync.RWMutex
	closeOnce sync.Once
}

func NewUser(conn *websocket.Conn, ID string, Name string) *User {
	return &User{
		conn: conn,
		ID:   ID,
		Name: Name,
	}
}

func (u *User) ReadLoop() {
	defer func() {
		log.Printf("Read loop for client %s ended.", u.ID)
		u.Close()
	}()

	for {
		msg, err := u.ReadMessage()
		if err != nil {
			log.Printf("Error reading message for client %s: %v", u.ID, err)
			break
		}
		if u.game != nil {
			u.game.HandleMessage(u, msg)
		}
	}
}

func (u *User) ReadMessage() ([]byte, error) {
	conn := u.getConn()
	if conn == nil {
		return nil, errors.New("connection is closed")
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		u.checkNormalClosure(err)
		return nil, err
	}
	log.Printf("Received message from %s: %s", u.Name, string(msg))
	return msg, nil
}

func (u *User) WriteMessage(messageType int, data []byte) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.conn == nil {
		return errors.New("connection is closed")
	}
	return u.conn.WriteMessage(messageType, data)
}

func (u *User) Close() {
	u.closeOnce.Do(func() {
		u.mu.Lock()
		if u.conn != nil {
			_ = u.conn.Close()
			u.conn = nil
		}
		u.mu.Unlock()
	})
}

func (u *User) getConn() *websocket.Conn {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.conn
}

func (u *User) checkNormalClosure(err error) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		log.Printf("Read error for client %s: %v", u.Name, err)
	}
}
