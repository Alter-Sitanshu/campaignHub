package api

import (
	"log"
	"net/http"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// User Request Payload
type UserPaylaod struct {
	Id            string     `json:"id" binding:"required"`
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

func (app *Application) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()
	var payload UserPaylaod
	// Checking for required fields
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError(err.Error()))
		return
	}
	// Unpacking the payload to a user object
	user := db.User{
		Id:            payload.Id,
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
	err := app.store.UserInterface.CreateUser(ctx, &user)
	if err != nil {
		// Internal error
		log.Printf("Could not create user: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
		return
	}

	// successfully user created
	c.JSON(http.StatusCreated, WriteResponse(&user))
}

func (app *Application) GetUser(c *gin.Context) {
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
	c.JSON(http.StatusOK, userResponse)
}
