package chats

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

var (
	ErrGettingInterests     = errors.New("error getting user's followed brands")
	ErrRegisteredClient     = errors.New("error registering client to hub: already exists")
	ErrConversationNotFound = errors.New("conversation not found")
	ErrUnAuthorisedAccess   = errors.New("unauthorized access to conversation")
	ErrMessageSaveFailed    = errors.New("failed to save message")
	ErrLastMessageUpdate    = errors.New("failed to update conversation last message timestamp")
	ErrMarkReadFailed       = errors.New("failed to mark messages as read")
	ErrMessageDropped       = errors.New("message dropped due to blocked client channel")
	ErrInvalidId            = errors.New("invalid id entered")
)

const (
	MessageTimeout = 10 * time.Second
)

type ServerMessage struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

// Run the hub instance in a separate go routine
func (h *Hub) Run() {
	log.Println("WebSocket Hub started")

	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case req := <-h.processMessage:
			h.handleMessage(req)

		case msg := <-h.broadcast:
			err := h.handleBroadcast(msg)
			if err != nil {
				log.Printf("error in broadcast: %s", err.Error())
			}
		case <-h.stop:
			log.Printf("Stopping the Hub instance...\n")
			h.clientsMU.Lock()
			defer h.clientsMU.Unlock()
			for id, c := range h.clients {
				close(c.Send)
				delete(h.clients, id)
			}
			return
		}
	}
}

// Closes the hub routine
func (h *Hub) Stop() {
	h.stopOnce.Do(func() { close(h.stop) })
}

// Adds the new client connection to the hub
// fetches the client's followed brands from the DB and assigns
// adds the new client to each brand's follower lists
func (h *Hub) handleRegister(client *Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), MessageTimeout)
	defer cancel()
	// client id
	id := client.ID
	h.clientsMU.RLock()
	defer h.clientsMU.RUnlock()
	if _, exist := h.clients[id]; exist {
		log.Printf("Client %s already registered\n", id)
		return ErrRegisteredClient
	}
	// New connection from client
	h.clients[id] = client

	// Load the clients data
	followedBrands, err := h.store.LoadFollowedBrands(ctx, id)
	if err != nil {
		return ErrGettingInterests
	} else {
		client.FollowedBrands = followedBrands
	}
	h.followersMu.Lock()
	for bid := range followedBrands {

		// Initialize brand's follower map if needed
		if h.brandFollowers[bid] == nil {
			h.brandFollowers[bid] = make(map[string]*Client)
		}

		// Add user to brand's followers
		h.brandFollowers[bid][client.ID] = client

		// Track in client for easy removal
		client.FollowedBrands[bid] = true
	}
	h.followersMu.Unlock()
	// Send a welcome message from the server
	resp, err := json.Marshal(&ServerMessage{
		Sender:  "server",
		Message: "Welcome to the CampaignHub server!",
	})
	if err != nil {
		log.Printf("error marshalling server message: %s", err.Error())
		return err
	}
	client.Conn.WriteJSON(resp)
	log.Printf("Client registered: %s\n", id)
	return nil
}

// Cleans up the spaces and channels
// occupied by the disconnected client
func (h *Hub) handleUnregister(client *Client) {
	h.clientsMU.RLock()
	if _, ok := h.clients[client.ID]; ok {
		// delete the client from the hub
		delete(h.clients, client.ID)
		// close the Send channel to stop any further sends
		close(client.Send)
	}
	h.clientsMU.RUnlock()

	// Remove from all brand follower lists
	h.followersMu.Lock()
	for brandID := range client.FollowedBrands {
		if followers, exists := h.brandFollowers[brandID]; exists {
			// Removing the client from the list to
			// prevent null broadcast attempts
			delete(followers, client.ID)

			// Clean up empty follower lists
			if len(followers) == 0 {
				delete(h.brandFollowers, brandID)
			}
		}
	}
	h.followersMu.Unlock()
}

