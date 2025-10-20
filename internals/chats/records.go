package chats

import (
	"context"
	"database/sql"
	"log"
)

type HubStore struct {
	db *sql.DB
}

type HubStoreInterface interface {
	// making map[brandID]bool for easy lookup
	LoadFollowedBrands(userID string) (map[string]bool, error)
	GetConversationByID(conversationID string) (*Conversation, error)
	SaveMessage(msg *Message) error
	UpdateLastMessageAt(ctx context.Context, conversationID string) error
	MarkMessagesAsRead(ctx context.Context, conversationID string, userID string) error
}

// making map[brandID]bool for easy lookup
func (hs *HubStore) LoadFollowedBrands(ctx context.Context, userID string) (map[string]bool, error) {
	query := `
		SELECT brand_id FROM following_list WHERE user_id = $1
	`
	var output = make(map[string]bool)
	rows, err := hs.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("error getting bands: %s", err.Error())
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var brand string
		if err := rows.Scan(&brand); err != nil {
			log.Printf("error scanning brand id: %s", err.Error())
			return nil, err
		}

		output[brand] = true
	}

	return output, nil
}

func (hs *HubStore) GetConversationByID(ctx context.Context, conversationID string) (*Conversation, error) {
	query := `
		SELECT id, participant_one, participant_two, type, campaign_id, status, created_at,
		last_message_at
		FROM conversations
		WHERE id = $1
	`
	var conv Conversation
	err := hs.db.QueryRowContext(ctx, query, conversationID).Scan(
		&conv.ID,
		&conv.ParticipantOne,
		&conv.ParticipantTwo,
		&conv.Type,
		&conv.CampaignID,
		&conv.Status,
		&conv.CreatedAt,
		&conv.LastMessageAt,
	)
	if err != nil {
		log.Printf("error fetching conversation by ID: %s", err.Error())
		return nil, err
	}
	return &conv, nil
}

func (hs *HubStore) SaveMessage(ctx context.Context, msg *Message) error {
	query := `
		INSERT INTO messages (id, conversation_id, sender_id, message_type, content, is_read)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := hs.db.ExecContext(ctx, query,
		msg.ID,
		msg.ConversationID,
		msg.SenderID,
		msg.MessageType,
		msg.Content,
		msg.IsRead,
	)
	if err != nil {
		log.Printf("error saving message: %s", err.Error())
		return err
	}
	// message saved successfully
	return nil
}

func (hs *HubStore) UpdateLastMessageAt(ctx context.Context, conversationID string) error {
	query := `
		UPDATE conversations
		SET last_message_at = NOW()
		WHERE id = $1
	`
	_, err := hs.db.ExecContext(ctx, query, conversationID)
	if err != nil {
		log.Printf("error updating last_message_at: %s", err.Error())
		return err
	}
	return nil
}

// marks all the messages in a conversation as read
// once the user opens the chat
func (hs *HubStore) MarkMessagesAsRead(ctx context.Context, conversationID string, userID string) error {
	query := `
		UPDATE messages
		SET is_read = TRUE
		WHERE conversation_id = $1 AND sender_id = $2 AND is_read = FALSE
	`
	_, err := hs.db.ExecContext(ctx, query, conversationID, userID)
	if err != nil {
		log.Printf("error marking messages as read: %s", err.Error())
		return err
	}
	return nil
}
