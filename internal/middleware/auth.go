package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/private-tf-runners/server/internal/config"
	"github.com/private-tf-runners/server/internal/database"
	"github.com/private-tf-runners/server/internal/models"
)

type RunnerAuthMiddleware struct {
	db *database.DB
}

func NewRunnerAuthMiddleware(db *database.DB) *RunnerAuthMiddleware {
	return &RunnerAuthMiddleware{db: db}
}

func (m *RunnerAuthMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Authorization header required",
				Code:  "UNAUTHORIZED",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid authorization header format",
				Code:  "INVALID_AUTH_FORMAT",
			})
			return
		}

		token := parts[1]
		runner, err := m.db.Runner().GetByToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid runner token",
				Code:  "INVALID_TOKEN",
			})
			return
		}

		c.Set("runner", runner)
		c.Set("runner_id", runner.ID)
		c.Next()
	}
}

func (m *RunnerAuthMiddleware) GetRunner(c *gin.Context) *models.Runner {
	if runner, exists := c.Get("runner"); exists {
		return runner.(*models.Runner)
	}
	return nil
}

const (
	ContextUserKey    = "user"
	ContextClaimsKey  = "claims"
	ContextRequestID  = "request_id"
)

func GetUser(c *gin.Context) *models.User {
	if user, exists := c.Get(ContextUserKey); exists {
		return user.(*models.User)
	}
	return nil
}

func GetClaims(c *gin.Context) *models.TokenClaims {
	if claims, exists := c.Get(ContextClaimsKey); exists {
		return claims.(*models.TokenClaims)
	}
	return nil
}

func RequirePermission(permission models.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Unauthorized",
				Code:  "UNAUTHORIZED",
			})
			return
		}

		role := models.Role(claims.Role)
		if !role.HasPermission(permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Insufficient permissions",
				Code:  "FORBIDDEN",
			})
			return
		}

		c.Next()
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Authorization header required",
				Code:  "UNAUTHORIZED",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid authorization header format",
				Code:  "INVALID_AUTH_FORMAT",
			})
			return
		}

		tokenString := parts[1]
		claims, err := ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid or expired token",
				Code:  "INVALID_TOKEN",
			})
			return
		}

		if claims.Type != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid token type",
				Code:  "INVALID_TOKEN_TYPE",
			})
			return
		}

		c.Set(ContextClaimsKey, claims)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil || claims.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Admin access required",
				Code:  "FORBIDDEN",
			})
			return
		}
		c.Next()
	}
}

func GenerateAccessToken(user *models.User) (string, int64, error) {
	cfg := config.Get()
	expiresAt := time.Now().Add(cfg.Security.JWTExpiration)

	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"type":     "access",
		"iat":      time.Now().Unix(),
		"exp":      expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(cfg.Security.JWTSecret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

func GenerateRefreshToken(user *models.User) (string, error) {
	cfg := config.Get()
	expiresAt := time.Now().Add(cfg.Security.RefreshExpiration)

	claims := jwt.MapClaims{
		"sub":  user.ID,
		"type": "refresh",
		"iat":  time.Now().Unix(),
		"exp":  expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(cfg.Security.JWTSecret)
}

func ValidateToken(tokenString string) (*models.TokenClaims, error) {
	cfg := config.Get()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		if token.Method.Alg() != "HS512" {
			return nil, jwt.ErrSignatureInvalid
		}
		return cfg.Security.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	claims := &models.TokenClaims{
		UserID:   mapClaims["sub"].(string),
		Username: mapClaims["username"].(string),
		Email:    mapClaims["email"].(string),
		Role:     mapClaims["role"].(string),
		Type:     mapClaims["type"].(string),
	}

	return claims, nil
}

func ValidateRefreshToken(tokenString string) (string, error) {
	cfg := config.Get()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return cfg.Security.JWTSecret, nil
	})

	if err != nil {
		return "", err
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || mapClaims["type"] != "refresh" {
		return "", jwt.ErrTokenInvalidClaims
	}

	userID, ok := mapClaims["sub"].(string)
	if !ok {
		return "", jwt.ErrTokenInvalidClaims
	}

	return userID, nil
}
