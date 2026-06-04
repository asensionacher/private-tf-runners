package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/private-tf-runners/server/internal/models"
)

type RunRepository struct {
	db *DB
}

func (r *RunRepository) Create(run *models.Run) error {
	query := `INSERT INTO runs (id, stack_id, trigger_type, branch, commit_sha, status, phase, logs,
			  backend_type, backend_key, tfvars_files, tfvars_values, env_vars, commands, runner_id, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now().UTC()
	tfvarsFilesJSON, _ := json.Marshal(run.TfvarsFiles)
	
	// Convert empty runner_id to NULL for proper SQL behavior
	var runnerID interface{} = run.RunnerID
	if runnerID != nil && runnerID.(string) == "" {
		runnerID = nil
	}
	
	_, err := r.db.conn.Exec(query,
		run.ID, run.StackID, run.TriggerType, run.Branch, run.CommitSHA,
		run.Status, run.Phase, run.Logs, run.BackendType, run.BackendKey,
		string(tfvarsFilesJSON), run.TfvarsValues, run.EnvVars, run.Commands, runnerID, now,
	)
	return err
}

func (r *RunRepository) GetByID(id string) (*models.Run, error) {
	query := `SELECT r.id, r.stack_id, s.name as stack_name, r.trigger_type, r.branch, r.commit_sha,
			 r.status, r.phase, r.logs, r.plan_output, r.apply_output, r.backend_type, r.backend_key, r.tfvars_files,
			 r.tfvars_values, r.env_vars, r.commands, r.runner_id, r.created_at, r.started_at, r.finished_at,
			 rn.name as runner_name, r.work_dir
			 FROM runs r
			 LEFT JOIN stacks s ON r.stack_id = s.id
			 LEFT JOIN runners rn ON r.runner_id = rn.id
			 WHERE r.id = ?`
	return r.scanRun(r.db.conn.QueryRow(query, id))
}

func (r *RunRepository) GetAll(limit, offset int) ([]*models.Run, int64, error) {
	countQuery := `SELECT COUNT(*) FROM runs`
	var total int64
	if err := r.db.conn.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT r.id, r.stack_id, s.name as stack_name, r.trigger_type, r.branch, r.commit_sha,
			 r.status, r.phase, r.logs, r.plan_output, r.apply_output, r.backend_type, r.backend_key, r.tfvars_files,
			 r.tfvars_values, r.env_vars, r.commands, r.runner_id, r.created_at, r.started_at, r.finished_at,
			 rn.name as runner_name, r.work_dir
			 FROM runs r
			 LEFT JOIN stacks s ON r.stack_id = s.id
			 LEFT JOIN runners rn ON r.runner_id = rn.id
			 ORDER BY r.created_at DESC
			 LIMIT ? OFFSET ?`
	rows, err := r.db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var runs []*models.Run
	for rows.Next() {
		run, err := r.scanRunRow(rows)
		if err != nil {
			return nil, 0, err
		}
		runs = append(runs, run)
	}
	return runs, total, nil
}

