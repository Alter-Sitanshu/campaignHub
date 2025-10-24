package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/Alter-Sitanshu/campaignHub/internals/mailer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Support Tickets model
type TicketPayload struct {
	CustomerId string `json:"customer_id" binding:"required"`              // id of the entity
	Type       string `json:"type" binding:"required,oneof=creator brand"` // creator or brand whoever raised a ticket
	Subject    string `json:"subject" binding:"required"`
	Message    string `json:"message" binding:"required"`
}

func (app *Application) RaiseTicket(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	Entity, ok := LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	var payload TicketPayload
	// validate the ticket request sent
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError(err.Error()))
		return
	}
	// check if the user is raising ticket for themselves
	if Entity.GetID() != payload.CustomerId {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// make a ticket object out of the payload
	ticket := db.Ticket{
		Id:         uuid.New().String(),
		CustomerId: payload.CustomerId,
		Type:       payload.Type,
		Subject:    payload.Subject,
		Message:    payload.Message,
		Status:     db.OpenTicket,
	}
	// raise the ticket
	err := app.store.TicketInterface.OpenTicket(ctx, &ticket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// Notify the admin mail
	InvitationReq := mailer.EmailRequest{
		To:      app.cfg.MailCfg.Support,
		Subject: "Verify your account",
		Body: mailer.GenerateTicketEmail(
			app.cfg.MailCfg.Support,
			ticket,
		),
	}
	// Implementing a retry fallback
	tries := 1
	for tries <= app.cfg.MailCfg.MailRetries {
		err = app.mailer.PushMail(InvitationReq)
		if err == nil {
			break
		}
		tries++
	}
	if err != nil && tries > app.cfg.MailCfg.MailRetries {
		log.Printf("error sending verification to %s: %v\n", "support", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
		return
	}

	// successfully raised the ticket
	c.JSON(http.StatusCreated, WriteResponse(ticket))
}

func (app *Application) CloseTicket(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	Entity, ok := LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	ticket_id := c.Param("ticket_id")
	// validating the uuid
	if ok := uuid.Validate(ticket_id); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}
	if Entity.GetRole() != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// resolve the ticket
	if err := app.store.TicketInterface.ResolveTicket(ctx, ticket_id); err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// successfully resolved the ticket
	c.JSON(http.StatusNoContent, WriteResponse("ticket resolved"))
}

func (app *Application) GetRecentTickets(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	Entity, ok := LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// Authorise the user
	if Entity.GetRole() != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	status_param := c.Query("status")
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid query"))
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}
	var status int
	// verify the query
	switch status_param {
	case "open":
		status = db.OpenTicket
	case "close":
		status = db.CloseTicket
	default:
		c.JSON(http.StatusBadRequest, WriteError("invalid query"))
		return
	}
	// DB call
	tickets, err := app.store.TicketInterface.GetRecentTickets(ctx, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// successfully fetched tickets
	c.JSON(http.StatusOK, WriteResponse(tickets))
}

func (app *Application) DeleteTicket(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	Entity, ok := LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// Authorise the user
	if Entity.GetRole() != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	ticket_id := c.Param("ticket_id")
	// validating the uuid
	if ok := uuid.Validate(ticket_id); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}
	if err := app.store.TicketInterface.DeleteTicket(ctx, ticket_id); err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}

	// successfully deleted ticket
	c.JSON(http.StatusNoContent, WriteResponse("ticket deleted"))
}

func (app *Application) GetTicket(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	Entity, ok := LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	ticket_id := c.Param("ticket_id")
	// validating the uuid
	if ok := uuid.Validate(ticket_id); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}
	// fetching the ticket
	ticket, err := app.store.TicketInterface.FindTicket(ctx, ticket_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// authorise the user
	if ticket.CustomerId != Entity.GetID() {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// fetched ticket
	c.JSON(http.StatusOK, WriteResponse(ticket))
}
