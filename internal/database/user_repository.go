package database

import (
	"database/sql"
	"time"

	"github.com/private-tf-runners/server/internal/config"
	"github.com/private-tf-runners/server/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *DB
}

func (r *UserRepository) Create(user *models.User, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), config.Get().Security.GetBCryptCost())
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `INSERT INTO users (id, username, email, password_hash, role, provider, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = r.db.conn.Exec(query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.Role,
		user.Provider, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *UserRepository) Update(user *models.User) error {
	user.UpdatedAt = time.Now()
	query := `UPDATE users SET username = ?, email = ?, role = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, user.Username, user.Email, user.Role, user.UpdatedAt, user.ID)
	return err
}

func (r *UserRepository) Delete(id string) error {
	_, err := r.db.conn.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

func (r *UserRepository) UpdatePassword(id string, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), config.Get().Security.GetBCryptCost())
	if err != nil {
		return err
	}
	_, err = r.db.conn.Exec(`UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`,
		string(hashedPassword), time.Now(), id)
	return err
}

func (r *UserRepository) GetByID(id string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, provider, created_at, updated_at,
			  login_attempts, locked_until, last_login_at, last_login_ip, two_factor_enabled
			  FROM users WHERE id = ?`
	user := &models.User{}
	var lockedUntil, lastLoginAt sql.NullTime
	var lastLoginIP sql.NullString
	err := r.db.conn.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
		&user.Provider, &user.CreatedAt, &user.UpdatedAt,
		&user.LoginAttempts, &lockedUntil, &lastLoginAt, &lastLoginIP, &user.TwoFactorEnabled,
	)
	if err != nil {
		return nil, err
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if lastLoginIP.Valid {
		user.LastLoginIP = lastLoginIP.String
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, provider, created_at, updated_at,
			  login_attempts, locked_until, last_login_at, last_login_ip, two_factor_enabled
			  FROM users WHERE username = ?`
	user := &models.User{}
	var lockedUntil, lastLoginAt sql.NullTime
	var lastLoginIP sql.NullString
	err := r.db.conn.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
		&user.Provider, &user.CreatedAt, &user.UpdatedAt,
		&user.LoginAttempts, &lockedUntil, &lastLoginAt, &lastLoginIP, &user.TwoFactorEnabled,
	)
	if err != nil {
		return nil, err
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if lastLoginIP.Valid {
		user.LastLoginIP = lastLoginIP.String
	}
	return user, nil
}

func (r *UserRepository) UpdateLoginAttempts(id string, attempts int, lockedUntil *time.Time) error {
	query := `UPDATE users SET login_attempts = ?, locked_until = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, attempts, lockedUntil, id)
	return err
}

func (r *UserRepository) UpdateLastLogin(id string, ip string) error {
	query := `UPDATE users SET last_login_at = ?, last_login_ip = ?, login_attempts = 0 WHERE id = ?`
	_, err := r.db.conn.Exec(query, time.Now().UTC(), ip, id)
	return err
}

func (r *UserRepository) GetAll(limit, offset int) ([]*models.User, int64, error) {
	var total int64
	if err := r.db.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, username, email, role, provider, created_at, updated_at, last_login_at, last_login_ip
			  FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var lastLoginAt sql.NullTime
		var lastLoginIP sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.Provider,
			&user.CreatedAt, &user.UpdatedAt, &lastLoginAt, &lastLoginIP); err != nil {
			return nil, 0, err
		}
		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}
		if lastLoginIP.Valid {
			user.LastLoginIP = lastLoginIP.String
		}
		users = append(users, user)
	}
	return users, total, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, provider, created_at, updated_at,
			  login_attempts, locked_until, last_login_at, last_login_ip, two_factor_enabled
			  FROM users WHERE email = ?`
	user := &models.User{}
	var lockedUntil, lastLoginAt sql.NullTime
	var lastLoginIP sql.NullString
	err := r.db.conn.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
		&user.Provider, &user.CreatedAt, &user.UpdatedAt,
		&user.LoginAttempts, &lockedUntil, &lastLoginAt, &lastLoginIP, &user.TwoFactorEnabled,
	)
	if err != nil {
		return nil, err
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if lastLoginIP.Valid {
		user.LastLoginIP = lastLoginIP.String
	}
	return user, nil
}
