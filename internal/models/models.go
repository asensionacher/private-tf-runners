package models

import (
	"time"

	"github.com/private-tf-runners/server/internal/backends"
)

type Provider string

const (
	ProviderOpenTofu  Provider = "opentofu"
	ProviderTerraform Provider = "terraform"
)

func (p Provider) IsValid() bool {
	return p == ProviderOpenTofu || p == ProviderTerraform
}

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleOperator  Role = "operator"
	RoleViewer    Role = "viewer"
)

func (r Role) IsValid() bool {
	return r == RoleAdmin || r == RoleOperator || r == RoleViewer
}

type Permission string

const (
	PermissionStackRead    Permission = "stack:read"
	PermissionStackCreate  Permission = "stack:create"
	PermissionStackDelete  Permission = "stack:delete"
	PermissionRunRead      Permission = "run:read"
	PermissionRunCreate    Permission = "run:create"
	PermissionRunDelete    Permission = "run:delete"
	PermissionRunnerRead   Permission = "runner:read"
	PermissionRunnerCreate Permission = "runner:create"
	PermissionRunnerDelete Permission = "runner:delete"
	PermissionRunnerToken  Permission = "runner:token"
	PermissionUserCreate   Permission = "user:create"
	PermissionUserAdmin    Permission = "user:admin"
)

var RolePermissions = map[Role][]Permission{
	RoleViewer: {
		PermissionStackRead,
		PermissionRunRead,
		PermissionRunnerRead,
	},
	RoleOperator: {
		PermissionStackRead,
		PermissionStackCreate,
		PermissionStackDelete,
		PermissionRunRead,
		PermissionRunCreate,
		PermissionRunDelete,
		PermissionRunnerRead,
		PermissionRunnerCreate,
		PermissionRunnerDelete,
		PermissionRunnerToken,
	},
	RoleAdmin: {
		PermissionStackRead,
		PermissionStackCreate,
		PermissionStackDelete,
		PermissionRunRead,
		PermissionRunCreate,
		PermissionRunDelete,
		PermissionRunnerRead,
		PermissionRunnerCreate,
		PermissionRunnerDelete,
		PermissionRunnerToken,
		PermissionUserCreate,
		PermissionUserAdmin,
	},
}

func (r Role) HasPermission(p Permission) bool {
	perms, ok := RolePermissions[r]
	if !ok {
		return false
	}
	for _, perm := range perms {
		if perm == p {
			return true
		}
	}
	return false
}

type TriggerType string

const (
	TriggerManual   TriggerType = "manual"
	TriggerPush     TriggerType = "push"
	TriggerSchedule TriggerType = "schedule"
)

type RunStatus string

const (
	StatusPending   RunStatus = "pending"
	StatusRunning   RunStatus = "running"
	StatusSuccess   RunStatus = "success"
	StatusFailed    RunStatus = "failed"
	StatusApproved  RunStatus = "approved"
	StatusRejected  RunStatus = "rejected"
	StatusPlanned   RunStatus = "planned"
	StatusApplied   RunStatus = "applied"
	StatusCanceled  RunStatus = "canceled"
)

type RunPhase string

const (
	PhasePlan   RunPhase = "plan"
	PhaseApply  RunPhase = "apply"
	PhaseFinish RunPhase = "finish"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

type Backend struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Config    string    `json:"config"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID               string     `json:"id"`
	Username         string     `json:"username"`
	Email            string     `json:"email"`
	PasswordHash     string     `json:"-"`
	Role             string     `json:"role"`
	Provider         Provider   `json:"provider,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LoginAttempts    int        `json:"-"`
	LockedUntil      *time.Time `json:"-"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP      string     `json:"last_login_ip,omitempty"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
	TwoFactorSecret  string     `json:"-"`
}

