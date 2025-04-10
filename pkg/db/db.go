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

func SaveOneDriveRefreshToken(pool *Pool, ownerID int64, userID string, refreshToken string) error {
	query := `
		INSERT INTO onedrive_integrations
		(owner_id, user_id, refreshToken, is_active)
		VALUES ($1, $2, $3, TRUE)
		ON CONFLICT (owner_id)
		DO UPDATE SET
			user_id = EXCLUDED.user_id,
			refresh_token = EXCLUDED.refresh_token,
			is_active = TRUE,
			updated_at = NOW()
	`

	_, err := pool.DB.Exec(query, ownerID, userID, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	return nil
}

func GetOneDriveRefreshToken(pool *Pool, ownerID int64) (string, error) {
	query := `
		SELECT refresh_token
		FROM onedrive_integrations
		WHERE owner_id = $1 AND is_active = TRUE
	`

	var refreshToken string
	err := pool.DB.QueryRow(query, ownerID).Scan(&refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no active OneDrive integration found for owner %d", ownerID)
		}
		return "", fmt.Errorf("failed to query refresh token: %w", err)
	}

	return refreshToken, nil
}
