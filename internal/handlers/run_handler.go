package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/private-tf-runners/server/internal/backends"
	"github.com/private-tf-runners/server/internal/database"
	"github.com/private-tf-runners/server/internal/middleware"
	"github.com/private-tf-runners/server/internal/models"
)

type RunHandler struct {
	db *database.DB
}

func NewRunHandler(db *database.DB) *RunHandler {
	return &RunHandler{db: db}
}

func (h *RunHandler) List(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	runs, total, err := h.db.Run().GetAll(pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve runs",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       runs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func (h *RunHandler) Get(c *gin.Context) {
	id := c.Param("id")

	run, err := h.db.Run().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Run not found",
				Code:  "RUN_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve run",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, run)
}

func (h *RunHandler) Create(c *gin.Context) {
	var req models.CreateRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	log.Printf("[DEBUG] CreateRun request received: %+v", req)
	stack, err := h.db.Stack().GetByID(req.StackID)
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

	if !h.isBranchOrTagPublished(stack, req.Branch) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Branch or tag not published for this stack",
			Code:  "BRANCH_NOT_PUBLISHED",
		})
		return
	}

	runner, err := h.db.Runner().GetByID(req.RunnerID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Runner not found",
				Code:  "RUNNER_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve runner",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	if runner.Status == models.RunnerStatusBusy {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Runner is busy with another run",
			Code:  "RUNNER_BUSY",
		})
		return
	}

	tfvarsValuesJSON, _ := json.Marshal(req.TfvarsValues)
	backendConfigEnvJSON, _ := json.Marshal(req.BackendConfigEnv)

	backendType := models.TerraformBackendType(req.BackendType)
	commands := h.generateCommands(stack, backendType, req.Branch, req.BackendConfig, req.BackendConfigEnv, req.TfvarsFiles, req.TfvarsValues)

	backendConfigJSON, _ := json.Marshal(req.BackendConfig)

	var backendSchema *backends.BackendSchema
	var backendConfigFields []backends.Field
	if req.BackendType != "" {
		backendSchema = backends.GetBackendSchema(req.BackendType)
		if backendSchema != nil {
			for _, method := range backendSchema.AuthMethods {
				backendConfigFields = append(backendConfigFields, method.Fields...)
			}
			if len(backendSchema.RequiredFields) > 0 {
				backendConfigFields = append(backendConfigFields, backendSchema.RequiredFields...)
			}
			if len(backendSchema.OptionalFields) > 0 {
				backendConfigFields = append(backendConfigFields, backendSchema.OptionalFields...)
			}
		}
	}

	run := &models.Run{
		ID:           uuid.New().String(),
		StackID:      req.StackID,
		RunnerID:     req.RunnerID,
		TriggerType:  models.TriggerManual,
		Branch:       req.Branch,
		CommitSHA:    req.CommitSHA,
		Status:       models.StatusPending,
		Phase:        models.PhasePlan,
		BackendType:  req.BackendType,
		BackendKey:   string(backendConfigJSON),
		BackendSchema: backendSchema,
		BackendConfigFields: backendConfigFields,
		TfvarsFiles:  req.TfvarsFiles,
		TfvarsValues: string(tfvarsValuesJSON),
		EnvVars:      string(backendConfigEnvJSON),
		Commands:     commands,
	}

	if err := h.db.Run().Create(run); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create run",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "create", "run", run.ID, "Run created for stack: "+stack.Name)

	c.JSON(http.StatusCreated, run)
}

func (h *RunHandler) GetByStackID(c *gin.Context) {
	stackID := c.Param("id")

	_, err := h.db.Stack().GetByID(stackID)
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

	runs, err := h.db.Run().GetByStackID(stackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve runs",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, runs)
}

func (h *RunHandler) isBranchOrTagPublished(stack *models.Stack, value string) bool {
	for _, branch := range stack.PublishedBranches {
		if branch == value {
			return true
		}
	}
	for _, tag := range stack.PublishedTags {
		if tag == value {
			return true
		}
	}
	return false
}

func (h *RunHandler) generateCommands(stack *models.Stack, backendType models.TerraformBackendType, branch string, backendConfig map[string]string, backendConfigEnv map[string]string, tfvarsFiles []string, tfvarsValues map[string]string) string {
	var commands []string
	provider := "tofu"
	if stack.Provider == models.ProviderTerraform {
		provider = "terraform"
	}

	gitFolder := ""
	if stack.GitFolder != "" {
		gitFolder = stack.GitFolder + "/"
	}

	commands = append(commands, fmt.Sprintf("git clone %s /tmp/%s-%s", stack.GitURL, stack.ID, branch))
	commands = append(commands, fmt.Sprintf("cd /tmp/%s-%s && git checkout %s", stack.ID, branch, branch))

	for k, v := range backendConfigEnv {
		if v != "" {
			commands = append(commands, fmt.Sprintf("export %s='%s'", k, v))
		}
	}

	var initParts []string
	initParts = append(initParts, provider, "init")

	if backendType != "" && backendType != models.TerraformBackendLocal {
		initParts = append(initParts, fmt.Sprintf("-backend-config='type=%s'", backendType))

		for k, v := range backendConfig {
			if v != "" && v != "{}" {
				initParts = append(initParts, fmt.Sprintf("-backend-config='%s=%s'", k, v))
			}
		}
	}

	var planCmd, applyCmd string
	planCmd = provider + " plan"
	applyCmd = provider + " apply"

	for _, tfvarsFile := range tfvarsFiles {
		if tfvarsFile != "" {
			planCmd = fmt.Sprintf("%s -var-file='%s'", planCmd, tfvarsFile)
			applyCmd = fmt.Sprintf("%s -var-file='%s'", applyCmd, tfvarsFile)
		}
	}

	if len(tfvarsValues) > 0 {
		var tfvarsArgs []string
		for k, v := range tfvarsValues {
			tfvarsArgs = append(tfvarsArgs, fmt.Sprintf("-var='%s=%s'", k, v))
		}
		tfvarsStr := strings.Join(tfvarsArgs, " ")
		planCmd = fmt.Sprintf("%s %s", planCmd, tfvarsStr)
		applyCmd = fmt.Sprintf("%s %s", applyCmd, tfvarsStr)
	}

	commands = append(commands, fmt.Sprintf("cd /tmp/%s-%s/%s", stack.ID, branch, gitFolder))
	commands = append(commands, strings.Join(initParts, " "))
	commands = append(commands, planCmd)
	commands = append(commands, applyCmd)

	return strings.Join(commands, "\n")
}

func (h *RunHandler) logAudit(c *gin.Context, claims *models.TokenClaims, action, resource, resourceID, details string) {
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