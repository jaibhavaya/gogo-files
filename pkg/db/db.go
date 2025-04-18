package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// TODO: probably should create some sort of db service and keep this module low level for db setup/querying
// or maybe just separate into different files for subject areas (onedrive / files / etc)

type OneDriveIntegration struct {
	OwnerID      int64  `db:"owner_id"`
	UserID       string `db:"user_id"`
	RefreshToken string `db:"refresh_token"`
}

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

func GetOneDriveIntegration(pool *Pool, ownerID int64) (*OneDriveIntegration, error) {
	query := `
        SELECT owner_id, user_id, refresh_token
        FROM onedrive_integrations
        WHERE owner_id = $1
    `

	var integration OneDriveIntegration
	err := pool.DB.QueryRow(query, ownerID).Scan(
		&integration.OwnerID,
		&integration.UserID,
		&integration.RefreshToken,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get OneDrive integration: %w", err)
	}

	return &integration, nil
}

func SaveOneDriveRefreshToken(pool *Pool, ownerID int64, userID string, refreshToken string) error {
	query := `
		INSERT INTO onedrive_integrations
		(owner_id, user_id, refresh_token)
		VALUES ($1, $2, $3)
		ON CONFLICT (owner_id)
		DO UPDATE SET
			user_id = EXCLUDED.user_id,
			refresh_token = EXCLUDED.refresh_token
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
		WHERE owner_id = $1
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
