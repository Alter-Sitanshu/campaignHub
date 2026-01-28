package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals/chats"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/gin-gonic/gin"
)

const (
	MessageLimit = 25
)

type MessageMeta struct {
	Meta
	Timestamp string `json:"timestamp"`
}

type MessageResponse struct {
	Messages []chats.MessageResp `json:"messages"`
	Meta     MessageMeta         `json:"meta"`
}

const (
	timeOut = time.Second * 5
)

// handles what to do of the campaign conversation life-cycle
// if the status given is to Activate -> Opens a new campaign
// if the status is to End a campaign Cycle -> Invalidates the conversation
// Participant One is always the user and Two is the brand of the campaign
func (app *Application) handleCampaignConversation(
	conv *chats.Conversation, newStatus int,
) error {
	if conv.CampaignID == nil {
		return fmt.Errorf("campaign id is required for conversation")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	switch newStatus {
	case db.AcceptedStatus:
		// the accepted status is an application status (1 for accepted)
		// when the brand accepts an application a new conversation is registered
		return app.msgHub.Store.CreateConversation(ctx, conv)
	case db.ExpiredStatus:
		// the expired status is a campaign status (3 for expired/ended)
		// when the campaign decides to end a campaign, the conversation closes
		return app.msgHub.Store.MarkCampaignConversationClosed(ctx, *conv.CampaignID)
	}

	return nil
}

func (app *Application) GetEntityConversations(c *gin.Context) {
	ctx := c.Request.Context()
	// Fetching the logged in user
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	user, ok := LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}

	userConversations, err := app.msgHub.Store.GetUserConversations(ctx, string(user.GetEntityType()), user.GetID())
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("failed to load conversations"))
		return
	}
	c.JSON(http.StatusOK, WriteResponse(userConversations))
}

func (app *Application) GetConversationMessages(c *gin.Context) {
	ctx := c.Request.Context()
	// Fetching the logged in user
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	_, ok = LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	conversationID := c.Param("conversation")
	timestamp := c.Query("timestamp")
	cursor := c.Query("cursor")
	var lastSeq string
	if cursor != "" {
		seqBytes, err := base64.RawStdEncoding.DecodeString(cursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, WriteError("bad request parameters"))
		}
		lastSeq = string(seqBytes)
	}
	var (
		date time.Time
		err  error
	)
	if timestamp != "" {
		date, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			c.JSON(http.StatusBadRequest, WriteError("timestamp invalid"))
			return
		}
	}
	// first check if there are conversations in the message buffer of the conversation
	// if there are, flush them and then fetch the messages from the next cursor
	// this will prevent droppend/un-flushed messages or ghost messages
	app.msgHub.SyncMessages(conversationID)
	output, next, hasMore, err := app.msgHub.Store.GetConversationMessages(ctx, date, lastSeq, conversationID, MessageLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("could not load messages"))
		return
	}
	if len(output) == 0 || output == nil {
		// just return an empty object for the UI
		c.JSON(http.StatusOK, WriteResponse(MessageResponse{}))
		return
	}
	var buf []byte
	c.JSON(http.StatusOK, WriteResponse(MessageResponse{
		Messages: output,
		Meta: MessageMeta{
			Meta: Meta{
				HasMore: hasMore,
				Cursor:  base64.RawStdEncoding.EncodeToString(fmt.Appendf(buf, "%d", next)),
			},
			Timestamp: output[len(output)-1].CreatedAt.Format(time.RFC3339),
		},
	}))
}