// Routes the message to the appropriate handler based on its type
func (h *Hub) handleMessage(req *MessageRequest) {
	// Process incoming message from client
	ctx, cancel := context.WithTimeout(context.Background(), MessageTimeout)
	defer cancel()
	switch req.Message.Type {
	// Handle 1-to-1 chat message
	case "chat_message":
		h.handleChatMessage(ctx, req)

	// Handle read receipts
	case "mark_read":
		h.handleMarkRead(ctx, req)

	// Handle typing indicator
	case "typing":
		h.handleTyping(ctx, req)

	// Handle follow brand
	case "follow_brand":
		h.handleFollowBrand(req)

	// Handle unfollow brand
	case "unfollow_brand":
		h.handleUnfollowBrand(req)
	default:
		log.Printf("Unknown message type: %s from client: %s", req.Message.Type, req.Client.ID)
	}
}

// 1-to-1 Chat Message
func (h *Hub) handleChatMessage(ctx context.Context, req *MessageRequest) error {
	// Verify access
	conv, err := h.store.GetConversationByID(ctx, req.Message.ConversationID)
	if err != nil {
		log.Printf("Error fetching conversation: %v", err)
		return err
	}

	if conv.ParticipantOne != req.Client.ID && conv.ParticipantTwo != req.Client.ID {
		log.Printf("Unauthorized chat access")
		return ErrUnAuthorisedAccess
	}

	// Save message
	msg := Message{
		ID:             uuid.New().String(),
		ConversationID: req.Message.ConversationID,
		SenderID:       string(req.Client.ID),
		MessageType:    req.Message.MessageType,
		Content:        req.Message.Content,
		IsRead:         false,
	}

	if err := h.store.SaveMessage(ctx, &msg); err != nil {
		log.Printf("Error saving message: %s\n", err.Error())
		return ErrMessageSaveFailed
	}

	// Update conversation timestamp
	if err := h.store.UpdateLastMessageAt(ctx, conv.ID); err != nil {
		log.Printf("Error updating conversation timestamp: %s", err.Error())
		return ErrLastMessageUpdate
	}

	// Determine recipient (1-to-1 routing)
	recipientID := conv.ParticipantTwo
	if conv.ParticipantTwo == req.Client.ID {
		recipientID = conv.ParticipantOne
	}

	msg.CreatedAt = time.Now().String()[:19] // Simplified timestamp

	log.Printf(
		"Message from %s to %s in conversation %s\n",
		req.Client.ID, recipientID, conv.ID,
	)
	// Send to recipient using DIRECT broadcast
	h.broadcast <- &BroadcastMessage{
		Type:   "direct",
		UserID: recipientID,
		Payload: map[string]any{
			"type":    "new_message",
			"message": msg,
		},
	}

	return nil
}

// Read Receipt Handler
func (h *Hub) handleMarkRead(ctx context.Context, req *MessageRequest) error {
	// Verify access
	conv, err := h.store.GetConversationByID(ctx, req.Message.ConversationID)
	if err != nil {
		log.Printf("Error fetching conversation: %v", err)
		return err
	}

	if conv.ParticipantOne != req.Client.ID && conv.ParticipantTwo != req.Client.ID {
		log.Printf("Unauthorized chat access")
		return ErrUnAuthorisedAccess
	}

	// Mark messages as read
	if err := h.store.MarkMessagesAsRead(ctx, req.Message.ConversationID, req.Client.ID); err != nil {
		log.Printf("error marking read: %s", err.Error())
		return ErrMarkReadFailed
	}

	// Notify the other participant
	recipientID := conv.ParticipantTwo
	if conv.ParticipantTwo == req.Client.ID {
		recipientID = conv.ParticipantOne
	}
	h.broadcast <- &BroadcastMessage{
		Type:   "direct",
		UserID: recipientID,
		Payload: map[string]any{
			"type":            "mark_read",
			"conversation_id": conv.ID,
			"reader_id":       req.Client.ID,
		},
	}

	return nil
}

// Typing Indicator Handler
func (h *Hub) handleTyping(ctx context.Context, req *MessageRequest) error {
	// Verify access
	conv, err := h.store.GetConversationByID(ctx, req.Message.ConversationID)
	if err != nil {
		log.Printf("Error fetching conversation: %v", err)
		return err
	}

	if conv.ParticipantOne != req.Client.ID && conv.ParticipantTwo != req.Client.ID {
		log.Printf("Unauthorized chat access")
		return ErrUnAuthorisedAccess
	}

	// Notify the other participant
	recipientID := conv.ParticipantTwo
	if conv.ParticipantTwo == req.Client.ID {
		recipientID = conv.ParticipantOne
	}

	h.broadcast <- &BroadcastMessage{
		Type:   "direct",
		UserID: recipientID,
		Payload: map[string]any{
			"type":            "typing",
			"conversation_id": conv.ID,
			"typer_id":        req.Client.ID,
		},
	}

	return nil
}

