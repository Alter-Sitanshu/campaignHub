package chats

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader websocket.Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "https://frogmedia-tawny.vercel.app"
	},
}

func UpgradeToWebSocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading connection to socket: %s", err.Error())
		return nil, err
	}

	// conncetion upgraded successfully
	return conn, nil
}
