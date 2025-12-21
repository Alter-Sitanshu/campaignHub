package chats

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"
)

type HubStore struct {
	db *sql.DB
}

type HubStoreInterface interface {
	// making map[brandID]bool for easy lookup
	LoadFollowedBrands(userID string) (map[string]bool, error)
	GetConversationByID(ctx context.Context, conversationID string) (*Conversation, error)
	GetUserConversations(ctx context.Context, entity, entityID string) ([]ConversationResponse, error)
	MarkConversationClosed(ctx context.Context, conversationID string) error
	DeleteConversation(ctx context.Context, conversationID string) error
	CreateConversation(ctx context.Context, conv *Conversation) error
	SaveMessage(msg *Message) error
	GetConversationMessages(ctx context.Context, date time.Time, cursorSeq, conversationID string, limit int) ([]Message, int64, bool, error)
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

func (hs *HubStore) GetUserConversations(ctx context.Context, entity, entityID string) ([]ConversationResponse, error) {
	var query string

	switch entity {
	case "user":
		// TODO: later modify the database service to separate the direct conversations and
		// campaign conversations. Here assume that the userid will always be in the participant_one
		query = `
			SELECT c.id, participant_two, b.name as participant_name , c.type, c.campaign_id, c.status,
			c.created_at, c.last_message_at, lm.content
			FROM conversations c
			LEFT JOIN brands b ON b.id = c.participant_two
			LEFT JOIN LATERAL (
				SELECT m.content
				FROM messages m
				WHERE m.conversation_id = c.id
				ORDER BY m.created_at DESC
				LIMIT 1
			) lm ON TRUE
			WHERE participant_one = $1 AND c.type = 'campaign'
			ORDER BY c.last_message_at DESC
		`
	case "brand":
		query = `
			SELECT c.id, participant_one, u.first_name as participant_name, c.type, c.campaign_id,
			c.status, c.created_at, c.last_message_at, lm.content
			FROM conversations c
			LEFT JOIN users u ON u.id = c.participant_one
			LEFT JOIN LATERAL (
				SELECT m.content
				FROM messages m
				WHERE m.conversation_id = c.id
				ORDER BY m.created_at DESC
				LIMIT 1
			) lm ON TRUE
			WHERE participant_two = $1 AND c.type = 'campaign'
			ORDER BY c.last_message_at DESC
		`
	}

	rows, err := hs.db.QueryContext(ctx, query, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var output []ConversationResponse
	for rows.Next() {
		var conv ConversationResponse
		err = rows.Scan(
			&conv.ID,
			&conv.ParticipantID,
			&conv.Participant,
			&conv.Type,
			&conv.CampaignID,
			&conv.Status,
			&conv.CreatedAt,
			&conv.LastMessageAt,
			&conv.LastMessage,
		)
		if err != nil {
			return nil, err
		}
		output = append(output, conv)
	}

	return output, nil
}

func (hs *HubStore) GetConversationMessages(ctx context.Context, date time.Time,
	cursorSeq, conversationID string, limit int) ([]MessageResp, int64, bool, error) {
	var query = `
		SELECT id, conversation_id, sender_id, message_type, content, is_read, created_at, seq
		FROM messages
		WHERE conversation_id = $1
	`

	args := []any{conversationID}

	if !date.IsZero() {
		query += `
		AND (
			created_at < $2
			OR (created_at = $2 AND seq < $3)
		)`
		args = append(args, date, cursorSeq)
	}

	query += `
		ORDER BY created_at DESC, seq DESC
		LIMIT $` + strconv.Itoa(len(args)+1)

	args = append(args, limit+1)
	rows, err := hs.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("error messages of conv: %s\n", err.Error())
		return nil, 0, false, err
	}
	defer rows.Close()
	var (
		output                 []MessageResp
		nextCursor, prevCursor int64
	)
	for rows.Next() {
		var msg MessageResp
		prevCursor = nextCursor
		if err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.MessageType,
			&msg.Content,
			&msg.IsRead,
			&msg.CreatedAt,
			&nextCursor,
		); err != nil {
			log.Printf("error scanning: %s\n", err.Error())
			return nil, 0, false, err
		}
		output = append(output, msg)
	}
	HasMore := len(output) > limit
	n := min(limit, len(output))
	if HasMore {
		nextCursor = prevCursor
	}
	return output[:n], nextCursor, HasMore, nil
}

func (hs *HubStore) CreateConversation(ctx context.Context, conv *Conversation) error {
	query := `
		INSERT INTO conversations
		(id, participant_one, participant_two, type, campaign_id)
		VALUES ($1, $2, $3, $4, $5)
	`
	// campaign_id is optional for direct conversations - handle nil safely
	var campaignID any
	if conv.CampaignID != nil {
		campaignID = *conv.CampaignID
	} else {
		campaignID = nil
	}
	_, err := hs.db.ExecContext(ctx, query,
		conv.ID,
		conv.ParticipantOne,
		conv.ParticipantTwo,
		conv.Type,
		campaignID,
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

func (hs *HubStore) MarkConversationClosed(ctx context.Context, ConvID string) error {
	query := `
		UPDATE conversations
		SET status = 'closed'
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
		INSERT INTO messages (client_id, id, conversation_id, sender_id, message_type, content, is_read)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := hs.db.ExecContext(ctx, query,
		msg.ClientID,
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
	hs.UpdateLastMessageAt(ctx, msg.ConversationID)
	return nil
}

func (hs *HubStore) SaveMessages(ctx context.Context, buf []Message) error {
	if len(buf) == 0 {
		return nil
	}
	conv := buf[0].ConversationID
	tx, err := hs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO messages 
		(client_id, id, conversation_id, sender_id, message_type, content, is_read)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, msg := range buf {
		_, err := stmt.ExecContext(
			ctx,
			msg.ClientID,
			msg.ID,
			msg.ConversationID,
			msg.SenderID,
			msg.MessageType,
			msg.Content,
			msg.IsRead,
		)
		if err != nil {
			log.Printf("error saving message %s: %v", msg.ID, err)
			return err
		}
	}

	hs.UpdateLastMessageAt(ctx, conv)
	return tx.Commit()
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
		WHERE conversation_id = $1 AND sender_id = $2 AND is_read = false
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
