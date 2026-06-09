package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/private-tf-runners/server/internal/database"
	"github.com/private-tf-runners/server/internal/middleware"
	"github.com/private-tf-runners/server/internal/models"
)

type RunnerHandler struct {
	db *database.DB
}

func NewRunnerHandler(db *database.DB) *RunnerHandler {
	return &RunnerHandler{db: db}
}

func (h *RunnerHandler) List(c *gin.Context) {
	runners, err := h.db.Runner().GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve runners",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	if runners == nil {
		runners = []models.Runner{}
	}

	c.JSON(http.StatusOK, runners)
}

func (h *RunnerHandler) Get(c *gin.Context) {
	id := c.Param("id")

	runner, err := h.db.Runner().GetByID(id)
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

	c.JSON(http.StatusOK, runner)
}

func (h *RunnerHandler) Create(c *gin.Context) {
	var req models.CreateRunnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	token, err := generateToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create runner",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	runner := &models.Runner{
		Name:   req.Name,
		Token:  token,
		Status: models.RunnerStatusOffline,
	}

	if err := h.db.Runner().Create(runner); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create runner: " + err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "create", "runner", runner.ID, "Runner created: "+runner.Name)

	c.JSON(http.StatusCreated, models.RunnerCreatedResponse{
		ID:    runner.ID,
		Name:  runner.Name,
		Token: runner.Token,
	})
}

func (h *RunnerHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateRunnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	runner, err := h.db.Runner().GetByID(id)
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

	if req.Name != nil {
		runner.Name = *req.Name
	}

	if err := h.db.Runner().Update(runner); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update runner",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "update", "runner", runner.ID, "Runner updated: "+runner.Name)

	c.JSON(http.StatusOK, runner)
}

func (h *RunnerHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	runner, err := h.db.Runner().GetByID(id)
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

	if err := h.db.Runner().Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete runner",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "delete", "runner", id, "Runner deleted: "+runner.Name)

	c.JSON(http.StatusOK, gin.H{"message": "Runner deleted successfully"})
}

func (h *RunnerHandler) ResetToken(c *gin.Context) {
	id := c.Param("id")

	runner, err := h.db.Runner().GetByID(id)
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

	newToken, err := generateToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to reset token",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	runner.Token = newToken
	if err := h.db.Runner().Update(runner); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to reset token",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "reset-token", "runner", id, "Runner token reset: "+runner.Name)

	c.JSON(http.StatusOK, models.RunnerCreatedResponse{
		ID:    runner.ID,
		Name:  runner.Name,
		Token: runner.Token,
	})
}

func (h *RunnerHandler) GetRunnerRuns(c *gin.Context) {
	id := c.Param("id")

	runnerFromToken, exists := c.Get("runner")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
			Code:  "UNAUTHORIZED",
		})
		return
	}
	runner := runnerFromToken.(*models.Runner)

	if runner.ID != id {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Cannot access runs for another runner",
			Code:  "FORBIDDEN",
		})
		return
	}

	_, err := h.db.Runner().GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Runner not found",
				Code:  "RUNNER_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve runner",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	runs, err := h.db.Run().GetByRunnerID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve runs: " + err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	if runs == nil {
		runs = []*models.Run{}
	}

	c.JSON(http.StatusOK, runs)
}

func (h *RunnerHandler) Heartbeat(c *gin.Context) {
	id := c.Param("id")

	var req models.RunnerHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	runnerFromToken, exists := c.Get("runner")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	runner := runnerFromToken.(*models.Runner)

	if runner.ID != id {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Runner ID mismatch",
			Code:  "FORBIDDEN",
		})
		return
	}

	if err := h.db.Runner().Heartbeat(id, req.Status, req.CurrentRunID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update heartbeat",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      runner.ID,
		"name":    runner.Name,
		"status":  req.Status,
		"message": "Heartbeat received",
	})
}

