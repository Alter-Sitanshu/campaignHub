package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AccountPayload struct {
	HolderId string  `json:"holder_id" binding:"required"`
	Type     string  `json:"type" binding:"required,oneof=creator brand"` // either creator or brand
	Amount   float64 `json:"amount"`
}

func (app *Application) CreateAccount(c *gin.Context) {
	ctx := c.Request.Context()
	var payload AccountPayload
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError(err.Error()))
		return
	}
	// validate the holder id provided first
	if ok := uuid.Validate(payload.HolderId); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}

	formattedTime := time.Now().Format("2006-01-02 15:04:05-07:00")
	acc := db.Account{
		Id:       uuid.New().String(),
		HolderId: payload.HolderId,
		Type:     payload.Type,
		Amount:   payload.Amount,
	}
	// open the account for the user
	err := app.store.TransactionInterface.OpenAccount(ctx, &acc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError(err.Error()))
		return
	}
	// add the created at time and return
	acc.CreatedAt = formattedTime
	c.JSON(http.StatusCreated, WriteResponse(acc))
}

func (app *Application) GetUserAccount(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	User, ok := LogInUser.(*db.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	acc_id := c.Param("acc_id")
	if ok := uuid.Validate(acc_id); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}

	// fetch the user account details
	acc, err := app.store.TransactionInterface.GetAccount(ctx, acc_id)
	if err != nil {
		if err == db.ErrNotFound {
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// check if the user is the owner of the account
	if acc.HolderId != User.Id && User.Role != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorized request"))
		return
	}

	// return acc details to the user
	c.JSON(http.StatusOK, WriteResponse(acc))
}

func (app *Application) DisableUserAccount(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	User, ok := LogInUser.(*db.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	acc_id := c.Param("acc_id")
	// validate the acc_id uuid
	if ok := uuid.Validate(acc_id); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}

	// fetch the user account details
	acc, err := app.store.TransactionInterface.GetAccount(ctx, acc_id)
	if err != nil {
		if err == db.ErrNotFound {
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// check if the user is the owner of the account
	if acc.HolderId != User.Id && User.Role != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorized request"))
		return
	}

	// disable the user account
	err = app.store.TransactionInterface.DisableAccount(ctx, acc_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}

	// successfully disabled account
	c.JSON(http.StatusNoContent, WriteResponse("account disabled"))
}

func (app *Application) DeleteUserAccount(c *gin.Context) {
	ctx := c.Request.Context()
	// fetch the current user
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	User, ok := LogInUser.(*db.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	acc_id := c.Param("acc_id")
	// validate the acc_id uuid
	if ok := uuid.Validate(acc_id); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}
	// fetch the user account details
	acc, err := app.store.TransactionInterface.GetAccount(ctx, acc_id)
	if err != nil {
		if err == db.ErrNotFound {
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// check if the user is the owner of the account
	if acc.HolderId != User.Id && User.Role != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorized request"))
		return
	}

	// disable the user account
	err = app.store.TransactionInterface.DeleteAccount(ctx, acc_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}

	// successfully deleted account
	c.JSON(http.StatusNoContent, WriteResponse("account deleted"))
}

func (app *Application) GetAllAccounts(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	User, ok := LogInUser.(*db.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	if User.Role != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorized request"))
		return
	}
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
	accounts, err := app.store.TransactionInterface.GetAllAccounts(ctx, limit, offset)
	if err != nil {
		// internal error
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}

	// sucessful in fetching accounts
	c.JSON(http.StatusOK, WriteResponse(accounts))
}