func (r *RunRepository) GetByStackID(stackID string) ([]*models.Run, error) {
query := `SELECT r.id, r.stack_id, s.name as stack_name, r.trigger_type, r.branch, r.commit_sha,
			 r.status, r.phase, r.logs, r.plan_output, r.apply_output, r.backend_type, r.backend_key, r.tfvars_files,
			 r.tfvars_values, r.env_vars, r.commands, r.runner_id, r.created_at, r.started_at, r.finished_at,
			 rn.name as runner_name, r.work_dir
			 FROM runs r
			 LEFT JOIN stacks s ON r.stack_id = s.id
			 LEFT JOIN runners rn ON r.runner_id = rn.id
			 WHERE r.stack_id = ?
			 ORDER BY r.created_at DESC`
	rows, err := r.db.conn.Query(query, stackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*models.Run
	for rows.Next() {
		run, err := r.scanRunRow(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}

func (r *RunRepository) GetByRunnerID(runnerID string) ([]*models.Run, error) {
	query := `SELECT r.id, r.stack_id, s.name as stack_name, r.trigger_type, r.branch, r.commit_sha,
			 r.status, r.phase, r.logs, r.plan_output, r.apply_output, r.backend_type, r.backend_key, r.tfvars_files,
			 r.tfvars_values, r.env_vars, r.commands, r.runner_id, r.created_at, r.started_at, r.finished_at,
			 rn.name as runner_name, r.work_dir
			 FROM runs r
			 LEFT JOIN stacks s ON r.stack_id = s.id
			 LEFT JOIN runners rn ON r.runner_id = rn.id
			 WHERE r.runner_id = ?
			 ORDER BY r.created_at DESC
			 LIMIT 100`

	rows, err := r.db.conn.Query(query, runnerID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var runs []*models.Run
	for rows.Next() {
		run, err := r.scanRunRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		runs = append(runs, run)
	}
	
	fmt.Printf("[DEBUG] GetByRunnerID: Found %d rows\n", len(runs))
	return runs, nil
}

func (r *RunRepository) GetPendingForRunner(runnerID string) ([]*models.Run, error) {
	query := `SELECT r.id, r.stack_id, s.name as stack_name, r.trigger_type, r.branch, r.commit_sha,
			 r.status, r.phase, r.logs, r.plan_output, r.apply_output, r.backend_type, r.backend_key, r.tfvars_files,
			 r.tfvars_values, r.env_vars, r.commands, r.runner_id, r.created_at, r.started_at, r.finished_at,
			 rn.name as runner_name, r.work_dir
			 FROM runs r
			 LEFT JOIN stacks s ON r.stack_id = s.id
			 LEFT JOIN runners rn ON r.runner_id = rn.id
			 WHERE r.runner_id = ? AND r.status IN ('pending', 'approved', 'planned') AND r.phase IN ('plan', 'apply')
			 ORDER BY r.created_at ASC
			 LIMIT 1`
	rows, err := r.db.conn.Query(query, runnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*models.Run
	for rows.Next() {
		run, err := r.scanRunRow(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}

func (r *RunRepository) UpdateStatus(id string, status models.RunStatus) error {
	query := `UPDATE runs SET status = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, status, id)
	return err
}

func (r *RunRepository) UpdateStatusWithLogs(id string, status models.RunStatus, logs string) error {
	query := `UPDATE runs SET status = ?, logs = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, status, logs, id)
	return err
}

func (r *RunRepository) UpdatePhase(id string, phase models.RunPhase) error {
	query := `UPDATE runs SET phase = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, phase, id)
	return err
}

func (r *RunRepository) AssignRunner(runID, runnerID string) error {
	query := `UPDATE runs SET runner_id = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, runnerID, runID)
	return err
}

func (r *RunRepository) SetWorkDir(id string, workDir string) error {
	query := `UPDATE runs SET work_dir = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, workDir, id)
	return err
}

func (r *RunRepository) SetPlanOutput(id string, planOutput string) error {
	query := `UPDATE runs SET plan_output = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, planOutput, id)
	return err
}

func (r *RunRepository) SetApplyOutput(id string, applyOutput string) error {
	query := `UPDATE runs SET apply_output = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, applyOutput, id)
	return err
}

func (r *RunRepository) Start(id string) error {
	query := `UPDATE runs SET status = ?, started_at = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, models.StatusRunning, time.Now().UTC(), id)
	return err
}

func (r *RunRepository) Finish(id string, status models.RunStatus, logs string) error {
	query := `UPDATE runs SET status = ?, finished_at = ?, logs = ?, phase = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, status, time.Now().UTC(), logs, models.PhaseFinish, id)
	return err
}

func (r *RunRepository) scanRun(row *sql.Row) (*models.Run, error) {
	run := &models.Run{}
	var startedAt, finishedAt sql.NullTime
	var stackName, backendType, backendKey, tfvarsFiles, tfvarsValues, envVars, commands, runnerID, runnerName, planOutput, applyOutput sql.NullString
	var phase, workDir sql.NullString
	err := row.Scan(
		&run.ID, &run.StackID, &stackName, &run.TriggerType, &run.Branch,
		&run.CommitSHA, &run.Status, &phase, &run.Logs, &planOutput, &applyOutput, &backendType, &backendKey,
		&tfvarsFiles, &tfvarsValues, &envVars, &commands, &runnerID,
		&run.CreatedAt, &startedAt, &finishedAt, &runnerName, &workDir,
	)
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		run.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		run.FinishedAt = &finishedAt.Time
	}
	if stackName.Valid {
		run.StackName = stackName.String
	}
	if runnerName.Valid {
		run.RunnerName = runnerName.String
	}
	if runnerID.Valid {
		run.RunnerID = runnerID.String
	}
	if phase.Valid {
		run.Phase = models.RunPhase(phase.String)
	}
	if planOutput.Valid {
		run.PlanOutput = planOutput.String
	}
	if applyOutput.Valid {
		run.ApplyOutput = applyOutput.String
	}
	if backendType.Valid {
		run.BackendType = backendType.String
	}
	if backendKey.Valid {
		run.BackendKey = backendKey.String
	}
	if tfvarsFiles.Valid && tfvarsFiles.String != "" && tfvarsFiles.String != "[]" {
		json.Unmarshal([]byte(tfvarsFiles.String), &run.TfvarsFiles)
	}
	if tfvarsValues.Valid {
		run.TfvarsValues = tfvarsValues.String
	}
	if envVars.Valid {
		run.EnvVars = envVars.String
	}
	if commands.Valid {
		run.Commands = commands.String
	}
	return run, nil
}

func (r *RunRepository) scanRunRow(rows *sql.Rows) (*models.Run, error) {
	run := &models.Run{}
	var startedAt, finishedAt sql.NullTime
	var stackName, backendType, backendKey, tfvarsFiles, tfvarsValues, envVars, commands, runnerID, runnerName, planOutput, applyOutput, workDir sql.NullString
	var phase sql.NullString
	var workDirDest string
	err := rows.Scan(
		&run.ID, &run.StackID, &stackName, &run.TriggerType, &run.Branch,
		&run.CommitSHA, &run.Status, &phase, &run.Logs, &planOutput, &applyOutput, &backendType, &backendKey,
		&tfvarsFiles, &tfvarsValues, &envVars, &commands, &runnerID,
		&run.CreatedAt, &startedAt, &finishedAt, &runnerName, &workDirDest,
	)
	if err != nil {
		return nil, err
	}
	if workDirDest != "" {
		run.WorkDir = workDirDest
	}
	if startedAt.Valid {
		run.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		run.FinishedAt = &finishedAt.Time
	}
	if stackName.Valid {
		run.StackName = stackName.String
	}
	if runnerName.Valid {
		run.RunnerName = runnerName.String
	}
	if runnerID.Valid {
		run.RunnerID = runnerID.String
	}
	if phase.Valid {
		run.Phase = models.RunPhase(phase.String)
	}
	if planOutput.Valid {
		run.PlanOutput = planOutput.String
	}
	if applyOutput.Valid {
		run.ApplyOutput = applyOutput.String
	}
	if backendType.Valid {
		run.BackendType = backendType.String
	}
	if backendKey.Valid {
		run.BackendKey = backendKey.String
	}
	if tfvarsFiles.Valid && tfvarsFiles.String != "" && tfvarsFiles.String != "[]" {
		json.Unmarshal([]byte(tfvarsFiles.String), &run.TfvarsFiles)
	}
	if tfvarsValues.Valid {
		run.TfvarsValues = tfvarsValues.String
	}
	if envVars.Valid {
		run.EnvVars = envVars.String
	}
	if commands.Valid {
		run.Commands = commands.String
	}
	if workDir.Valid {
		run.WorkDir = workDir.String
	}
	return run, nil
}