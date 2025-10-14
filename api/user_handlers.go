package api

import (
	"log"
	"net/http"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals/cache"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/Alter-Sitanshu/campaignHub/internals/mailer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// User Request Payload
type UserPaylaod struct {
	FirstName     string     `json:"first_name" binding:"required"`
	LastName      string     `json:"last_name" binding:"required"`
	Email         string     `json:"email" binding:"required"`
	Password      string     `json:"password" binding:"required"`
	Gender        string     `json:"gender" binding:"required"`
	Age           int        `json:"age" binding:"required"`
	PlatformLinks []db.Links `json:"links" binding:"required"`
}

type UserResponse struct {
	Id            string     `json:"id" binding:"required"`
	FirstName     string     `json:"first_name" binding:"required"`
	LastName      string     `json:"last_name" binding:"required"`
	Email         string     `json:"email" binding:"required"`
	Gender        string     `json:"gender" binding:"required"`
	Amount        float64    `json:"amount" binding:"required,min=0"`
	Age           int        `json:"age" binding:"required"`
	PlatformLinks []db.Links `json:"links" binding:"required"`
}

func (app *Application) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()
	var err error        // error for the functions used while creating a user
	var flag bool = true // flag to check if the user is created in the db
	var payload UserPaylaod
	// Checking for required fields
	if err = c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError(err.Error()))
		return
	}
	// Unpacking the payload to a user object
	user := db.User{
		Id:            uuid.New().String(),
		FirstName:     payload.FirstName,
		LastName:      payload.LastName,
		Email:         payload.Email,
		Gender:        payload.Gender,
		Age:           payload.Age,
		Role:          defaultUserLVL,
		PlatformLinks: payload.PlatformLinks,
	}
	if err := user.Password.Hash(payload.Password); err != nil {
		log.Printf("could not hash the password: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
		return
	}
	// Clean up before return
	defer func() {
		if err != nil && flag {
			// Encountered an error while Verification
			// Undo the user creation
			app.store.UserInterface.DeleteUser(ctx, user.Id)
		}
	}()
	// First create the User in the DB
	err = app.store.UserInterface.CreateUser(ctx, &user)
	if err != nil {
		flag = false // setting the flag false to indicate no creation in DB
		switch {
		case err == db.ErrDupliName:
			c.JSON(http.StatusInternalServerError, WriteError("name already taken"))
			return
		case err == db.ErrDupliMail:
			c.JSON(http.StatusInternalServerError, WriteError("email already exists"))
		default:
			c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
		}
		// Internal error
		log.Printf("Could not create user: %v\n", err.Error())
		return
	}
	// Create the user verification JWT Token
	tokenSub := user.Id
	token, err := app.jwtMaker.CreateToken(
		app.cfg.ISS, app.cfg.AUD, tokenSub, time.Hour,
	)
	if err != nil {
		log.Printf("error creating user JWt: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
		return
	}
	// Mail the user to a custom url to verify them
	InvitationReq := mailer.EmailRequest{
		To:      user.Email,
		Subject: "Verify your account",
		Body:    mailer.InviteBody(user.Email, "users", token),
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
		log.Printf("error sending verification to %s: %v\n", user.Email, err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
		return
	}

	// successfully user created with is_verified = false
	c.JSON(http.StatusCreated, WriteResponse(user))
}

func (app *Application) GetUserById(c *gin.Context) {
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
	ID := c.Param("id")
	if ok := uuid.Validate(ID); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}

	var profile cache.UserResponse
	// pass the pointer to the response var so that the UnMarshal works
	err := app.cache.GetUserProfile(ctx, ID, &profile)
	// Cache hit
	if err == nil {
		c.JSON(http.StatusOK, WriteResponse(profile))
		return
	}

	// In case of a cache miss
	// fetching the user
	user, err := app.store.UserInterface.GetUserById(ctx, ID)
	if err != nil {
		if err == db.ErrNotFound {
			// bad request error
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		// server error
		log.Printf("could not find user: %v", err.Error()) // log to fix the error
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// make the response object
	userResponse := UserResponse{
		Id:            user.Id,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		Gender:        user.Gender,
		Amount:        user.Amount,
		Age:           user.Age,
		PlatformLinks: user.PlatformLinks,
	}

	// set the user profile in the cache
	err = app.cache.SetUserProfile(ctx, ID, cache.UserResponse(userResponse))
	if err != nil {
		log.Printf("error caching the user profile: %s\n", err.Error())
	}
	// successfully retreived the user
	c.JSON(http.StatusOK, WriteResponse(userResponse))
}

func (app *Application) GetUserByEmail(c *gin.Context) {
	ctx := c.Request.Context()
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
	mail := c.Param("email")

	var profile cache.UserResponse
	err := app.cache.GetUserProfileByMail(ctx, mail, &profile)
	// Cache hit
	if err == nil {
		c.JSON(http.StatusOK, WriteResponse(profile))
		return
	}

	// In case of Cache Miss
	user, err := app.store.UserInterface.GetUserByEmail(ctx, mail)
	if err != nil {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// make the response object
	userResponse := UserResponse{
		Id:            user.Id,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		Gender:        user.Gender,
		Amount:        user.Amount,
		Age:           user.Age,
		PlatformLinks: user.PlatformLinks,
	}

	// Set the user profile in cache
	err = app.cache.SetUserProfileByMail(ctx, mail, profile)
	if err != nil {
		log.Printf("error caching user profile: %s\n", err.Error())
	}

	// successfully retreived the user
	c.JSON(http.StatusOK, WriteResponse(userResponse))
}

func (app *Application) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	// get the ucrrent logged in user
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
	UserID := Entity.GetID()
	ID := c.Param("id")
	if ok := uuid.Validate(ID); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}
	// Authorise the user
	if ID != UserID {
		log.Printf("unauthorised access from: %v\n", UserID)
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// deleting the user
	err := app.store.UserInterface.DeleteUser(ctx, ID)
	if err != nil {
		if err == db.ErrNotFound {
			// bad request error
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		// server error
		log.Printf("could not find user: %v", err.Error()) // log to fix the error
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}

	// invalidate the user cache
	err = app.cache.InvalidateUserProfile(ctx, ID)
	if err != nil {
		log.Printf("error invalidating the user cache: %s\n", err.Error())
	}

	app.cache.InvalidateCreatorSubmissions(ctx, ID)
	app.cache.InvalidateUserProfile(ctx, ID)

	// successfully retreived the user
	c.JSON(http.StatusNoContent, WriteResponse("user deleted"))
}

func (app *Application) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()
	// Fetch the logged in user
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
	UserID := Entity.GetID()

	ID := c.Param("id")
	if ok := uuid.Validate(ID); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}
	// Authorise the user
	if ID != UserID {
		log.Printf("unauthorised access from: %v\n", LogInUser)
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	var payload db.UpdatePayload
	// Checking for required fields
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError(err.Error()))
		return
	}
	// updating the user
	err := app.store.UserInterface.UpdateUser(ctx, ID, payload)
	if err != nil {
		if err == db.ErrNotFound {
			// bad request error
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		// server error
		log.Printf("could not find user: %v", err.Error()) // log to fix the error
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// fetching the updated user
	user, err := app.store.UserInterface.GetUserById(ctx, ID)
	if err != nil {
		if err == db.ErrNotFound {
			// bad request error
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		// server error
		log.Printf("could not find user: %v", err.Error()) // log to fix the error
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
	}
	// make the response object
	userResponse := UserResponse{
		Id:            user.Id,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		Gender:        user.Gender,
		Amount:        user.Amount,
		Age:           user.Age,
		PlatformLinks: user.PlatformLinks,
	}

	// cache the user profile
	err = app.cache.SetUserProfile(ctx, ID, cache.UserResponse(userResponse))
	if err != nil {
		log.Printf("error caching user profile: %s\n", err.Error())
	}
	// I dont neeed to set the user balance separately as the profile already
	// has that information in it.

	// successfully retreived the user
	c.JSON(http.StatusOK, WriteResponse(userResponse))
}
