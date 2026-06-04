package models

import (
	"time"
)

type RunnerStatus string

const (
	RunnerStatusOnline  RunnerStatus = "online"
	RunnerStatusOffline RunnerStatus = "offline"
	RunnerStatusBusy    RunnerStatus = "busy"
)

type Runner struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Token        string       `json:"-"`
	Status       RunnerStatus `json:"status"`
	CurrentRunID string       `json:"current_run_id,omitempty"`
	LastSeen     *time.Time   `json:"last_seen,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}

type CreateRunnerRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type RunnerCreatedResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

type RunnerHeartbeatRequest struct {
	Status       RunnerStatus `json:"status"`
	CurrentRunID string       `json:"current_run_id,omitempty"`
}

type UpdateRunnerRequest struct {
	Name   *string `json:"name" binding:"omitempty,min=1,max=100"`
	Status *string `json:"status"`
}

type RunWithRunner struct {
	Run
	RunnerName string `json:"runner_name,omitempty"`
	RunnerID   string `json:"runner_id,omitempty"`
}