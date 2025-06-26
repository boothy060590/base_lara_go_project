package middlewares

import (
	"base_lara_go_project/app/utils/token"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RBAC middleware: pass allowed roles as arguments
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := token.IsTokenValid(c)
		if err != nil {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		userRole, err := token.ExtractTokenRole(c)
		if err != nil {
			c.String(http.StatusForbidden, "Forbidden: role not found")
			c.Abort()
			return
		}
		allowed := false
		for _, r := range roles {
			if userRole == r {
				allowed = true
				break
			}
		}
		if !allowed {
			c.String(http.StatusForbidden, "Forbidden: insufficient role")
			c.Abort()
			return
		}
		c.Next()
	}
}
