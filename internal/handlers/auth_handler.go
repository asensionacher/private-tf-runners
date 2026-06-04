package handlers

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/private-tf-runners/server/internal/config"
	"github.com/private-tf-runners/server/internal/database"
	"github.com/private-tf-runners/server/internal/middleware"
	"github.com/private-tf-runners/server/internal/models"
)

type AuthHandler struct {
	db *database.DB
}

func NewAuthHandler(db *database.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	user, err := h.db.User().GetByUsername(req.Username)
	if err != nil {
		h.logFailedLogin(c, req.Username, "user_not_found")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Invalid credentials",
			Code:  "INVALID_CREDENTIALS",
		})
		return
	}

	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		h.logFailedLogin(c, req.Username, "account_locked")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Account temporarily locked. Please try again later.",
			Code:  "ACCOUNT_LOCKED",
		})
		return
	}

	cfg := config.Get()
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		attempts := user.LoginAttempts + 1
		var lockedUntil *time.Time
		if attempts >= cfg.Security.MaxLoginAttempts {
			t := time.Now().Add(cfg.Security.LockoutDuration)
			lockedUntil = &t
		}
		h.db.User().UpdateLoginAttempts(user.ID, attempts, lockedUntil)
		h.logFailedLogin(c, req.Username, "invalid_password")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Invalid credentials",
			Code:  "INVALID_CREDENTIALS",
		})
		return
	}

	h.db.User().UpdateLastLogin(user.ID, c.ClientIP())

	accessToken, expiresAt, err := middleware.GenerateAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to generate token",
			Code:  "TOKEN_GENERATION_FAILED",
		})
		return
	}

	refreshToken, err := middleware.GenerateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to generate token",
			Code:  "TOKEN_GENERATION_FAILED",
		})
		return
	}

	h.logAudit(c, user.ID, user.Email, "login", "user", user.ID, "success", c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         *user,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims != nil {
		h.logAudit(c, claims.UserID, claims.Email, "logout", "user", claims.UserID, "success", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	user, err := h.db.User().GetByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "User not found",
			Code:  "USER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) RefreshTokenGET(c *gin.Context) {
	refreshToken := c.GetHeader("X-Refresh-Token")
	if refreshToken == "" {
		cookie, err := c.Cookie("refresh_token")
		if err == nil {
			refreshToken = cookie
		}
	}

	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Refresh token required",
			Code:  "REFRESH_TOKEN_REQUIRED",
		})
		return
	}

	userID, err := middleware.ValidateRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Invalid refresh token",
			Code:  "INVALID_REFRESH_TOKEN",
		})
		return
	}

	user, err := h.db.User().GetByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found",
			Code:  "USER_NOT_FOUND",
		})
		return
	}

	accessToken, expiresAt, err := middleware.GenerateAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to generate token",
			Code:  "TOKEN_GENERATION_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:     accessToken,
		ExpiresAt: expiresAt,
		User:      *user,
	})
}

func (h *AuthHandler) GetCSRFToken(c *gin.Context) {
	cfg := config.Get()
	token, err := cfg.Security.GenerateCSRFToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to generate CSRF token",
			Code:  "CSRF_GENERATION_FAILED",
		})
		return
	}

	encoded := base64.URLEncoding.EncodeToString([]byte(token))

	c.SetCookie("csrf_token", encoded, 3600*24, "/", "", false, false)

	c.JSON(http.StatusOK, gin.H{"csrf_token": token})
}

func (h *AuthHandler) logFailedLogin(c *gin.Context, username, reason string) {
	h.logAudit(c, "", username, "login_failed", "user", "", reason, c.ClientIP(), c.GetHeader("User-Agent"))
}

func (h *AuthHandler) logAudit(c *gin.Context, userID, userEmail, action, resource, resourceID, details, ip, userAgent string) {
	log := &models.AuditLog{
		UserID:     userID,
		UserEmail:  userEmail,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IP:         ip,
		UserAgent:  userAgent,
	}
	h.db.AuditLog().Create(log)
}
