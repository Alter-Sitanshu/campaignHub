package api

import (
	"net/http"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// AuthMiddleware ensures that the user is authenticated before accessing protected routes
func (app *Application) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionToken, err := c.Cookie("session")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, WriteError("error: session token missing"))
			return
		}
		payload, err := app.pasetoMaker.VerifyToken(sessionToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, WriteError("error: invalid session token"))
			return
		}
		// set the user/brand in the context for further handlers
		user, err := app.store.GetEntity(c.Request.Context(), payload.Sub)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, WriteError("server error"))
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func (app *Application) AuthoriseUser() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		ID := c.Param("user_id")
		// Authorise the user
		if ID != UserID {
			c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
			return
		}

		c.Next()
	}
}

func (app *Application) AuthoriseBrand() gin.HandlerFunc {
	return func(c *gin.Context) {
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
		if Entity.GetEntityType() != "brands" {
			c.JSON(http.StatusForbidden, WriteError("method/operation forbidden"))
			return
		}

		ID := c.Param("brand_id")
		// Authorise the user
		if ID != UserID {
			c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
			return
		}

		c.Next()
	}
}

func (app *Application) AuthoriseAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Fetching the logged in user
		LogInUser, ok := c.Get("user")
		if !ok {
			c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
			return
		}
		user, ok := LogInUser.(db.AuthenticatedEntity)
		if !ok {
			c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
			return
		}
		if user.GetRole() != "admin" {
			c.JSON(http.StatusForbidden, WriteError("forbidden method or operation"))
			return
		}

		c.Next()
	}
}

// Rate limiting middleware checks and moderates the number
// of requests alloed per second to hit the endpoints
func (app *Application) RateLimitter(limiter *rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the request is allowed. Allow() consumes a token.
		if !limiter.Allow() {
			// If not allowed, abort the request with a "Too Many Requests" status
			c.AbortWithStatusJSON(http.StatusTooManyRequests, WriteError("too many requests"))
			return
		}
		// If allowed, proceed to the next handler
		c.Next()
	}
}
