package database

import (
	"database/sql"
	"time"

	"github.com/private-tf-runners/server/internal/models"
)

type RunnerRepository struct {
	db *DB
}

func (r *RunnerRepository) Create(runner *models.Runner) error {
	runner.ID = generateID()
	runner.CreatedAt = time.Now()

	_, err := r.db.conn.Exec(
		`INSERT INTO runners (id, name, token, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		runner.ID, runner.Name, runner.Token, runner.Status, runner.CreatedAt,
	)
	return err
}

func (r *RunnerRepository) GetAll() ([]models.Runner, error) {
	rows, err := r.db.conn.Query(`SELECT id, name, token, status, current_run_id, last_seen, created_at FROM runners ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runners []models.Runner
	for rows.Next() {
		var runner models.Runner
		var currentRunID sql.NullString
		var lastSeen sql.NullTime
		if err := rows.Scan(&runner.ID, &runner.Name, &runner.Token, &runner.Status, &currentRunID, &lastSeen, &runner.CreatedAt); err != nil {
			return nil, err
		}
		if currentRunID.Valid {
			runner.CurrentRunID = currentRunID.String
		}
		if lastSeen.Valid {
			runner.LastSeen = &lastSeen.Time
		}
		runners = append(runners, runner)
	}
	return runners, rows.Err()
}

func (r *RunnerRepository) GetByID(id string) (*models.Runner, error) {
	runner := &models.Runner{}
	var currentRunID sql.NullString
	var lastSeen sql.NullTime
	err := r.db.conn.QueryRow(
		`SELECT id, name, token, status, current_run_id, last_seen, created_at FROM runners WHERE id = ?`,
		id,
	).Scan(&runner.ID, &runner.Name, &runner.Token, &runner.Status, &currentRunID, &lastSeen, &runner.CreatedAt)
	if err != nil {
		return nil, err
	}
	if currentRunID.Valid {
		runner.CurrentRunID = currentRunID.String
	}
	if lastSeen.Valid {
		runner.LastSeen = &lastSeen.Time
	}
	return runner, nil
}

func (r *RunnerRepository) GetByToken(token string) (*models.Runner, error) {
	runner := &models.Runner{}
	var currentRunID sql.NullString
	var lastSeen sql.NullTime
	err := r.db.conn.QueryRow(
		`SELECT id, name, token, status, current_run_id, last_seen, created_at FROM runners WHERE token = ?`,
		token,
	).Scan(&runner.ID, &runner.Name, &runner.Token, &runner.Status, &currentRunID, &lastSeen, &runner.CreatedAt)
	if err != nil {
		return nil, err
	}
	if currentRunID.Valid {
		runner.CurrentRunID = currentRunID.String
	}
	if lastSeen.Valid {
		runner.LastSeen = &lastSeen.Time
	}
	return runner, nil
}

func (r *RunnerRepository) Update(runner *models.Runner) error {
	_, err := r.db.conn.Exec(
		`UPDATE runners SET name = ?, token = ?, status = ?, current_run_id = ?, last_seen = ? WHERE id = ?`,
		runner.Name, runner.Token, runner.Status, runner.CurrentRunID, runner.LastSeen, runner.ID,
	)
	return err
}

func (r *RunnerRepository) UpdateStatus(id string, status models.RunnerStatus, currentRunID string) error {
	var err error
	if currentRunID != "" {
		_, err = r.db.conn.Exec(
			`UPDATE runners SET status = ?, current_run_id = ?, last_seen = ? WHERE id = ?`,
			status, currentRunID, time.Now(), id,
		)
	} else {
		_, err = r.db.conn.Exec(
			`UPDATE runners SET status = ?, current_run_id = NULL, last_seen = ? WHERE id = ?`,
			status, time.Now(), id,
		)
	}
	return err
}

func (r *RunnerRepository) Heartbeat(id string, status models.RunnerStatus, currentRunID string) error {
	var err error
	if currentRunID != "" {
		_, err = r.db.conn.Exec(
			`UPDATE runners SET status = ?, current_run_id = ?, last_seen = ? WHERE id = ?`,
			status, currentRunID, time.Now(), id,
		)
	} else {
		_, err = r.db.conn.Exec(
			`UPDATE runners SET status = ?, last_seen = ? WHERE id = ?`,
			status, time.Now(), id,
		)
	}
	return err
}

func (r *RunnerRepository) MarkStaleOffline(staleness time.Duration) error {
	cutoff := time.Now().Add(-staleness)
	_, err := r.db.conn.Exec(
		`UPDATE runners SET status = ? WHERE status = ? AND last_seen < ?`,
		models.RunnerStatusOffline, models.RunnerStatusOnline, cutoff,
	)
	return err
}

func (r *RunnerRepository) Delete(id string) error {
	_, err := r.db.conn.Exec(`DELETE FROM runners WHERE id = ?`, id)
	return err
}

func (r *RunnerRepository) GetByStatus(status models.RunnerStatus) ([]models.Runner, error) {
	rows, err := r.db.conn.Query(`SELECT id, name, token, status, current_run_id, last_seen, created_at FROM runners WHERE status = ? ORDER BY name`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runners []models.Runner
	for rows.Next() {
		var runner models.Runner
		var currentRunID sql.NullString
		var lastSeen sql.NullTime
		if err := rows.Scan(&runner.ID, &runner.Name, &runner.Token, &runner.Status, &currentRunID, &lastSeen, &runner.CreatedAt); err != nil {
			return nil, err
		}
		if currentRunID.Valid {
			runner.CurrentRunID = currentRunID.String
		}
		if lastSeen.Valid {
			runner.LastSeen = &lastSeen.Time
		}
		runners = append(runners, runner)
	}
	return runners, rows.Err()
}