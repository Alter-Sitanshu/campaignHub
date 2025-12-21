package chats

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// waits until a client is registered or times out
func waitForClient(h *Hub, id string, timeout time.Duration) (*Client, bool) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if c, ok := h.clients[id]; ok {
			return c, true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil, false
}

func TestRegister(t *testing.T) {
	testHub := NewHub(MockHubStore, MockCacheService)
	go testHub.Run()
	defer testHub.Stop()
	ClientID := "real1"
	var client *Client
	// test server to mock the connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		client = &Client{
			ID:   ClientID,
			Conn: conn,
			Send: make(chan []byte, 256),
		}
		testHub.register <- client
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()
	time.Sleep(50 * time.Millisecond)
	if _, exists := waitForClient(testHub, ClientID, 100*time.Millisecond); !exists {
		t.Errorf("expected registered user: %s\n", ClientID)
	}

	// Unregister the client from the Hub
	testHub.unregister <- client
	time.Sleep(50 * time.Millisecond)
	if _, exists := testHub.clients[ClientID]; exists {
		t.Errorf("expected unregistered user: %s\n", ClientID)
	}
}

func TestBroadcast(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut)
	defer cancel()

	// Start the Hub GoRoutine
	testHub := NewHub(MockHubStore, MockCacheService)
	go testHub.Run()

	ClientID1 := "real1"
	ClientID2 := "real2"
	// test server to mock the connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "upgrade failed", http.StatusInternalServerError)
			return
		}

	}))
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect client1
	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	testHub.register <- &Client{ID: "real1", Conn: ws1, Send: make(chan []byte, 256)}
	defer ws1.Close()

	// Connect client2
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	testHub.register <- &Client{ID: "real2", Conn: ws2, Send: make(chan []byte, 256)}
	defer ws2.Close()

	time.Sleep(50 * time.Millisecond)
	c1, ok1 := waitForClient(testHub, ClientID1, 100*time.Millisecond)
	c2, ok2 := waitForClient(testHub, ClientID2, 100*time.Millisecond)
	if !ok1 || !ok2 {
		t.Fatalf("expected clients but not found.")
	}

	conv := &Conversation{
		ID:             "mock_conversation_direct",
		ParticipantOne: c1.ID, // the Participant IDs do represent the user id
		ParticipantTwo: c2.ID, // but here using client ID just as a place
		Type:           Direct,
	}

	// start a conversation
	err := testHub.Store.CreateConversation(ctx, conv)
	if err != nil {
		t.Errorf("error could not start conversation: %q", err.Error())
	}
	defer func() {
		ClearMessages(ctx)
		testHub.Store.DeleteConversation(ctx, conv.ID)
	}()

	msgReq := &MessageRequest{
		Client: c1,
		Message: IncomingMessage{
			Type:           "chat_message",
			ConversationID: conv.ID,
			Content:        []byte("Hello world"),
			MessageType:    "txt",
		},
	}

	// wait a bit to finish writing to Database
	time.Sleep(50 * time.Millisecond)
	testHub.processMessage <- msgReq

	// delay
	time.Sleep(50 * time.Millisecond)
	// Expect recipient to receive the message
	select {
	case msg := <-c2.Send:
		t.Logf("hub sent message to recipient: %q", string(msg))
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for hub to push message to recipient client.Send")
	}

	// Expect sender to receive an acknowledgement
	select {
	case ack := <-c1.Send:
		t.Logf("hub sent ack to sender: %q", string(ack))
		// simple sanity check: ack payload should contain "message:ack"
		if !strings.Contains(string(ack), "message:ack") {
			t.Fatalf("expected ack payload, got: %s", string(ack))
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for hub to push ack to sender client.Send")
	}
	testHub.Stop()
}
