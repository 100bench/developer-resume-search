package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AuthRequired is a middleware to check if the user is authenticated
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("userID")
		if userID == nil {
			// User is not authenticated, redirect to login page
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		// User is authenticated, proceed to the next handler
		c.Next()
	}
}
