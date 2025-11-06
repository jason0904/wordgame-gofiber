package main

import (
	"log"

	"github.com/gofiber/contrib/websocket"
)

type User struct {
	conn *websocket.Conn
	ID   string
	game *Game //자신이 속한 게임 방에 관련된 참조
}

func NewUser(conn *websocket.Conn, ID string) *User {
	return &User{
		conn: conn,
		ID:   ID,
	}
}

func (u *User) readLoop() {
	defer func() {
		//정상 종료 로직 추가.
	}

	for {
		_, msg, err := u.conn.ReadMessage()
		if err != nil {
			//정상적으로 종료되었는지 체크
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error for client %s: %v", u.name, err)
			}
			break
		}
		log.Printf("Received message from %s: %s", u.name, string(msg))
		// 메시지 처리 로직 추가
		u.game.handleMessage(u, msg)
	}
}