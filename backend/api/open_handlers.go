package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Alter-Sitanshu/campaignHub/internals/auth"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/Alter-Sitanshu/campaignHub/internals/mailer"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Just a dummy function that helps in checking if the server is working fine
func (app *Application) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, WriteResponse("Server Running !"))
}

// Account verification route handler
func (app *Application) Verification(c *gin.Context) {
	ctx := c.Request.Context()
	// extract the token from the query
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
	entity := payload.Sub[:2]
	TokenID := payload.Sub[3:]
	// User verified -> create session -> Redirect to welcome page
	err = app.store.UserInterface.VerifyUser(ctx, entity, TokenID)
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
	c.JSON(http.StatusOK, "OK")
}

// Login handler sets the SessionToken as `us-UUID` or `br-UUID`
func (app *Application) Login(c *gin.Context) {
	ctx := c.Request.Context()
	var payload struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Entity   string `json:"entity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}
	var SessionToken string
	type resp struct {
		Id       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	var response resp
	switch payload.Entity {
	case "users":
		user, err := app.store.UserInterface.GetUserByEmail(ctx, payload.Email)
		if err != nil {
			if err == db.ErrNotFound {
				// bad request error
				c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
				return
			}
			// server error
			log.Printf("could not find user: %v", err.Error()) // log to fix the error
			c.JSON(http.StatusInternalServerError, WriteError("server error. try again"))
			return
		}
		err = user.Password.Compare(payload.Password)
		if err != nil {
			if errors.Is(err, db.ErrInvalidPass) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
				return
			}
			// server error
			log.Printf("could not compare password: %v", err.Error()) // log to fix the error
			c.JSON(http.StatusInternalServerError, WriteError("server error"))
			return
		}
		if !user.IsVerified {
			c.JSON(http.StatusUnauthorized, WriteError("please verify your email to login"))
			return
		}
		// --> Create a Session Payload
		SessionToken, err = app.pasetoMaker.CreateToken(app.cfg.ISS,
			app.cfg.AUD, fmt.Sprintf("us-%s", user.Id), SessionTimeout,
		)
		if err != nil {
			log.Printf("error generating session for(%v): %v\n", payload.Email, err.Error())
			c.JSON(http.StatusInternalServerError, WriteError("try again"))
			return
		}
		response = resp{
			Id:       user.Id,
			Username: user.FirstName,
			Email:    user.Email,
		}
	case "brands":
		brand, err := app.store.BrandInterface.GetBrandByEmail(ctx, payload.Email)
		if err != nil {
			if err == db.ErrNotFound {
				// bad request error
				c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
				return
			}
			// server error
			log.Printf("could not find brand: %v", err.Error()) // log to fix the error
			c.JSON(http.StatusInternalServerError, WriteError("server error. try again"))
			return
		}
		err = brand.Password.Compare(payload.Password)
		if err != nil {
			if errors.Is(err, db.ErrInvalidPass) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
				return
			}
			// server error
			log.Printf("could not compare password: %v", err.Error()) // log to fix the error
			c.JSON(http.StatusInternalServerError, WriteError("server error"))
			return
		}
		if !brand.IsVerified {
			c.JSON(http.StatusUnauthorized, WriteError("please verify your email to login"))
			return
		}
		// --> Create a Session Payload
		SessionToken, err = app.pasetoMaker.CreateToken(app.cfg.ISS,
			app.cfg.AUD, fmt.Sprintf("br-%s", brand.Id), SessionTimeout,
		)
		if err != nil {
			log.Printf("error generating session for(%v): %v\n", payload.Email, err.Error())
			c.JSON(http.StatusInternalServerError, WriteError("try again"))
			return
		}
		response = resp{
			Id:       brand.Id,
			Username: brand.Name,
			Email:    brand.Email,
		}
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
	c.JSON(http.StatusOK, WriteResponse(response))
}

func (app *Application) ForgotPassword(c *gin.Context) {
	ctx := c.Request.Context()
	entity := c.Param("entity")
	// check if the entity parameter is valid
	if entity != "users" && entity != "brands" {
		c.JSON(http.StatusUnauthorized, WriteError("bad request"))
		return
	}
	var payload struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError("bad request"))
		return
	}
	// Send the email a password reset link
	user, err := app.store.UserInterface.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		c.JSON(http.StatusOK, WriteResponse("If the email is registered, you will receive a password reset link"))
		return
	}
	// generate a token for the user
	token, err := app.jwtMaker.CreateToken(
		app.cfg.ISS,
		app.cfg.AUD,
		user.Id,
		ResetTokenExpiry,
	)
	if err != nil {
		log.Printf("error generating reset token for(%v): %v\n", user.Email, err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// Send the email to the user with the token
	InvitationReq := mailer.EmailRequest{
		To:      user.Email,
		Subject: "Reset your password",
		Body:    mailer.GeneratePasswordResetEmail(user.Email, token),
	}
	// Implementing a retry fallback
	// Asynchronous mailing to avoid locking the user flow
	go func() {
		tries := 1
		for tries <= app.cfg.MailCfg.MailRetries {
			err = app.mailer.PushMail(InvitationReq)
			if err == nil {
				break
			}
			tries++
		}
		if err != nil && tries > app.cfg.MailCfg.MailRetries {
			log.Printf("error sending password reset to %s: %v\n", user.Email, err.Error())
			c.JSON(http.StatusInternalServerError, WriteError("Server Error"))
			return
		}
	}()
	// Successfully sent the mail
	c.JSON(http.StatusOK, WriteResponse("If the email is registered, you will receive a password reset link"))
}

func (app *Application) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()
	entity := c.Param("entity")
	// check if the entity parameter is valid
	if entity != "users" && entity != "brands" {
		c.JSON(http.StatusUnauthorized, WriteError("bad request"))
		return
	}
	// extract the token from the query
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
	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid request"))
		return
	}
	if entity == "users" {
		err := app.store.UserInterface.ChangePassword(ctx, payload.Sub, req.NewPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, WriteError("server error"))
			return
		}
	} else {
		err := app.store.BrandInterface.ChangePassword(ctx, payload.Sub, req.NewPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, WriteError("server error"))
			return
		}
	}

	// successfully changed the password
	c.JSON(http.StatusOK, WriteResponse("password changed successfully"))
}

// func (app *Application) AdminLogin(c *gin.Context) {

// }