// Broadcast Handler
// The broadcast handler currently uses the in-memory list of concurrent users
// TODO: Upgrade to a redis Pub/Sub architecture to handle horizontal scaling of servers
// One brand's broadcast from a server can be Published to a redis channel and the other Subscribed
// Clients can then pull this message concurrently
func (h *Hub) handleBroadcast(msg *BroadcastMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		// Log the error and continue to next follower
		log.Printf("error marshalling data for broadcast: %s\n", err.Error())
		return err
	}
	switch msg.Type {
	case "direct":
		h.clientsMU.RLock()
		recipientClient, exists := h.clients[msg.UserID]
		h.clientsMU.RUnlock()
		if !exists {
			// the message is still saved in the Database so we can fetch for the user
			// when they come online next time
			log.Printf("Recipient client not online: %s", msg.UserID)
			return nil
		}
		select {
		case recipientClient.Send <- data:
			// Message sent successfully
		default:
			// If the send channel is blocked, drop the message broadcast to this client
			// We know the client is online but their channel is blocked
			// So we log it for debugging purposes
			log.Printf("Dropping direct message to client: %s due to blocked channel\n", recipientClient.ID)
			return ErrMessageDropped
		}
		log.Printf("Direct message sent to client: %s\n", msg.UserID)

	// Broadcast to all followers of a brand
	case "followers":
		h.followersMu.RLock()
		followers, exists := h.brandFollowers[msg.BrandID]
		if !exists {
			h.followersMu.RUnlock()
			log.Printf("No followers for brand: %s", msg.BrandID)
			return nil
		}
		// Unlocking the mutex before iterating to avoid holding the lock too long
		h.followersMu.RUnlock()
		// Looping through all the online/active follwoers of the brand
		for _, followerClient := range followers {
			// Send the data over to the send channel of the client
			select {
			case followerClient.Send <- data:
				// Message sent successfully
			default:
				// If the send channel is blocked, drop the message broadcast to this client
				log.Printf("Dropping broadcast to client: %s due to blocked channel\n", followerClient.ID)
			}
		}
		// Broadcast complete
		log.Printf("Broadcasted message to followers of brand: %s\n", msg.BrandID)
	default:
		log.Printf("Unknown message type: %s to client: %s\n", msg.Type, msg.UserID)
	}
	return nil
}

// Handle follow brand
func (h *Hub) handleFollowBrand(req *MessageRequest) {
	exists := req.Message.BrandID
	if exists == nil {
		return
	}
	brandID := *req.Message.BrandID
	h.followersMu.Lock()
	// Initialize brand's follower map if needed
	if h.brandFollowers[brandID] == nil {
		h.brandFollowers[brandID] = make(map[string]*Client)
	}

	// Add user to brand's followers
	h.brandFollowers[brandID][req.Client.ID] = req.Client

	// Track in client for easy removal
	req.Client.FollowedBrands[brandID] = true
	h.followersMu.Unlock()
	log.Printf("Client %s followed brand %s\n", req.Client.ID, brandID)
	err := h.store.FollowBrand(req.Client.ID, brandID)
	if err != nil {
		log.Printf("error writing follow: %s\n", err.Error())
	}
}

// Handle unfollow brand
func (h *Hub) handleUnfollowBrand(req *MessageRequest) {
	exists := req.Message.BrandID
	if exists == nil {
		return
	}
	brandID := *req.Message.BrandID
	h.followersMu.Lock()
	// Check brand's follower map if needed
	if h.brandFollowers[brandID] == nil {
		// Log the invalid request
		log.Printf("error unfollow request for brand: %s, by client: %s\n", brandID, req.Client.ID)
		return
	}

	// Add user to brand's followers
	h.brandFollowers[brandID][req.Client.ID] = req.Client

	// Track in client for easy removal
	req.Client.FollowedBrands[brandID] = true
	h.followersMu.Unlock()
	log.Printf("Client %s followed brand %s\n", req.Client.ID, brandID)

}
