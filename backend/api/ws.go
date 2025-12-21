package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Alter-Sitanshu/campaignHub/internals/chats"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WSMessage is the generic envelope we receive from clients over the websocket
type WSMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// WebSocketHandler upgrades the HTTP connection to a websocket, registers the client
// with the Hub, and starts the reader/writer pumps.
func (app *Application) WebSocketHandler(c *gin.Context) {
	payload, exists := c.Get("user")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, WriteError("unauthorised request"))
	}
	user, ok := payload.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// Upgrade to websocket
	conn, err := chats.UpgradeToWebSocket(c.Writer, c.Request)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, WriteError("error: failed to upgrade websocket"))
		return
	}

	// Create client and register with hub
	client := &chats.Client{
		ID:             user.GetID(),
		Conn:           conn,
		Send:           make(chan []byte, 256),
		FollowedBrands: make(map[string]bool),
	}
	app.msgHub.Register(client)

	// Start writer pump in a goroutine
	go wsWriter(client)
	// Run reader in this goroutine, so the HTTP handler stays alive until read ends
	wsReader(app, client)
}

// wsReader reads messages from the websocket, unmarshals them, and forwards
// them to the Hub for processing.
func wsReader(app *Application, client *chats.Client) {
	defer func() {
		// Ensure we unregister and close the connection on exit
		app.msgHub.Unregister(client)
		client.Conn.Close()
	}()

	for {
		var msg WSMessage
		if err := client.Conn.ReadJSON(&msg); err != nil {
			// read error usually means client disconnected
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("socket read error (unexpected): %v", err)
			} else {
				log.Printf("socket read closed: %v", err)
			}
			return
		}

		// Forward the payload to the Hub as an IncomingMessage. The Hub handlers will
		// inspect IncomingMessage.Type to route to the correct action.
		var incoming chats.IncomingMessage
		if err := json.Unmarshal(msg.Data, &incoming); err != nil {
			log.Printf("invalid message payload from client %s: %v", client.ID, err)
			continue
		}
		// Push into hub's processing queue
		app.msgHub.SubmitMessage(&chats.MessageRequest{Client: client, Message: incoming})
	}
}

// wsWriter writes payloads from the Hub to the client's websocket.
func wsWriter(client *chats.Client) {
	for payload := range client.Send {
		if err := client.Conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("error writing to client %s: %v", client.ID, err)
			break
		}
	}
	// When the send channel is closed, ensure underlying connection is closed
	client.Conn.Close()
}
