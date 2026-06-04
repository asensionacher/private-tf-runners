package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/private-tf-runners/server/internal/database"
	"github.com/private-tf-runners/server/internal/middleware"
	"github.com/private-tf-runners/server/internal/models"
)

type UserHandler struct {
	db *database.DB
}

func NewUserHandler(db *database.DB) *UserHandler {
	return &UserHandler{db: db}
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func (h *UserHandler) logAudit(c *gin.Context, claims *models.TokenClaims, action, resource, resourceID, details string) {
	log := &models.AuditLog{
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
	}
	if claims != nil {
		log.UserID = claims.UserID
		log.UserEmail = claims.Email
	}
	log.IP = c.ClientIP()
	log.UserAgent = c.GetHeader("User-Agent")
	h.db.AuditLog().Create(log)
}

func (h *UserHandler) List(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page := 1
	if p, err := parseInt(pageStr); err == nil && p > 0 {
		page = p
	}
	pageSize := 20
	if ps, err := parseInt(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	users, total, err := h.db.User().GetAll(pageSize, (page-1)*pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve users",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        users,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}

func (h *UserHandler) Get(c *gin.Context) {
	id := c.Param("id")

	user, err := h.db.User().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "User not found",
				Code:  "USER_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve user",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Username string      `json:"username" binding:"required"`
		Email    string      `json:"email" binding:"required,email"`
		Password string      `json:"password" binding:"required,min=8"`
		Role     models.Role `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if !req.Role.IsValid() {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid role",
			Code:  "INVALID_ROLE",
		})
		return
	}

	existingUser, _ := h.db.User().GetByUsername(req.Username)
	if existingUser != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: "Username already exists",
			Code:  "USERNAME_EXISTS",
		})
		return
	}

	existingEmail, _ := h.db.User().GetByEmail(req.Email)
	if existingEmail != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: "Email already exists",
			Code:  "EMAIL_EXISTS",
		})
		return
	}

	user := &models.User{
		ID:       uuid.New().String(),
		Username: req.Username,
		Email:    req.Email,
		Role:     string(req.Role),
	}

	if err := h.db.User().Create(user, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create user",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "create", "user", user.ID, "User created: "+user.Username)

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	user, err := h.db.User().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "User not found",
				Code:  "USER_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve user",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	var req struct {
		Username *string      `json:"username"`
		Email    *string      `json:"email"`
		Role     *models.Role `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if req.Username != nil {
		existingUser, _ := h.db.User().GetByUsername(*req.Username)
		if existingUser != nil && existingUser.ID != id {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: "Username already exists",
				Code:  "USERNAME_EXISTS",
			})
			return
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		existingEmail, _ := h.db.User().GetByEmail(*req.Email)
		if existingEmail != nil && existingEmail.ID != id {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: "Email already exists",
				Code:  "EMAIL_EXISTS",
			})
			return
		}
		user.Email = *req.Email
	}

	if req.Role != nil {
		if !req.Role.IsValid() {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid role",
				Code:  "INVALID_ROLE",
			})
			return
		}
		user.Role = string(*req.Role)
	}

	if err := h.db.User().Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update user",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "update", "user", user.ID, "User updated: "+user.Username)

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	user, err := h.db.User().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "User not found",
				Code:  "USER_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve user",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	if err := h.db.User().Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete user",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "delete", "user", id, "User deleted: "+user.Username)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	user, err := h.db.User().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "User not found",
				Code:  "USER_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve user",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	if err := h.db.User().UpdatePassword(id, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to reset password",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "reset-password", "user", id, "Password reset for user: "+user.Username)

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}