func (h *RunnerHandler) AssignRun(c *gin.Context) {
	runID := c.Param("id")

	var req struct {
		RunnerID string `json:"runner_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	run, err := h.db.Run().GetByID(runID)
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

	if run.Status != models.StatusPending {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Run is not in pending status",
			Code:  "INVALID_STATUS",
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

	if err := h.db.Run().AssignRunner(runID, req.RunnerID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to assign runner",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "assign", "run", runID, "Run assigned to runner: "+runner.Name)

	c.JSON(http.StatusOK, gin.H{"message": "Run assigned successfully"})
}

func (h *RunnerHandler) ApproveRun(c *gin.Context) {
	runID := c.Param("id")

	run, err := h.db.Run().GetByID(runID)
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

	if (run.Status != models.StatusPending && run.Status != models.StatusRunning && run.Status != models.StatusPlanned) || run.Phase != models.PhasePlan {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Run is not awaiting plan approval",
			Code:  "INVALID_STATUS",
		})
		return
	}

	if err := h.db.Run().UpdateStatus(runID, models.StatusApproved); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to approve run",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	if err := h.db.Run().UpdatePhase(runID, models.PhaseApply); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update run phase",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "approve", "run", runID, "Run approved for apply")

	c.JSON(http.StatusOK, gin.H{"message": "Run approved successfully"})
}

func (h *RunnerHandler) RejectRun(c *gin.Context) {
	runID := c.Param("id")

	run, err := h.db.Run().GetByID(runID)
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

	if (run.Status != models.StatusPending && run.Status != models.StatusRunning && run.Status != models.StatusPlanned) || run.Phase != models.PhasePlan {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Run is not awaiting plan approval",
			Code:  "INVALID_STATUS",
		})
		return
	}

	if err := h.db.Runner().UpdateStatus(run.RunnerID, models.RunnerStatusOnline, ""); err != nil {
	}

	if err := h.db.Run().Finish(runID, models.StatusFailed, "Rejected by user"); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to reject run",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "reject", "run", runID, "Run rejected")

	c.JSON(http.StatusOK, gin.H{"message": "Run rejected successfully"})
}

func (h *RunnerHandler) CancelRun(c *gin.Context) {
	runID := c.Param("id")

	run, err := h.db.Run().GetByID(runID)
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

	if run.Status != models.StatusPending && run.Status != models.StatusRunning && run.Status != models.StatusApproved && run.Status != models.StatusPlanned {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Run cannot be canceled",
			Code:  "INVALID_STATUS",
		})
		return
	}

	if run.RunnerID != "" {
		h.db.Runner().UpdateStatus(run.RunnerID, models.RunnerStatusOnline, "")
	}

	if err := h.db.Run().Finish(runID, models.StatusCanceled, "Canceled by user"); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to cancel run",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	claims := middleware.GetClaims(c)
	h.logAudit(c, claims, "cancel", "run", runID, "Run canceled")

	c.JSON(http.StatusOK, gin.H{"message": "Run canceled successfully"})
}

func (h *RunnerHandler) UpdatePlanOutput(c *gin.Context) {
	runID := c.Param("id")

	runnerFromToken, exists := c.Get("runner")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
			Code:  "UNAUTHORIZED",
		})
		return
	}
	runner := runnerFromToken.(*models.Runner)

	var req struct {
		PlanOutput string `json:"plan_output"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	run, err := h.db.Run().GetByID(runID)
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

	if run.RunnerID != runner.ID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Run is not assigned to this runner",
			Code:  "FORBIDDEN",
		})
		return
	}

	if run.Status != models.StatusRunning && run.Status != models.StatusPending {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Run is not in running or pending status",
			Code:  "INVALID_STATUS",
		})
		return
	}

	if err := h.db.Run().SetPlanOutput(runID, req.PlanOutput); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update plan output",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plan output updated"})
}

