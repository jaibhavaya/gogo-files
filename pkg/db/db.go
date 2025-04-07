package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Pool struct {
	DB *sql.DB
}

func Connect(connectionString string) (*Pool, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{DB: db}, nil
}

func (p *Pool) Close() error {
	return p.DB.Close()
}

// Really need to refactor this to use built in migrations
func RunMigrations(connectionString string) error {
	// Let's create a simpler implementation using direct SQL execution
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	// Check if the files table exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'files')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if files table exists: %w", err)
	}

	if !exists {
		log.Println("Creating files table...")
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS files (
				id SERIAL PRIMARY KEY,
				name TEXT NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);
		`)
		if err != nil {
			return fmt.Errorf("failed to create files table: %w", err)
		}
	}

	// Check if the onedrive_integrations table exists
	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'onedrive_integrations')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if onedrive_integrations table exists: %w", err)
	}

	if !exists {
		log.Println("Creating onedrive_integrations table...")
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS onedrive_integrations (
				id SERIAL PRIMARY KEY,
				owner_id BIGINT NOT NULL,
				user_id TEXT NOT NULL,
				encrypted_refresh_token TEXT NOT NULL,
				is_active BOOLEAN DEFAULT TRUE,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				UNIQUE(owner_id)
			);
		`)
		if err != nil {
			return fmt.Errorf("failed to create onedrive_integrations table: %w", err)
		}
	}

	log.Println("Database schema is up to date")
	return nil
}

func SaveOneDriveRefreshToken(pool *Pool, ownerID int64, userID string, refreshToken, encryptionKey string) error {
	// In a real implementation, encrypt the refresh token before storing
	encryptedToken, err := encryptToken(refreshToken, encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %w", err)
	}

	query := `
		INSERT INTO onedrive_integrations
		(owner_id, user_id, encrypted_refresh_token, is_active)
		VALUES ($1, $2, $3, TRUE)
		ON CONFLICT (owner_id)
		DO UPDATE SET
			user_id = EXCLUDED.user_id,
			encrypted_refresh_token = EXCLUDED.encrypted_refresh_token,
			is_active = TRUE,
			updated_at = NOW()
	`

	_, err = pool.DB.Exec(query, ownerID, userID, encryptedToken)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	return nil
}

// Placeholder function for token encryption
func encryptToken(token, key string) (string, error) {
	// TODO: Implement real encryption
	// This is a placeholder that would be replaced with actual AES-GCM encryption
	return token + "_encrypted", nil
}

func GetOneDriveRefreshToken(pool *Pool, ownerID int64, encryptionKey string) (string, error) {
	query := `
		SELECT encrypted_refresh_token
		FROM onedrive_integrations
		WHERE owner_id = $1 AND is_active = TRUE
	`

	var encryptedToken string
	err := pool.DB.QueryRow(query, ownerID).Scan(&encryptedToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no active OneDrive integration found for owner %d", ownerID)
		}
		return "", fmt.Errorf("failed to query refresh token: %w", err)
	}

	refreshToken, err := decryptToken(encryptedToken, encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt token: %w", err)
	}

	return refreshToken, nil
}

func decryptToken(encryptedToken, key string) (string, error) {
	// TODO: Implement real decryption
	// This is a placeholder that would be replaced with actual AES-GCM decryption
	if len(encryptedToken) > 10 {
		return encryptedToken[:len(encryptedToken)-10], nil
	}
	return "", fmt.Errorf("invalid encrypted token")
}
