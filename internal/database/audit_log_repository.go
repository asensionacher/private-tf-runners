package database

import (
	"time"

	"github.com/google/uuid"

	"github.com/private-tf-runners/server/internal/models"
)

type AuditLogRepository struct {
	db *DB
}

func (r *AuditLogRepository) Create(log *models.AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	query := `INSERT INTO audit_logs (id, user_id, user_email, action, resource, resource_id, details, ip, user_agent, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.conn.Exec(query,
		log.ID, log.UserID, log.UserEmail, log.Action, log.Resource,
		log.ResourceID, log.Details, log.IP, log.UserAgent, time.Now().UTC(),
	)
	return err
}

func (r *AuditLogRepository) GetByUserID(userID string, limit int) ([]*models.AuditLog, error) {
	query := `SELECT id, user_id, user_email, action, resource, resource_id, details, ip, user_agent, created_at
			  FROM audit_logs WHERE user_id = ? ORDER BY created_at DESC LIMIT ?`
	rows, err := r.db.conn.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		log := &models.AuditLog{}
		if err := rows.Scan(&log.ID, &log.UserID, &log.UserEmail, &log.Action, &log.Resource,
			&log.ResourceID, &log.Details, &log.IP, &log.UserAgent, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *AuditLogRepository) GetAll(limit, offset int) ([]*models.AuditLog, int64, error) {
	countQuery := `SELECT COUNT(*) FROM audit_logs`
	var total int64
	if err := r.db.conn.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, user_id, user_email, action, resource, resource_id, details, ip, user_agent, created_at
			  FROM audit_logs ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		log := &models.AuditLog{}
		if err := rows.Scan(&log.ID, &log.UserID, &log.UserEmail, &log.Action, &log.Resource,
			&log.ResourceID, &log.Details, &log.IP, &log.UserAgent, &log.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}
	return logs, total, nil
}