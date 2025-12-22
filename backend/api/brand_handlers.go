package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/Alter-Sitanshu/campaignHub/internals/mailer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// User Request Payload
type BrandPaylaod struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Sector   string `json:"sector" binding:"required"`
	Password string `json:"password" binding:"required"`
	Website  string `json:"website" binding:"required"`
	Address  string `json:"address" binding:"required"`
}

type BrandResponse struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required"`
	Sector  string `json:"sector" binding:"required"`
	Website string `json:"website" binding:"required"`
	Address string `json:"address" binding:"required"`
}

func (app *Application) CreateBrand(c *gin.Context) {
	ctx := c.Request.Context()
	var err error        // error for the functions used while creating a brand
	var flag bool = true // flag to check if the brand is created in the db
	var payload BrandPaylaod
	// Checking for required fields
	if err = c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError(err.Error()))
		return
	}
	// Unpacking the payload to a user object
	brand := db.Brand{
		Id:      uuid.New().String(),
		Name:    payload.Name,
		Email:   payload.Email,
		Sector:  payload.Sector,
		Website: payload.Website,
		Address: payload.Address,
	}
	if err := brand.Password.Hash(payload.Password); err != nil {
		log.Printf("could not hash the password: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
		return
	}
	// Clean up before return
	defer func() {
		if err != nil && flag {
			// Encountered an error while Verification
			// Undo the brand creation
			app.store.BrandInterface.DeregisterBrand(ctx, brand.Id)
		}
	}()
	err = app.store.BrandInterface.RegisterBrand(ctx, &brand)
	if err != nil {
		// Internal error
		flag = false // setting flag to false to indicate no creation in DB
		log.Printf("Could not register brand: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
		return
	}
	// Mail the brand to a custom url to verify them
	// Create the brand verification JWT Token
	tokenSub := fmt.Sprintf("br-%s", brand.Id)
	token, err := app.jwtMaker.CreateToken(
		app.cfg.ISS, app.cfg.AUD, tokenSub, time.Hour,
	)
	if err != nil {
		log.Printf("error creating brand JWt: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
		return
	}
	// Mail the brand a custom url to verify them
	InvitationReq := mailer.EmailRequest{
		To:      brand.Email,
		Subject: "Verify your account",
		Body:    mailer.InviteBody(brand.Email, token),
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
		log.Printf("error sending verification to %s: %v\n", brand.Email, err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
		return
	}
	// successfully brand created
	c.JSON(http.StatusCreated, WriteResponse(brand))
}

func (app *Application) GetBrand(c *gin.Context) {
	ctx := c.Request.Context()
	ID := c.Param("brand_id")

	ok := uuid.Validate(ID) // validate the brand id
	if ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
	}
	// fetching the brand
	brand, err := app.store.BrandInterface.GetBrandById(ctx, ID)
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
	brandResponse := BrandResponse{
		Name:    brand.Name,
		Email:   brand.Email,
		Sector:  brand.Sector,
		Website: brand.Website,
		Address: brand.Address,
	}

	// successfully retreived the user
	c.JSON(http.StatusOK, WriteResponse(brandResponse))
}

func (app *Application) DeleteBrand(c *gin.Context) {
	ctx := c.Request.Context()
	// fetch the logged in user
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
	ID := c.Param("brand_id")
	if ok := uuid.Validate(ID); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}
	if ID != Entity.GetID() && Entity.GetRole() != "admin" {
		log.Printf("error: unauthorised access by: %v\n", Entity.GetID())
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised"))
		return
	}
	// deregistering the brand
	err := app.store.BrandInterface.DeregisterBrand(ctx, ID)
	if err != nil {
		if err == db.ErrNotFound {
			// bad request error
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		// server error
		log.Printf("could not find brand: %v", err.Error()) // log to fix the error
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}

	// successfully deregistered brand
	c.JSON(http.StatusNoContent, WriteResponse("brand deregistered"))
}

func (app *Application) UpdateBrand(c *gin.Context) {
	ctx := c.Request.Context()
	// fetching the logged in user
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
	brand_id := c.Param("brand_id")
	if brand_id != Entity.GetID() && Entity.GetRole() != "admin" {
		log.Printf("error: unauthorised access by: %v\n", Entity.GetID())
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised"))
		return
	}
	var payload db.BrandUpdatePayload
	// Checking for required fields
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError(err.Error()))
		return
	}
	// updating the user
	err := app.store.BrandInterface.UpdateBrand(ctx, brand_id, payload)
	if err != nil {
		if err == db.ErrNotFound {
			// bad request error
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		// server error
		log.Printf("could not find brand: %v", err.Error()) // log to fix the error
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// fetching the updated brand
	brand, err := app.store.BrandInterface.GetBrandById(ctx, brand_id)
	if err != nil {
		if err == db.ErrNotFound {
			// bad request error
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		// server error
		log.Printf("could not find brand: %v", err.Error()) // log to fix the error
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
	}
	// make the response object
	brandResponse := BrandResponse{
		Name:    brand.Name,
		Email:   brand.Email,
		Sector:  brand.Sector,
		Website: brand.Website,
		Address: brand.Address,
	}

	// successfully retreived the user
	c.JSON(http.StatusOK, WriteResponse(brandResponse))
}
