package api

import (
	"log"
	"net/http"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/gin-gonic/gin"
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
		Role:          "LVL1", // TODO: Make this make sense later
		PlatformLinks: payload.PlatformLinks,
	}
	if err := user.Password.Hash(payload.Password); err != nil {
		log.Printf("could not hash the password: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
	}
	err := app.store.UserInterface.CreateUser(ctx, &user)
	if err != nil {
		// Internal error
		log.Printf("Could not create user: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
	}

	// successfully user created
	c.JSON(http.StatusCreated, WriteResponse(&user))
}
