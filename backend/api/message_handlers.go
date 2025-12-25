package api

import (
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
	output, next, hasMore, err := app.msgHub.Store.GetConversationMessages(ctx, date, lastSeq, conversationID, MessageLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("could not load messages"))
		return
	}
	if len(output) == 0 || output == nil {
		c.JSON(http.StatusNoContent, WriteError("No messages Found"))
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
