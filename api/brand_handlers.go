package api

import (
	"log"
	"net/http"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
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
	var payload BrandPaylaod
	// Checking for required fields
	if err := c.ShouldBindJSON(&payload); err != nil {
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
	// TODO -> Mail the brand to a custom url to verify them
	// --> Create a Session Payload
	SessionToken, err := app.pasetoMaker.CreateToken(app.cfg.ISS,
		app.cfg.AUD, brand.Email, SessionTimeout,
	)
	if err != nil {
		log.Printf("error generating session for(%v): %v", brand.Email, err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
		return
	}
	err = app.store.BrandInterface.RegisterBrand(ctx, &brand)
	if err != nil {
		// Internal error
		log.Printf("Could not register brand: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
		return
	}
	// --> Return the session token/ Assign it to the cookie
	c.SetCookie(
		"session",
		SessionToken,
		CookieExp,
		"/",
		"",    // For Development (TODO : Change to domain)
		false, // Secure (HTTPS only)(TODO : Change later)
		true,  // HttpOnly
	)
	// successfully brand created
	c.JSON(http.StatusCreated, WriteResponse(brand))
}

func (app *Application) GetBrand(c *gin.Context) {
	ctx := c.Request.Context()
	ID := c.Request.PathValue("brand_id")

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
	ID := c.Request.PathValue("brand_id")
	if ok := uuid.Validate(ID); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}
	// deleting the user
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
	brand_id := c.Request.PathValue("brand_id")
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
