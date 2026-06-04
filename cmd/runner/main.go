package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/private-tf-runners/server/internal/models"
)

type RunnerClient struct {
	apiURL     string
	runnerID   string
	token      string
	httpClient *http.Client
	workDir    string
}

func NewRunnerClient(apiURL, runnerID, token string) *RunnerClient {
	return &RunnerClient{
		apiURL:   strings.TrimSuffix(apiURL, "/"),
		runnerID: runnerID,
		token:    token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		workDir: "/tmp/runner-work",
	}
}

func (rc *RunnerClient) heartbeat() error {
	data := map[string]interface{}{
		"status":         models.RunnerStatusOnline,
		"current_run_id": "",
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/runners/%s/heartbeat", rc.apiURL, rc.runnerID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+rc.token)

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed with status: %d", resp.StatusCode)
	}
	return nil
}

func (rc *RunnerClient) getAssignedRun() (*models.Run, error) {
	log.Printf("[DEBUG] getAssignedRun: Requesting runs from /api/runners/%s/runs", rc.runnerID)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/runners/%s/runs", rc.apiURL, rc.runnerID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+rc.token)

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get runs failed with status: %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var runs []models.Run
	if err := json.Unmarshal(body, &runs); err != nil {
		return nil, err
	}

	for i := range runs {
		log.Printf("[DEBUG] getAssignedRun: Found run %s with status %s", runs[i].ID, runs[i].Status)
		if runs[i].Status == models.StatusPending || runs[i].Status == models.StatusApproved || runs[i].Status == models.StatusPlanned {
			log.Printf("[DEBUG] getAssignedRun: Found pending/approved/planned run %s, will claim it", runs[i].ID)
			return &runs[i], nil
		}
	}
	log.Printf("[DEBUG] getAssignedRun: No pending/approved runs found")
	return nil, nil
}

