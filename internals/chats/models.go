package chats

import (
	"database/sql"
	"sync"

	"github.com/Alter-Sitanshu/campaignHub/internals/cache"
	"github.com/gorilla/websocket"
)

const (
	TextMessageType = 1
)

type Client struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	UserRole string // "user", "brand", "admin"

	// Brands followed by the user
	FollowedBrands map[string]bool
	// Additional fields can be added as needed
}

type MessageRequest struct {
	Client  *Client
	Message IncomingMessage
}

type Message struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	MessageType    string `json:"message_type"`
	Content        any    `json:"content"`
	IsRead         bool   `json:"is_read"`
	CreatedAt      string `json:"created_at"`
}

type IncomingMessage struct {
	// Type is to check if the message is for a conversation or a campaign specific chat
	Type           string `json:"type"`
	ConversationID string `json:"conversation_id,omitempty"`
	Content        any    `json:"content,omitempty"`
	// TODO: ADD  Videos support later
	// type is set for to support PDFs, Images, Text. (NO VIDEOS YET)
	MessageType string `json:"message_type,omitempty"`

	// This will used only for the user-brand chat feature after a campaign
	// application by the user gets accepted by the brand.
	BrandID *string `json:"brand_id,omitempty"` // For follow/unfollow
}

type Conversation struct {
	ID             string `json:"id"`
	ParticipantOne string `json:"participant_one"`
	ParticipantTwo string `json:"participant_two"`
	// type can be normal or campaign_connected
	Type string `json:"type"`

	// campaign_connected: means this conversation was created because of a campaign
	// application by the user which got accepted by the brand.
	// In this case, we can have additional logic like:
	CampaignID string `json:"campaign_id,omitempty"` // ID of the campaign if type is campaign_connected

	// make the conversation dead after the campaign expires/ends if campaign connected
	// status can be active or dead
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	LastMessageAt string `json:"last_message_at"`
}

type BroadcastMessage struct {
	Type    string // "direct" or "followers"
	UserID  string // For direct messages
	BrandID string // For brand broadcasts
	Payload any
}

type Hub struct {
	// WHO'S ONLINE: UserID → Connection
	clients   map[string]*Client
	clientsMU sync.RWMutex
	// Example: clients["user-123"].Conn = websocketConnection

	// WHO FOLLOWS WHOM: string → Set of Followers
	brandFollowers map[string]map[string]*Client
	followersMu    sync.RWMutex
	// Example: brandFollowers["nike-id"]["creator-1"] = client pointer / error

	// Message Queue Channels
	register       chan *Client           // New connections
	unregister     chan *Client           // Disconnections
	processMessage chan *MessageRequest   // Incoming messages
	broadcast      chan *BroadcastMessage // Outgoing messages
	store          *HubStore              // Database store
	cache          *cache.Service         // Simple REDIS instance
}

func NewHub(db *sql.DB, appCache *cache.Service) *Hub {
	return &Hub{
		clients:        make(map[string]*Client),
		brandFollowers: make(map[string]map[string]*Client),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		processMessage: make(chan *MessageRequest, 256),
		broadcast:      make(chan *BroadcastMessage, 256),
		store:          &HubStore{db: db},
		cache:          appCache,
	}
}
