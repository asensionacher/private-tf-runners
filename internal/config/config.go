package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	FrontendURL     string
}

type DatabaseConfig struct {
	Path string
}

type SecurityConfig struct {
	JWTSecret          []byte
	JWTExpiration      time.Duration
	RefreshExpiration  time.Duration
	CSRFSecret         []byte
	RateLimitWindow    time.Duration
	RateLimitMax       int
	bcryptCost         int
	MaxLoginAttempts   int
	LockoutDuration    time.Duration
	SecureCookie       bool
	SessionCookieName  string
	EncryptionKey      []byte
}

var globalConfig *Config

func Load() (*Config, error) {
	jwtSecret, err := generateSecureSecret(32)
	if err != nil {
		return nil, err
	}

	csrfSecret, err := generateSecureSecret(32)
	if err != nil {
		return nil, err
	}

	encryptionKey, err := generateSecureSecret(32)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:5173"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DATABASE_PATH", "./data/runners.db"),
		},
		Security: SecurityConfig{
			JWTSecret:         jwtSecret,
			JWTExpiration:     15 * time.Minute,
			RefreshExpiration: 7 * 24 * time.Hour,
			CSRFSecret:        csrfSecret,
			RateLimitWindow:   1 * time.Minute,
			RateLimitMax:      10,
			bcryptCost:        12,
			MaxLoginAttempts:  5,
			LockoutDuration:   15 * time.Minute,
			SecureCookie:      false,
			SessionCookieName: "__Host-session",
			EncryptionKey:     encryptionKey,
		},
	}

	globalConfig = cfg
	return cfg, nil
}

func Get() *Config {
	return globalConfig
}

func (c *Config) Encrypt(plaintext string) (string, error) {
	if len(c.Security.EncryptionKey) == 0 {
		return "", nil
	}

	block, err := aes.NewCipher(c.Security.EncryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (c *Config) Decrypt(encoded string) (string, error) {
	if len(c.Security.EncryptionKey) == 0 {
		return encoded, nil
	}

	ciphertext, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(c.Security.EncryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func generateSecureSecret(length int) ([]byte, error) {
	secret := make([]byte, length)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	return secret, nil
}

func (c *SecurityConfig) GetBCryptCost() int {
	return c.bcryptCost
}

func (c *SecurityConfig) GenerateCSRFToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(token), nil
}

func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func SecureRandomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

func SecureRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:n], nil
}
