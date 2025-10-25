package chats

import (
	"context"
	"database/sql"
	"errors"
	"log"
)

type HubStore struct {
	db *sql.DB
}

type HubStoreInterface interface {
	// making map[brandID]bool for easy lookup
	LoadFollowedBrands(userID string) (map[string]bool, error)
	GetConversationByID(ctx context.Context, conversationID string) (*Conversation, error)
	DeleteConversation(ctx context.Context, conversationID string) error
	CreateConversation(ctx context.Context, conv *Conversation) error
	SaveMessage(msg *Message) error
	GetConversationMessages(ctx context.Context, conversationID string, offset, limit int) ([]Message, error)
	UpdateLastMessageAt(ctx context.Context, conversationID string) error
	MarkMessagesAsRead(ctx context.Context, conversationID string, userID string) error
	UnfollowBrand(ctx context.Context, user, brand string) error
	FollowBrand(ctx context.Context, user, brand string) error
}

// making map[brandID]bool for easy lookup
func (hs *HubStore) LoadFollowedBrands(ctx context.Context, userID string) (map[string]bool, error) {
	query := `
		SELECT brand_id FROM following_list WHERE user_id = $1
	`
	var output = make(map[string]bool)
	rows, err := hs.db.QueryContext(ctx, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return output, nil
		}
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
	if conversationID == "" {
		return nil, ErrInvalidId
	}
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

func (hs *HubStore) GetConversationMessages(ctx context.Context, conversationID string, offset, limit int) ([]Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, message_type, content, is_read, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := hs.db.QueryContext(ctx, query, conversationID, limit, offset)
	if err != nil {
		log.Printf("error messages of conv: %s\n", err.Error())
		return nil, err
	}
	defer rows.Close()
	var output []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.MessageType,
			&msg.Content,
			&msg.IsRead,
			&msg.CreatedAt,
		); err != nil {
			log.Printf("error scanning: %s\n", err.Error())
			return nil, err
		}
		output = append(output, msg)
	}

	return output, nil
}

func (hs *HubStore) CreateConversation(ctx context.Context, conv *Conversation) error {
	query := `
		INSERT INTO conversations
		(id, participant_one, participant_two, type, campaign_id)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := hs.db.ExecContext(ctx, query,
		conv.ID,
		conv.ParticipantOne,
		conv.ParticipantTwo,
		conv.Type,
		conv.CampaignID,
	)
	if err != nil {
		log.Printf("error entering conversation: %s", err.Error())
		return err
	}
	return nil
}

func (hs *HubStore) DeleteConversation(ctx context.Context, ConvID string) error {
	query := `
		DELETE FROM conversations
		WHERE id = $1
	`
	rows, err := hs.db.ExecContext(ctx, query, ConvID)
	if count, _ := rows.RowsAffected(); count == 0 {
		return nil
	}
	if err != nil {
		log.Printf("error: %s", err.Error())
		return err
	}
	return nil
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

func (hs *HubStore) FollowBrand(ctx context.Context, user, brand string) error {
	query := `
		INSERT INTO following_list (user_id, brand_id)
		VALUES ($1, $2)
	`
	_, err := hs.db.Exec(query, user, brand)
	if err != nil {
		return err
	}

	return nil
}

func (hs *HubStore) UnfollowBrand(ctx context.Context, user, brand string) error {
	query := `
		DELETE FROM following_list
		WHERE user_id = $1 AND brand_id = $2
	`
	_, err := hs.db.Exec(query, user, brand)
	if err != nil {
		return err
	}

	return nil
}