type AuditLog struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	UserEmail   string    `json:"user_email"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource"`
	ResourceID  string    `json:"resource_id,omitempty"`
	Details     string    `json:"details,omitempty"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	CreatedAt   time.Time `json:"created_at"`
}

type Stack struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	GitURL             string    `json:"git_url"`
	GitFolder          string    `json:"git_folder"`
	Provider           Provider  `json:"provider"`
	PublishedBranches  []string  `json:"published_branches"`
	PublishedTags      []string  `json:"published_tags"`
	TfvarsFiles       []string  `json:"tfvars_files"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Run struct {
	ID              string       `json:"id"`
	StackID         string       `json:"stack_id"`
	StackName       string       `json:"stack_name,omitempty"`
	TriggerType     TriggerType  `json:"trigger_type"`
	Branch          string       `json:"branch"`
	CommitSHA       string       `json:"commit_sha"`
	Status          RunStatus    `json:"status"`
	Phase           RunPhase     `json:"phase,omitempty"`
	Logs            string       `json:"logs,omitempty"`
	PlanOutput      string       `json:"plan_output,omitempty"`
	ApplyOutput     string       `json:"apply_output,omitempty"`
	BackendType     string                  `json:"backend_type,omitempty"`
	BackendKey      string                 `json:"backend_key,omitempty"`
	BackendSchema   *backends.BackendSchema `json:"backend_schema,omitempty"`
	BackendConfigFields []backends.Field    `json:"backend_config_fields,omitempty"`
	TfvarsFiles     []string     `json:"tfvars_files,omitempty"`
	TfvarsValues    string       `json:"tfvars_values,omitempty"`
	EnvVars         string       `json:"env_vars,omitempty"`
	Commands        string       `json:"commands,omitempty"`
	RunnerID        string       `json:"runner_id,omitempty"`
	RunnerName      string       `json:"runner_name,omitempty"`
	WorkDir         string       `json:"work_dir,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	StartedAt       *time.Time   `json:"started_at,omitempty"`
	FinishedAt      *time.Time   `json:"finished_at,omitempty"`
}

type RepoInfo struct {
	Branches []string `json:"branches"`
	Tags     []string `json:"tags"`
}

type TokenClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Type     string `json:"type"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=12,max=128"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	User         User   `json:"user"`
}

type CreateStackRequest struct {
	Name        string   `json:"name" binding:"required,min=1,max=100"`
	Description string   `json:"description" binding:"max=500"`
	GitURL      string   `json:"git_url" binding:"required,url"`
	GitFolder   string   `json:"git_folder"`
	Provider    Provider `json:"provider" binding:"required"`
	Branch      string   `json:"branch"`
	Tags        string   `json:"tags"`
}

type UpdateStackRequest struct {
	Name               *string   `json:"name" binding:"omitempty,min=1,max=100"`
	Description        *string   `json:"description" binding:"omitempty,max=500"`
	GitURL             *string   `json:"git_url" binding:"omitempty,url"`
	GitFolder          *string   `json:"git_folder"`
	Provider           *Provider `json:"provider"`
	PublishedBranches  *[]string `json:"published_branches"`
	PublishedTags      *[]string `json:"published_tags"`
}

type CreateRunRequest struct {
	StackID          string            `json:"stack_id" binding:"required,uuid"`
	RunnerID         string            `json:"runner_id" binding:"required"`
	Branch           string            `json:"branch" binding:"required"`
	CommitSHA        string            `json:"commit_sha" binding:"required"`
	BackendType      string            `json:"backend_type"`
	BackendConfig    map[string]string `json:"backend_config"`
	BackendConfigEnv map[string]string `json:"backend_config_env"`
	TfvarsFiles      []string          `json:"tfvars_files"`
	TfvarsValues     map[string]string `json:"tfvars_values"`
}

type TerraformBackendType string

const (
	TerraformBackendLocal      TerraformBackendType = "local"
	TerraformBackendCloud      TerraformBackendType = "cloud"
	TerraformBackendS3         TerraformBackendType = "s3"
	TerraformBackendAzurerm    TerraformBackendType = "azurerm"
	TerraformBackendGcs        TerraformBackendType = "gcs"
	TerraformBackendKubernetes  TerraformBackendType = "kubernetes"
	TerraformBackendHTTP       TerraformBackendType = "http"
	TerraformBackendConsul     TerraformBackendType = "consul"
	TerraformBackendPg         TerraformBackendType = "pg"
)

func (b TerraformBackendType) IsValid() bool {
	validBackends := []TerraformBackendType{
		TerraformBackendLocal, TerraformBackendCloud, TerraformBackendS3,
		TerraformBackendAzurerm, TerraformBackendGcs, TerraformBackendKubernetes,
		TerraformBackendHTTP, TerraformBackendConsul, TerraformBackendPg,
	}
	for _, v := range validBackends {
		if b == v {
			return true
		}
	}
	return false
}

type CreateBackendRequest struct {
	Name   string                `json:"name" binding:"required,min=1,max=100"`
	Type   TerraformBackendType `json:"type" binding:"required"`
	Config string               `json:"config"`
}

type UpdateBackendRequest struct {
	Name   *string               `json:"name" binding:"omitempty,min=1,max=100"`
	Type   *TerraformBackendType `json:"type"`
	Config *string               `json:"config"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}