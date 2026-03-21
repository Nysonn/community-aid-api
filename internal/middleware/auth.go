package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"community-aid-api/internal/services"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gin-gonic/gin"
)

const (
	ContextKeyUserID   = "userID"
	ContextKeyClerkID  = "clerkID"
	ContextKeyUserRole = "userRole"
	ContextKeyIsActive = "isActive"
)

func ClerkAuth(userSvc *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed authorization header"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Debug: log a safe prefix of the token so we can confirm it is arriving.
		prefix := token
		if len(prefix) > 20 {
			prefix = prefix[:20]
		}
		log.Printf("ClerkAuth: received token prefix: %s", prefix)

		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{Token: token})
		if err != nil {
			log.Printf("ClerkAuth: verify error: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		log.Printf("ClerkAuth: verified clerkID: %s", claims.Subject)

		user, err := userSvc.GetUserByClerkID(c.Request.Context(), claims.Subject)
		if errors.Is(err, services.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user is not registered"})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to look up user"})
			return
		}

		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "account is deactivated"})
			return
		}

		c.Set(ContextKeyUserID, user.ID)
		c.Set(ContextKeyClerkID, user.ClerkID)
		c.Set(ContextKeyUserRole, user.Role)
		c.Set(ContextKeyIsActive, user.IsActive)
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyUserRole)
		if !exists || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
