package chats

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader websocket.Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking in production
		return true // Allow all origins for simplicity; adjust as needed for security
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
