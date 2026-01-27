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
	Type     string  `json:"type" binding:"required,oneof=user brand"` // either user or brand
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency" binding:"required,oneof=inr yen usd"`
}

type TxPayload struct {
	To       string  `json:"to_id" binding:"required"`
	From     string  `json:"from_id" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gte=0"`
	Currency string  `json:"currency" binding:"required,oneof= inr usd yen"`
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

	formattedTime := time.Now().Format(time.RFC3339)
	acc := db.Account{
		Id:       uuid.New().String(),
		HolderId: payload.HolderId,
		Type:     payload.Type,
		Amount:   payload.Amount,
		Currency: payload.Currency,
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
	// ONly admin can delete any account details of user
	// check if the user is the owner of the account
	if User.Role != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorized request"))
		return
	}
	// fetch the user account details
	_, err := app.store.TransactionInterface.GetAccount(ctx, acc_id)
	if err != nil {
		if err == db.ErrNotFound {
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
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

func (app *Application) WithdrawBalance(c *gin.Context) {
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
	var payload TxPayload
	if err := c.ShouldBindBodyWithJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError(err.Error()))
		return
	}
	// check if the user is the owner of the account
	if payload.From != User.Id || payload.To != User.Id {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorized request"))
		return
	}
	accountID, err := app.store.TransactionInterface.GetAccountID(ctx, payload.From)
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("account deactivated/not avaialable"))
		return
	}
	tx := db.Transaction{
		Id:       uuid.NewString(),
		FromId:   accountID,
		ToId:     accountID,
		Currency: payload.Currency,
		Amount:   payload.Amount,
		Type:     "withdraw",
	}
	err = app.store.TransactionInterface.Withdraw(ctx, &tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError(err.Error()))
		return
	}
	// update the cache
	app.cache.UpdateUserBalance(ctx, User.GetID(), -payload.Amount)
	c.JSON(http.StatusOK, WriteResponse("Withdraw successful."))

}

func (app *Application) DepositBalance(c *gin.Context) {
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
	var payload TxPayload
	if err := c.ShouldBindBodyWithJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError(err.Error()))
		return
	}
	// check if the user is the owner of the account
	if payload.From != User.Id || payload.To != User.Id {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorized request"))
		return
	}

	accountID, err := app.store.TransactionInterface.GetAccountID(ctx, payload.From)
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("account disabled/unavailable"))
		return
	}

	tx := db.Transaction{
		Id:       uuid.NewString(),
		FromId:   accountID,
		ToId:     accountID,
		Currency: payload.Currency,
		Amount:   payload.Amount,
		Type:     "deposit",
	}
	err = app.store.TransactionInterface.Deposit(ctx, &tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError(err.Error()))
		return
	}

	app.cache.UpdateUserBalance(ctx, User.GetID(), payload.Amount)
	c.JSON(http.StatusOK, WriteResponse("Deposit successful."))
}
