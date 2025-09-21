package api

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals/auth"
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
	Age           int        `json:"age" binding:"required"`
	PlatformLinks []db.Links `json:"links" binding:"required"`
}

// Just a dummy function that helps in checking if the server is working fine
func (app *Application) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, WriteResponse("Server Running !"))
}

// Account verification route handler
func (app *Application) Verification(c *gin.Context) {
	ctx := c.Request.Context()
	token := c.Query("token")
	payload, err := app.jwtMaker.VerifyToken(token)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			log.Printf("error token expired\n")
			c.JSON(http.StatusUnauthorized, WriteError("Token Expired"))
		} else {
			log.Printf("error token invalid")
			c.JSON(http.StatusUnauthorized, WriteError("Invalid Token"))
		}
		return
	}
	// content has ["type", "id"] : type can be U and B
	content := strings.Split(payload.Sub, " ")
	TokenTyp := content[0]
	TokenEmail := content[1]
	// User verified -> create session -> Redirect to welcome page
	err = app.store.UserInterface.VerifyUser(ctx, TokenTyp, TokenEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("Server Error. Try Again."))
		return
	}
	// --> Create a Session Payload
	SessionToken, err := app.pasetoMaker.CreateToken(app.cfg.ISS,
		app.cfg.AUD, payload.Sub, SessionTimeout,
	)
	if err != nil {
		log.Printf("error generating session for(%v): %v\n", payload.Sub, err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Please login manually"))
		return
	}
	// --> Assign it to the cookie
	c.SetCookie(
		"session",
		SessionToken,
		CookieExp,
		"/",
		"",    // For Development (TODO : Change to domain)
		false, // Secure (HTTPS only)(TODO : Change later)
		true,  // HttpOnly
	)
	// TODO: Redirect the user to Welcome Screen
	c.Redirect(http.StatusFound, "http://localhost:8080/")
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
	tokenSub := "U" + " " + user.Id
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
		Body:    mailer.InviteBody(token),
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
	ID := c.Request.PathValue("id")

	ok := uuid.Validate(ID) // validate the user id
	if ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
	}
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
		Age:           user.Age,
		PlatformLinks: user.PlatformLinks,
	}

	// successfully retreived the user
	c.JSON(http.StatusOK, WriteResponse(userResponse))
}

func (app *Application) GetUserByEmail(c *gin.Context) {
	ctx := c.Request.Context()

	mail := c.Request.PathValue("email")
	// fetching the user
	user, err := app.store.UserInterface.GetUserByEmail(ctx, mail)
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
		Age:           user.Age,
		PlatformLinks: user.PlatformLinks,
	}

	// successfully retreived the user
	c.JSON(http.StatusOK, WriteResponse(userResponse))
}

func (app *Application) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	ID := c.Request.PathValue("id")
	if ok := uuid.Validate(ID); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
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

	// successfully retreived the user
	c.JSON(http.StatusNoContent, WriteResponse("user deleted"))
}

func (app *Application) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Request.PathValue("id")
	var payload db.UpdatePayload
	// Checking for required fields
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError(err.Error()))
		return
	}
	// updating the user
	err := app.store.UserInterface.UpdateUser(ctx, id, payload)
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
	user, err := app.store.UserInterface.GetUserById(ctx, id)
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
		Age:           user.Age,
		PlatformLinks: user.PlatformLinks,
	}

	// successfully retreived the user
	c.JSON(http.StatusOK, WriteResponse(userResponse))
}
