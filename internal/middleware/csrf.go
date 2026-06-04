package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/private-tf-runners/server/internal/models"
)

func CSRFProtect() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			token = c.Request.Header.Get("X-CSRF-Token")
		}

		cookie, err := c.Cookie("csrf_token")
		if err != nil || token == "" || cookie == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{
				Error: "CSRF token required",
				Code:  "CSRF_REQUIRED",
			})
			return
		}

		expectedToken, err := decodeCSRFToken(cookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Invalid CSRF token",
				Code:  "CSRF_INVALID",
			})
			return
		}

		if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Invalid CSRF token",
				Code:  "CSRF_INVALID",
			})
			return
		}

		c.Next()
	}
}

func decodeCSRFToken(encoded string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
