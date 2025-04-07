package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
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

func RunMigrations(connectionString string) error {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	fsc, err := file.New("migrations", file.WithScheme("file"))
	if err != nil {
		return fmt.Errorf("failed to create file source: %w", err)
	}

	m, err := migrate.NewWithInstance("file", fsc, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func SaveOneDriveRefreshToken(pool *Pool, ownerID, userID int64, refreshToken, encryptionKey string) error {
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