func (h *RunnerHandler) UpdateApplyOutput(c *gin.Context) {
	runID := c.Param("id")

	runnerFromToken, exists := c.Get("runner")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
			Code:  "UNAUTHORIZED",
		})
		return
	}
	runner := runnerFromToken.(*models.Runner)

	var req struct {
		ApplyOutput string `json:"apply_output"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	run, err := h.db.Run().GetByID(runID)
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

	if run.RunnerID != runner.ID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Run is not assigned to this runner",
			Code:  "FORBIDDEN",
		})
		return
	}

	if run.Status != models.StatusRunning && run.Status != models.StatusApproved && run.Status != models.StatusPlanned {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Run is not in running, approved, or planned status",
			Code:  "INVALID_STATUS",
		})
		return
	}

	if err := h.db.Run().SetApplyOutput(runID, req.ApplyOutput); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update apply output",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Apply output updated"})
}

func (h *RunnerHandler) UpdateRunStatus(c *gin.Context) {
	runID := c.Param("id")

	runnerFromToken, exists := c.Get("runner")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
			Code:  "UNAUTHORIZED",
		})
		return
	}
	runner := runnerFromToken.(*models.Runner)

	var req struct {
		Status string `json:"status"`
		Logs   string `json:"logs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	run, err := h.db.Run().GetByID(runID)
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

	if run.RunnerID != runner.ID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Run is not assigned to this runner",
			Code:  "FORBIDDEN",
		})
		return
	}

	if err := h.db.Run().UpdateStatus(runID, models.RunStatus(req.Status)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update run status",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Run status updated"})
}

func (h *RunnerHandler) UpdateWorkDir(c *gin.Context) {
	runID := c.Param("id")

	runnerFromToken, exists := c.Get("runner")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
			Code:  "UNAUTHORIZED",
		})
		return
	}
	runner := runnerFromToken.(*models.Runner)

	var req struct {
		WorkDir string `json:"work_dir"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	run, err := h.db.Run().GetByID(runID)
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

	if run.RunnerID != runner.ID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Run is not assigned to this runner",
			Code:  "FORBIDDEN",
		})
		return
	}

	if err := h.db.Run().SetWorkDir(runID, req.WorkDir); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update work dir",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Run work dir updated"})
}

func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (h *RunnerHandler) WaitForRun(c *gin.Context) {
	runID := c.Param("id")

	runnerFromToken, isRunner := c.Get("runner")
	claims := middleware.GetClaims(c)

	if !isRunner && claims == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	timeoutStr := c.DefaultQuery("timeout", "30")
	timeout := 30
	if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 && t <= 60 {
		timeout = t
	}

	run, err := h.db.Run().GetByID(runID)
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

	if isRunner {
		runner := runnerFromToken.(*models.Runner)
		if run.RunnerID != runner.ID {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Run is not assigned to this runner",
				Code:  "FORBIDDEN",
			})
			return
		}
	}

	if isTerminalStatus(run.Status) {
		c.JSON(http.StatusOK, run)
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	
	for {
		select {
		case <-c.Request.Context().Done():
			c.JSON(http.StatusOK, run)
			return
		case <-ticker.C:
			updatedRun, err := h.db.Run().GetByID(runID)
			if err != nil {
				c.JSON(http.StatusOK, run)
				return
			}
			if updatedRun.Status != run.Status || isTerminalStatus(updatedRun.Status) {
				c.JSON(http.StatusOK, updatedRun)
				return
			}
			if time.Now().After(deadline) {
				c.JSON(http.StatusOK, updatedRun)
				return
			}
		}
	}
}

func isTerminalStatus(status models.RunStatus) bool {
	switch status {
	case models.StatusApplied, models.StatusFailed, models.StatusRejected, models.StatusCanceled:
		return true
	default:
		return false
	}
}

func (h *RunnerHandler) logAudit(c *gin.Context, claims *models.TokenClaims, action, resource, resourceID, details string) {
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