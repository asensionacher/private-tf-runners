package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"

	"github.com/private-tf-runners/server/internal/config"
	"github.com/private-tf-runners/server/internal/models"
)

type DB struct {
	conn *sql.DB
}

func New(cfg *config.DatabaseConfig) (*DB, error) {
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", cfg.Path+"?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(time.Hour)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	if err := db.seedDefaultUser(); err != nil {
		return nil, fmt.Errorf("failed to seed default user: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			provider TEXT DEFAULT 'local',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			login_attempts INTEGER DEFAULT 0,
			locked_until DATETIME,
			last_login_at DATETIME,
			last_login_ip TEXT,
			two_factor_enabled INTEGER DEFAULT 0,
			two_factor_secret TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			user_email TEXT,
			action TEXT NOT NULL,
			resource TEXT NOT NULL,
			resource_id TEXT,
			details TEXT,
			ip TEXT,
			user_agent TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS stacks (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			git_url TEXT NOT NULL,
			git_folder TEXT DEFAULT '',
			provider TEXT NOT NULL CHECK(provider IN ('opentofu', 'terraform')),
			published_branches TEXT DEFAULT '[]',
			published_tags TEXT DEFAULT '[]',
			backend_id TEXT,
			use_stack_backend INTEGER DEFAULT 0,
			tfvars_files TEXT DEFAULT '[]',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS runners (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			token TEXT NOT NULL,
			status TEXT DEFAULT 'offline' CHECK(status IN ('online', 'offline', 'busy')),
			current_run_id TEXT,
			last_seen DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS runs (
			id TEXT PRIMARY KEY,
			stack_id TEXT NOT NULL,
			trigger_type TEXT NOT NULL CHECK(trigger_type IN ('manual', 'push', 'schedule')),
			branch TEXT NOT NULL,
			commit_sha TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'running', 'success', 'failed', 'approved', 'planned', 'applied', 'rejected')),
			phase TEXT DEFAULT 'plan' CHECK(phase IN ('plan', 'apply', 'finish')),
			logs TEXT DEFAULT '',
			plan_output TEXT DEFAULT '',
			backend_id TEXT,
			backend_key TEXT,
			backend_type TEXT DEFAULT 'local',
			use_stack_backend INTEGER DEFAULT 0,
			tfvars_values TEXT DEFAULT '',
			tfvars_files TEXT DEFAULT '[]',
			env_vars TEXT DEFAULT '',
			commands TEXT DEFAULT '',
			runner_id TEXT,
			work_dir TEXT DEFAULT '',
			apply_output TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			finished_at DATETIME,
			FOREIGN KEY (stack_id) REFERENCES stacks(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_stack_id ON runs(stack_id)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_runner_id ON runs(runner_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_runners_status ON runners(status)`,
		`CREATE TRIGGER IF NOT EXISTS trg_updated_at_stacks AFTER UPDATE ON stacks BEGIN UPDATE stacks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id; END`,
		`CREATE TRIGGER IF NOT EXISTS trg_updated_at_users AFTER UPDATE ON users BEGIN UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id; END`,
	}

	for i, migration := range migrations {
		if _, err := db.conn.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	return nil
}

func (db *DB) seedDefaultUser() error {
	var count int
	if err := db.conn.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", "admin").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	passwordHash, err := hashPassword("Admin123@Test")
	if err != nil {
		return err
	}

	id := uuid.New().String()
	_, err = db.conn.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, provider) VALUES (?, ?, ?, ?, ?, ?)`,
		id, "admin", "admin@example.com", passwordHash, string(models.RoleAdmin), "local",
	)
	return err
}

func hashPassword(password string) (string, error) {
	cfg := config.Get()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.Security.GetBCryptCost())
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (d *DB) User() *UserRepository {
	return &UserRepository{db: d}
}

func (d *DB) Stack() *StackRepository {
	return &StackRepository{db: d}
}

func (d *DB) Run() *RunRepository {
	return &RunRepository{db: d}
}

func (d *DB) AuditLog() *AuditLogRepository {
	return &AuditLogRepository{db: d}
}

func (d *DB) Backend() *BackendRepository {
	return &BackendRepository{db: d}
}

func (d *DB) Runner() *RunnerRepository {
	return &RunnerRepository{db: d}
}

func (d *DB) BeginTx() (*sql.Tx, error) {
	return d.conn.Begin()
}