func (rc *RunnerClient) updateRunStatus(runID string, status models.RunStatus, logs string) error {
	path := fmt.Sprintf("%s/api/runner/runs/%s/status", rc.apiURL, runID)

	data := map[string]interface{}{
		"status": status,
		"logs":   logs,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+rc.token)

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update status failed with status: %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (rc *RunnerClient) setPlanOutput(runID, planOutput string) error {
	path := fmt.Sprintf("%s/api/runner/runs/%s/plan-output", rc.apiURL, runID)

	data := map[string]interface{}{
		"plan_output": planOutput,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+rc.token)

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set plan output failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Plan output saved for run %s", runID)
	return nil
}

func (rc *RunnerClient) setWorkDir(runID, workDir string) error {
	path := fmt.Sprintf("%s/api/runner/runs/%s/work-dir", rc.apiURL, runID)

	data := map[string]interface{}{
		"work_dir": workDir,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+rc.token)

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set work dir failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Work dir saved for run %s: %s", runID, workDir)
	return nil
}

func (rc *RunnerClient) setApplyOutput(runID, applyOutput string) error {
	path := fmt.Sprintf("%s/api/runner/runs/%s/apply-output", rc.apiURL, runID)

	data := map[string]interface{}{
		"apply_output": applyOutput,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+rc.token)

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set apply output failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Apply output saved for run %s", runID)
	return nil
}

func (rc *RunnerClient) cloneRepo(repoURL, branch, destDir string) error {
	log.Printf("Cloning repository %s branch %s to %s", repoURL, branch, destDir)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	cmd := exec.Command("git", "clone", "--branch", branch, "--single-branch", repoURL, destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

func (rc *RunnerClient) runPlan(ctx context.Context, run *models.Run, stack *models.Stack, gitFolder string) (string, string, error) {
	provider := "terraform"
	if stack.Provider == models.ProviderOpenTofu {
		provider = "tofu"
	}

	var backendType string
	if run.BackendType != "" {
		backendType = run.BackendType
	} else {
		backendType = "local"
	}

	var backendConfig map[string]string
	if run.BackendKey != "" && run.BackendKey != "{}" {
		json.Unmarshal([]byte(run.BackendKey), &backendConfig)
	}

	var envVars map[string]string
	if run.EnvVars != "" {
		json.Unmarshal([]byte(run.EnvVars), &envVars)
	}

	log.Printf("Running %s init and plan in %s", provider, gitFolder)
	log.Printf("Run TfvarsFiles: %v", run.TfvarsFiles)
	log.Printf("Run TfvarsValues: %s", run.TfvarsValues)
	log.Printf("Run EnvVars: %s", run.EnvVars)
	log.Printf("Backend config: %v, backend type: %s", backendConfig, backendType)

	var initArgs, planArgs []string
	planFile := filepath.Join(gitFolder, "tfplan")
	initArgs = append(initArgs, "init")
	if backendType != "local" {
		for k, v := range backendConfig {
			if k != "type" && k != "update_address" && v != "" {
				initArgs = append(initArgs, fmt.Sprintf("-backend-config=%s=%s", k, v))
			}
		}
	}

	planArgs = append(planArgs, "plan", "-out="+planFile)

	for _, tfvarsFile := range run.TfvarsFiles {
		if tfvarsFile != "" {
			planArgs = append(planArgs, "-var-file="+tfvarsFile)
		}
	}
	if len(run.TfvarsFiles) > 0 {
		log.Printf("Plan args after tfvars: %v", planArgs)
	}

	if run.TfvarsValues != "" {
		var tfvarsValues map[string]string
		json.Unmarshal([]byte(run.TfvarsValues), &tfvarsValues)
		for k, v := range tfvarsValues {
			planArgs = append(planArgs, fmt.Sprintf("-var=%s=%s", k, v))
		}
	}

	var allOutput bytes.Buffer

	log.Printf("Running init: %s %v", provider, initArgs)
	initCmd := exec.CommandContext(ctx, provider, initArgs...)
	initCmd.Dir = gitFolder
	initCmd.Env = os.Environ()
	for k, v := range envVars {
		initCmd.Env = append(initCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	initCmd.Stdout = &allOutput
	initCmd.Stderr = &allOutput

	if err := initCmd.Run(); err != nil {
		log.Printf("Init failed, output: %s", allOutput.String())
		rc.setPlanOutput(run.ID, allOutput.String())
		return allOutput.String(), planFile, fmt.Errorf("%s init failed: %w", provider, err)
	}

	log.Printf("Init succeeded, output: %s", allOutput.String())
	rc.setPlanOutput(run.ID, allOutput.String())

	log.Printf("Running plan: %s %v", provider, planArgs)
	planCmd := exec.CommandContext(ctx, provider, planArgs...)
	planCmd.Dir = gitFolder
	planCmd.Env = os.Environ()
	for k, v := range envVars {
		planCmd.Env = append(planCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	planCmd.Stdout = &allOutput
	planCmd.Stderr = &allOutput

	if err := planCmd.Run(); err != nil {
		rc.setPlanOutput(run.ID, allOutput.String())
		return allOutput.String(), planFile, fmt.Errorf("%s plan failed: %w", provider, err)
	}

	return allOutput.String(), planFile, nil
}

func (rc *RunnerClient) runApply(ctx context.Context, run *models.Run, stack *models.Stack, planFile string) (string, error) {
	var workDir string
	if run.WorkDir != "" {
		workDir = run.WorkDir
	} else {
		workDir = filepath.Join(rc.workDir, "repos", fmt.Sprintf("%s-%s-%s", run.StackID, run.Branch, run.ID[:8]))
	}

	gitFolder := workDir
	if stack.GitFolder != "" {
		gitFolder = filepath.Join(workDir, stack.GitFolder)
	}

	provider := "terraform"
	if stack.Provider == models.ProviderOpenTofu {
		provider = "tofu"
	}

	var backendType string
	if run.BackendType != "" {
		backendType = run.BackendType
	} else {
		backendType = "local"
	}

	var backendConfig map[string]string
	if run.BackendKey != "" && run.BackendKey != "{}" {
		json.Unmarshal([]byte(run.BackendKey), &backendConfig)
	}

	var envVars map[string]string
	if run.EnvVars != "" {
		json.Unmarshal([]byte(run.EnvVars), &envVars)
	}

	log.Printf("Running %s apply in %s", provider, gitFolder)

	var allOutput bytes.Buffer

	initArgs := []string{"init"}
	if backendType != "local" {
		for k, v := range backendConfig {
			if k != "type" && k != "update_address" && v != "" {
				initArgs = append(initArgs, fmt.Sprintf("-backend-config=%s=%s", k, v))
			}
		}
	}

	log.Printf("Running init: %s %v", provider, initArgs)
	initCmd := exec.CommandContext(ctx, provider, initArgs...)
	initCmd.Dir = gitFolder
	initCmd.Env = os.Environ()
	for k, v := range envVars {
		initCmd.Env = append(initCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	initCmd.Stdout = &allOutput
	initCmd.Stderr = &allOutput

	if err := initCmd.Run(); err != nil {
		return allOutput.String(), fmt.Errorf("%s init failed: %w", provider, err)
	}

	applyArgs := []string{"apply", "-auto-approve", planFile}

	log.Printf("Running apply: %s %v", provider, applyArgs)

	applyCmd := exec.CommandContext(ctx, provider, applyArgs...)
	applyCmd.Dir = gitFolder
	applyCmd.Env = os.Environ()
	for k, v := range envVars {
		applyCmd.Env = append(applyCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
applyCmd.Stdout = &allOutput
	applyCmd.Stderr = &allOutput

	applyDone := make(chan error, 1)
	go func() {
		applyDone <- applyCmd.Run()
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case err := <-applyDone:
			if err != nil {
				rc.setApplyOutput(run.ID, allOutput.String())
				return allOutput.String(), fmt.Errorf("%s apply failed: %w", provider, err)
			}
			goto getOutputs
		case <-ticker.C:
			rc.setApplyOutput(run.ID, allOutput.String())
		}
	}

getOutputs:
	outputCmd := exec.CommandContext(ctx, provider, "output", "-json")
	outputCmd.Dir = gitFolder
	outputCmd.Env = os.Environ()
	for k, v := range envVars {
		outputCmd.Env = append(outputCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	outputResult, err := outputCmd.CombinedOutput()
	if err != nil {
		allOutput.WriteString("\n\n[No outputs or failed to get outputs]")
	} else {
		allOutput.WriteString("\n\n=== Terraform Outputs ===\n")
		allOutput.Write(outputResult)
	}

	rc.setApplyOutput(run.ID, allOutput.String())
	return allOutput.String(), nil
}

func (rc *RunnerClient) ClaimRun(ctx context.Context) error {
	run, err := rc.getAssignedRun()
	if err != nil {
		return err
	}

	if run == nil {
		return nil
	}

	stack, err := rc.getStack(run.StackID)
	if err != nil {
		return err
	}

	log.Printf("Running %s plan phase for run %s", run.Branch, run.ID)

	var workDir string
	if run.WorkDir != "" {
		workDir = run.WorkDir
		log.Printf("Reusing work dir from run: %s", workDir)
	} else {
		workDir = filepath.Join(rc.workDir, "repos", fmt.Sprintf("%s-%s-%s", run.StackID, run.Branch, run.ID[:8]))
		log.Printf("Created new work dir: %s", workDir)
		if err := rc.setWorkDir(run.ID, workDir); err != nil {
			log.Printf("Failed to save work dir: %v", err)
		}
	}

	gitFolder := workDir
	if stack.GitFolder != "" {
		gitFolder = filepath.Join(workDir, stack.GitFolder)
	}
	defer func() {
		if run.WorkDir == "" {
			log.Printf("Cleaning up work dir: %s", workDir)
			os.RemoveAll(workDir)
		} else {
			log.Printf("Keeping work dir for apply phase: %s", workDir)
		}
	}()

	if err := rc.updateRunStatus(run.ID, models.StatusRunning, "Cloning repository..."); err != nil {
		log.Printf("Failed to update run status: %v", err)
	}

	repoURL := fmt.Sprintf("https://%s@%s", rc.token, strings.Replace(stack.GitURL, "https://", "", 1))
	if err := rc.cloneRepo(repoURL, run.Branch, gitFolder); err != nil {
		errMsg := fmt.Sprintf("Failed to clone repository: %v", err)
		rc.updateRunStatus(run.ID, models.StatusFailed, errMsg)
		return fmt.Errorf(errMsg)
	}

	if err := rc.updateRunStatus(run.ID, models.StatusRunning, "Running init and plan..."); err != nil {
		log.Printf("Failed to update run status: %v", err)
	}

	planOutput, planFile, err := rc.runPlan(ctx, run, stack, gitFolder)
	if err != nil {
		errMsg := fmt.Sprintf("Plan failed: %v", err)
		log.Printf("%s", errMsg)
		rc.updateRunStatus(run.ID, models.StatusFailed, planOutput+"\n\n"+errMsg)
		rc.setPlanOutput(run.ID, planOutput)
		os.RemoveAll(workDir)
		return nil
	}

	if err := rc.setPlanOutput(run.ID, planOutput); err != nil {
		log.Printf("Failed to save plan output: %v", err)
	}

	if err := rc.updateRunStatus(run.ID, models.StatusPlanned, "Plan completed. Waiting for approval..."); err != nil {
		log.Printf("Failed to update run status: %v", err)
	}

	for {
		run, err := rc.getAssignedRun()
		if err != nil {
			log.Printf("Failed to get run: %v, retrying...", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
			continue
		}
		if run == nil {
			log.Printf("No approved run yet, waiting...")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
			continue
		}

		switch run.Status {
		case models.StatusApproved:
			log.Printf("Run %s approved, starting apply", run.ID)
			goto approved
		case models.StatusRejected:
			log.Printf("Run %s rejected", run.ID)
			return nil
		case models.StatusPending:
		case models.StatusPlanned:
			log.Printf("Run %s status: %s, waiting for approval", run.ID, run.Status)
		default:
			log.Printf("Run %s status: %s", run.ID, run.Status)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}

approved:

	log.Printf("Running %s apply phase for run %s", run.Branch, run.ID)

	if err := rc.updateRunStatus(run.ID, models.StatusRunning, "Running apply..."); err != nil {
		log.Printf("Failed to update run status: %v", err)
	}

	applyOutput, err := rc.runApply(ctx, run, stack, planFile)
	if err != nil {
		errMsg := fmt.Sprintf("Apply failed: %v", err)
		log.Printf("%s", errMsg)
		rc.updateRunStatus(run.ID, models.StatusFailed, applyOutput+"\n\n"+errMsg)
		rc.setApplyOutput(run.ID, applyOutput)
		return nil
	}

	log.Printf("=== APPLY COMPLETE for run %s, updating status to Applied ===", run.ID)
	err = rc.updateRunStatus(run.ID, models.StatusApplied, "Apply completed successfully")
	if err != nil {
		log.Printf("=== updateRunStatus FAILED: %v", err)
	} else {
		log.Printf("=== updateRunStatus SUCCESS")
	}
	rc.setApplyOutput(run.ID, applyOutput)

	log.Printf("=== APPLY COMPLETE for run %s ===", run.ID)
	return nil
}

func (rc *RunnerClient) getStack(stackID string) (*models.Stack, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/runner/stacks/%s", rc.apiURL, stackID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+rc.token)

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get stack failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var stack models.Stack
	if err := json.Unmarshal(body, &stack); err != nil {
		return nil, err
	}

	return &stack, nil
}

func (rc *RunnerClient) Start(ctx context.Context) error {
	log.Printf("Runner %s starting, connecting to %s", rc.runnerID, rc.apiURL)

	for {
		if err := rc.heartbeat(); err != nil {
			log.Printf("Failed to send heartbeat: %v", err)
		}

		if err := rc.ClaimRun(ctx); err != nil {
			log.Printf("Failed to claim run: %v", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(30 * time.Second):
		}
	}
}
func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	rc := NewRunnerClient(cfg.APIURL, cfg.RunnerID, cfg.Token)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := rc.Start(ctx); err != nil {
		log.Fatalf("Runner error: %v", err)
	}
}

type Config struct {
	APIURL   string
	RunnerID string
	Token    string
}

func loadConfig() (*Config, error) {
	apiURL := os.Getenv("RUNNER_API_URL")
	runnerID := os.Getenv("RUNNER_ID")
	token := os.Getenv("RUNNER_TOKEN")

	if apiURL == "" || runnerID == "" || token == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	return &Config{
		APIURL:   apiURL,
		RunnerID: runnerID,
		Token:    token,
	}, nil
}
