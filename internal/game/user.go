package game

import (
	"log"

	"github.com/gofiber/contrib/websocket"
)

type User struct {
	conn *websocket.Conn
	ID   string
	Name string
	game *Game //자신이 속한 게임 방에 관련된 참조
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
		//정상 종료 로직 추가.
		log.Printf("Read loop for client %s ended.", u.ID)
	}()

	for {
		msg, err := u.ReadMessage()
		if err != nil {
			log.Printf("Error reading message for client %s: %v", u.ID, err)
			break
		}
		u.game.HandleMessage(u, msg)
	}
}

func (u *User) ReadMessage() ([]byte, error) {
	_, msg, err := u.conn.ReadMessage()
	if err != nil {
		u.checkNormalClosure(err)
		return nil, err
	}
	log.Printf("Received message from %s: %s", u.Name, string(msg))
	// 메시지 처리 로직 추가
	return msg, nil
}

func (u *User) checkNormalClosure(err error) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		log.Printf("Read error for client %s: %v", u.Name, err)
	}
}
