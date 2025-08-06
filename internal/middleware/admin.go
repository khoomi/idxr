package middleware

import (
	"net/http"

	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
)

// AdminOnly middleware restricts access to Super and Mod users only
func AdminOnly(userService services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authenticated user session
		session, err := auth.GetSessionAuto(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}

		// Get current user to check their role
		currentUser, err := userService.GetUserByID(c.Request.Context(), session.UserId)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}

		// Check if user has admin privileges (Super or Mod role)
		if currentUser.Role != models.Super && currentUser.Role != models.Mod {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions: admin access required",
			})
			c.Abort()
			return
		}

		// Continue to next handler
		c.Next()
	}
}