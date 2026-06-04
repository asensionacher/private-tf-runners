package database

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/private-tf-runners/server/internal/models"
)

type BackendRepository struct {
	db *DB
}

func (r *BackendRepository) Create(backend *models.Backend) error {
	backend.ID = generateID()
	backend.CreatedAt = time.Now()
	backend.UpdatedAt = time.Now()

	_, err := r.db.conn.Exec(
		`INSERT INTO backends (id, name, type, config, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		backend.ID, backend.Name, backend.Type, backend.Config, backend.CreatedAt, backend.UpdatedAt,
	)
	return err
}

func (r *BackendRepository) GetAll() ([]models.Backend, error) {
	rows, err := r.db.conn.Query(`SELECT id, name, type, config, created_at, updated_at FROM backends ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backends []models.Backend
	for rows.Next() {
		var b models.Backend
		if err := rows.Scan(&b.ID, &b.Name, &b.Type, &b.Config, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		backends = append(backends, b)
	}
	return backends, rows.Err()
}

func (r *BackendRepository) GetByID(id string) (*models.Backend, error) {
	var b models.Backend
	err := r.db.conn.QueryRow(
		`SELECT id, name, type, config, created_at, updated_at FROM backends WHERE id = ?`,
		id,
	).Scan(&b.ID, &b.Name, &b.Type, &b.Config, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BackendRepository) Update(backend *models.Backend) error {
	backend.UpdatedAt = time.Now()
	_, err := r.db.conn.Exec(
		`UPDATE backends SET name = ?, type = ?, config = ?, updated_at = ? WHERE id = ?`,
		backend.Name, backend.Type, backend.Config, backend.UpdatedAt, backend.ID,
	)
	return err
}

func (r *BackendRepository) Delete(id string) error {
	_, err := r.db.conn.Exec(`DELETE FROM backends WHERE id = ?`, id)
	return err
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}