package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/private-tf-runners/server/internal/backends"
	"github.com/private-tf-runners/server/internal/database"
	"github.com/private-tf-runners/server/internal/middleware"
	"github.com/private-tf-runners/server/internal/models"
)

type BackendHandler struct {
	db *database.DB
}

func NewBackendHandler(db *database.DB) *BackendHandler {
	return &BackendHandler{db: db}
}

func (h *BackendHandler) List(c *gin.Context) {
	backends, err := h.db.Backend().GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve backends",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	if backends == nil {
		backends = []models.Backend{}
	}

	c.JSON(http.StatusOK, backends)
}

func (h *BackendHandler) Get(c *gin.Context) {
	id := c.Param("id")

	backend, err := h.db.Backend().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Backend not found",
				Code:  "BACKEND_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve backend",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, backend)
}

func (h *BackendHandler) Create(c *gin.Context) {
	var req models.CreateBackendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if !req.Type.IsValid() {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid backend type. Must be 'local', 's3', 'azurerm', 'gcs', 'kubernetes', 'http', 'consul', or 'pg'",
			Code:  "INVALID_BACKEND_TYPE",
		})
		return
	}

	if strings.ContainsAny(req.Name, "/\\") {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Backend name cannot contain path separators",
			Code:  "INVALID_NAME",
		})
		return
	}

	if err := backends.ValidateConfig(string(req.Type), req.Config); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_BACKEND_CONFIG",
		})
		return
	}

	backend := &models.Backend{
		ID:     uuid.New().String(),
		Name:   strings.TrimSpace(req.Name),
		Type:   string(req.Type),
		Config: req.Config,
	}

	if err := h.db.Backend().Create(backend); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: "Backend with this name already exists",
				Code:  "BACKEND_EXISTS",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create backend",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "create", "backend", backend.ID, "Backend created: "+backend.Name)

	c.JSON(http.StatusCreated, backend)
}

func (h *BackendHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateBackendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	backend, err := h.db.Backend().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Backend not found",
				Code:  "BACKEND_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve backend",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Name cannot be empty",
				Code:  "INVALID_NAME",
			})
			return
		}
		if strings.ContainsAny(name, "/\\") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Backend name cannot contain path separators",
				Code:  "INVALID_NAME",
			})
			return
		}
		backend.Name = name
	}

	if req.Type != nil {
		if !req.Type.IsValid() {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid backend type",
				Code:  "INVALID_BACKEND_TYPE",
			})
			return
		}
		backend.Type = string(*req.Type)
	}

	if req.Config != nil {
		backend.Config = *req.Config
	}

	backendType := backend.Type
	if req.Type != nil {
		backendType = string(*req.Type)
	}
	if err := backends.ValidateConfig(backendType, backend.Config); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_BACKEND_CONFIG",
		})
		return
	}

	if err := h.db.Backend().Update(backend); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update backend",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "update", "backend", backend.ID, "Backend updated: "+backend.Name)

	c.JSON(http.StatusOK, backend)
}

func (h *BackendHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.db.Backend().Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Backend not found",
				Code:  "BACKEND_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete backend",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "delete", "backend", id, "Backend deleted")

	c.JSON(http.StatusOK, gin.H{"message": "Backend deleted successfully"})
}

func (h *BackendHandler) Schema(c *gin.Context) {
	backendType := c.Param("type")
	schema := backends.GetBackendSchema(backendType)
	if schema == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Unknown backend type",
			Code:  "BACKEND_NOT_FOUND",
		})
		return
	}
	c.JSON(http.StatusOK, schema)
}

func (h *BackendHandler) Schemas(c *gin.Context) {
	schemas := backends.GetAllBackendSchemas()
	c.JSON(http.StatusOK, schemas)
}

func (h *BackendHandler) logAudit(c *gin.Context, claims *models.TokenClaims, action, resource, resourceID, details string) {
	userID := ""
	userEmail := ""
	if claims != nil {
		userID = claims.UserID
		userEmail = claims.Email
	}

	log := &models.AuditLog{
		UserID:    userID,
		UserEmail: userEmail,
		Action:    action,
		Resource:  resource,
		ResourceID: resourceID,
		Details:   details,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}
	h.db.AuditLog().Create(log)
}