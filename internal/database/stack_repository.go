package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/private-tf-runners/server/internal/models"
)

type StackRepository struct {
	db *DB
}

func (r *StackRepository) Create(stack *models.Stack) error {
	query := `INSERT INTO stacks (id, name, description, git_url, git_folder, provider,
			  published_branches, published_tags, tfvars_files, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	branchesJSON, _ := json.Marshal(stack.PublishedBranches)
	tagsJSON, _ := json.Marshal(stack.PublishedTags)
	tfvarsJSON, _ := json.Marshal(stack.TfvarsFiles)
	now := time.Now().UTC()
	_, err := r.db.conn.Exec(query,
		stack.ID, stack.Name, stack.Description, stack.GitURL, stack.GitFolder,
		stack.Provider, string(branchesJSON), string(tagsJSON), string(tfvarsJSON), now, now,
	)
	return err
}

func (r *StackRepository) GetByID(id string) (*models.Stack, error) {
	query := `SELECT id, name, description, git_url, git_folder, provider,
			 published_branches, published_tags, tfvars_files, created_at, updated_at
			 FROM stacks WHERE id = ?`
	return r.scanStack(r.db.conn.QueryRow(query, id))
}

func (r *StackRepository) GetByName(name string) (*models.Stack, error) {
	query := `SELECT id, name, description, git_url, git_folder, provider,
			 published_branches, published_tags, tfvars_files, created_at, updated_at
			 FROM stacks WHERE name = ?`
	return r.scanStack(r.db.conn.QueryRow(query, name))
}

func (r *StackRepository) GetAll() ([]*models.Stack, error) {
	query := `SELECT id, name, description, git_url, git_folder, provider,
			 published_branches, published_tags, tfvars_files, created_at, updated_at
			 FROM stacks ORDER BY created_at DESC`
	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stacks []*models.Stack
	for rows.Next() {
		stack, err := r.scanStackRow(rows)
		if err != nil {
			return nil, err
		}
		stacks = append(stacks, stack)
	}
	return stacks, nil
}

func (r *StackRepository) Update(stack *models.Stack) error {
	query := `UPDATE stacks SET name = ?, description = ?, git_url = ?, git_folder = ?,
			 provider = ?, published_branches = ?, published_tags = ?, tfvars_files = ?, updated_at = ?
			 WHERE id = ?`
	branchesJSON, _ := json.Marshal(stack.PublishedBranches)
	tagsJSON, _ := json.Marshal(stack.PublishedTags)
	tfvarsJSON, _ := json.Marshal(stack.TfvarsFiles)
	now := time.Now().UTC()
	result, err := r.db.conn.Exec(query,
		stack.Name, stack.Description, stack.GitURL, stack.GitFolder,
		stack.Provider, string(branchesJSON), string(tagsJSON), string(tfvarsJSON), now, stack.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *StackRepository) Delete(id string) error {
	query := `DELETE FROM stacks WHERE id = ?`
	result, err := r.db.conn.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *StackRepository) scanStack(row *sql.Row) (*models.Stack, error) {
	stack := &models.Stack{}
	var branchesJSON, tagsJSON, tfvarsJSON string
	err := row.Scan(
		&stack.ID, &stack.Name, &stack.Description, &stack.GitURL, &stack.GitFolder,
		&stack.Provider, &branchesJSON, &tagsJSON, &tfvarsJSON, &stack.CreatedAt, &stack.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(branchesJSON), &stack.PublishedBranches)
	json.Unmarshal([]byte(tagsJSON), &stack.PublishedTags)
	json.Unmarshal([]byte(tfvarsJSON), &stack.TfvarsFiles)
	return stack, nil
}

func (r *StackRepository) scanStackRow(rows *sql.Rows) (*models.Stack, error) {
	stack := &models.Stack{}
	var branchesJSON, tagsJSON, tfvarsJSON string
	err := rows.Scan(
		&stack.ID, &stack.Name, &stack.Description, &stack.GitURL, &stack.GitFolder,
		&stack.Provider, &branchesJSON, &tagsJSON, &tfvarsJSON, &stack.CreatedAt, &stack.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(branchesJSON), &stack.PublishedBranches)
	json.Unmarshal([]byte(tagsJSON), &stack.PublishedTags)
	json.Unmarshal([]byte(tfvarsJSON), &stack.TfvarsFiles)
	return stack, nil
}