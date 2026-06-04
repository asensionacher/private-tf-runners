package handlers

import (
	"database/sql"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/private-tf-runners/server/internal/database"
	"github.com/private-tf-runners/server/internal/middleware"
	"github.com/private-tf-runners/server/internal/models"
	"github.com/private-tf-runners/server/internal/services"
)

type StackHandler struct {
	db         *database.DB
	gitService *services.GitService
}

func NewStackHandler(db *database.DB) *StackHandler {
	return &StackHandler{db: db, gitService: services.NewGitService()}
}

func (h *StackHandler) List(c *gin.Context) {
	stacks, err := h.db.Stack().GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve stacks",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, stacks)
}

func (h *StackHandler) Get(c *gin.Context) {
	id := c.Param("id")

	stack, err := h.db.Stack().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Stack not found",
				Code:  "STACK_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve stack",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, stack)
}

func (h *StackHandler) GetWithRefs(c *gin.Context) {
	id := c.Param("id")

	stack, err := h.db.Stack().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Stack not found",
				Code:  "STACK_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve stack",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, stack)
}

func (h *StackHandler) Create(c *gin.Context) {
	var req models.CreateStackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if !req.Provider.IsValid() {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid provider. Must be 'opentofu' or 'terraform'",
			Code:  "INVALID_PROVIDER",
		})
		return
	}

	normalizedURL := normalizeGitURL(req.GitURL)
	if normalizedURL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid git URL",
			Code:  "INVALID_GIT_URL",
		})
		return
	}

	if strings.ContainsAny(req.Name, "/\\") {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Stack name cannot contain path separators",
			Code:  "INVALID_NAME",
		})
		return
	}

	stack := &models.Stack{
		ID:                 uuid.New().String(),
		Name:               strings.TrimSpace(req.Name),
		Description:        strings.TrimSpace(req.Description),
		GitURL:             normalizedURL,
		GitFolder:          strings.Trim(req.GitFolder, "/"),
		Provider:           req.Provider,
		PublishedBranches:  parseCommaSeparated(req.Branch),
		PublishedTags:      parseCommaSeparated(req.Tags),
	}

	if err := h.db.Stack().Create(stack); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: "Stack with this name already exists",
				Code:  "STACK_EXISTS",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create stack",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "create", "stack", stack.ID, "Stack created: "+stack.Name)

	c.JSON(http.StatusCreated, stack)
}

func (h *StackHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateStackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	stack, err := h.db.Stack().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Stack not found",
				Code:  "STACK_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve stack",
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
				Error: "Stack name cannot contain path separators",
				Code:  "INVALID_NAME",
			})
			return
		}
		stack.Name = name
	}

	if req.Description != nil {
		stack.Description = strings.TrimSpace(*req.Description)
	}

	if req.GitURL != nil {
		normalizedURL := normalizeGitURL(*req.GitURL)
		if normalizedURL == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid git URL",
				Code:  "INVALID_GIT_URL",
			})
			return
		}
		stack.GitURL = normalizedURL
	}

	if req.GitFolder != nil {
		stack.GitFolder = strings.Trim(*req.GitFolder, "/")
	}

	if req.Provider != nil {
		if !req.Provider.IsValid() {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid provider",
				Code:  "INVALID_PROVIDER",
			})
			return
		}
		stack.Provider = *req.Provider
	}

	if req.PublishedBranches != nil {
		stack.PublishedBranches = *req.PublishedBranches
	}

	if req.PublishedTags != nil {
		stack.PublishedTags = *req.PublishedTags
	}

	if err := h.db.Stack().Update(stack); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update stack",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "update", "stack", stack.ID, "Stack updated: "+stack.Name)

	c.JSON(http.StatusOK, stack)
}

func (h *StackHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.db.Stack().Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Stack not found",
				Code:  "STACK_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete stack",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "delete", "stack", id, "Stack deleted")

	c.JSON(http.StatusOK, gin.H{"message": "Stack deleted successfully"})
}

func (h *StackHandler) ValidateRepo(c *gin.Context) {
	gitURL := c.Query("url")
	if gitURL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Git URL is required",
			Code:  "GIT_URL_REQUIRED",
		})
		return
	}

	normalizedURL := normalizeGitURL(gitURL)
	if normalizedURL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid git URL",
			Code:  "INVALID_GIT_URL",
		})
		return
	}

	info, err := h.gitService.FetchRepoInfo(normalizedURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Failed to fetch repository: " + err.Error(),
			Code:  "FETCH_REPO_ERROR",
		})
		return
	}

	if len(info.Branches) == 0 && len(info.Tags) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "No branches or tags found in repository",
			Code:  "NO_REFS_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

func (h *StackHandler) RefetchRepo(c *gin.Context) {
	id := c.Param("id")

	stack, err := h.db.Stack().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Stack not found",
				Code:  "STACK_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve stack",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	info, err := h.gitService.FetchRepoInfo(stack.GitURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Failed to fetch repository: " + err.Error(),
			Code:  "FETCH_REPO_ERROR",
		})
		return
	}

	if len(info.Branches) == 0 && len(info.Tags) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "No branches or tags found in repository",
			Code:  "NO_REFS_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

func (h *StackHandler) SyncRefs(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Branches []string `json:"branches"`
		Tags     []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	stack, err := h.db.Stack().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Stack not found",
				Code:  "STACK_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve stack",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	if len(req.Branches) == 0 && len(req.Tags) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "At least one branch or tag is required",
			Code:  "REFS_REQUIRED",
		})
		return
	}

	stack.PublishedBranches = req.Branches
	stack.PublishedTags = req.Tags

	if err := h.db.Stack().Update(stack); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update stack",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "sync_refs", "stack", stack.ID, "Refs synced: "+stack.Name)

	c.JSON(http.StatusOK, stack)
}

func (h *StackHandler) logAudit(c *gin.Context, claims *models.TokenClaims, action, resource, resourceID, details string) {
	userID := ""
	userEmail := ""
	if claims != nil {
		userID = claims.UserID
		userEmail = claims.Email
	}

	log := &models.AuditLog{
		UserID:     userID,
		UserEmail:  userEmail,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IP:         c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}
	h.db.AuditLog().Create(log)
}

func normalizeGitURL(gitURL string) string {
	gitURL = strings.TrimSpace(gitURL)
	if gitURL == "" {
		return ""
	}

	parsed, err := url.Parse(gitURL)
	if err != nil {
		return ""
	}

	if parsed.Scheme == "git" {
		parsed.Scheme = "https"
		if !strings.HasSuffix(parsed.Path, ".git") {
			parsed.Path = parsed.Path + ".git"
		}
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		if strings.Contains(gitURL, "@") {
			return ""
		}
		parsed, err = url.Parse("https://" + gitURL)
		if err != nil {
			return ""
		}
	}

	switch strings.ToLower(parsed.Host) {
	case "github.com":
		parsed.Scheme = "https"
	case "gitlab.com":
		parsed.Scheme = "https"
	case "bitbucket.org":
		parsed.Scheme = "https"
	}

	return parsed.String()
}

func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
