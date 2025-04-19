package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*Pool, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}

	pool := &Pool{DB: db}
	return pool, mock
}

func TestGetOneDriveIntegration_Success(t *testing.T) {
	pool, mock := setupMockDB(t)
	defer pool.Close()

	repo := NewPostgresRepository(pool)

	ownerID := int64(123)
	expectedIntegration := &OneDriveIntegration{
		OwnerID:      ownerID,
		UserID:       "test-user",
		RefreshToken: "test-token",
	}

	rows := sqlmock.NewRows([]string{"owner_id", "user_id", "refresh_token"}).
		AddRow(expectedIntegration.OwnerID, expectedIntegration.UserID, expectedIntegration.RefreshToken)

	mock.ExpectQuery("SELECT owner_id, user_id, refresh_token FROM onedrive_integrations WHERE owner_id = \\$1").
		WithArgs(ownerID).
		WillReturnRows(rows)

	integration, err := repo.GetOneDriveIntegration(ownerID)

	assert.NoError(t, err)
	assert.NotNil(t, integration)
	assert.Equal(t, expectedIntegration.OwnerID, integration.OwnerID)
	assert.Equal(t, expectedIntegration.UserID, integration.UserID)
	assert.Equal(t, expectedIntegration.RefreshToken, integration.RefreshToken)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOneDriveIntegration_NotFound(t *testing.T) {
	pool, mock := setupMockDB(t)
	defer pool.Close()

	repo := NewPostgresRepository(pool)

	ownerID := int64(123)

	mock.ExpectQuery("SELECT owner_id, user_id, refresh_token FROM onedrive_integrations WHERE owner_id = \\$1").
		WithArgs(ownerID).
		WillReturnError(sql.ErrNoRows)

	integration, err := repo.GetOneDriveIntegration(ownerID)

	assert.NoError(t, err) // No error, just nil result
	assert.Nil(t, integration)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOneDriveIntegration_DatabaseError(t *testing.T) {
	pool, mock := setupMockDB(t)
	defer pool.Close()

	repo := NewPostgresRepository(pool)

	ownerID := int64(123)

	expectedErr := errors.New("database connection error")
	mock.ExpectQuery("SELECT owner_id, user_id, refresh_token FROM onedrive_integrations WHERE owner_id = \\$1").
		WithArgs(ownerID).
		WillReturnError(expectedErr)

	integration, err := repo.GetOneDriveIntegration(ownerID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get OneDrive integration")
	assert.Nil(t, integration)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveOneDriveRefreshToken_Success(t *testing.T) {
	pool, mock := setupMockDB(t)
	defer pool.Close()

	repo := NewPostgresRepository(pool)

	ownerID := int64(123)
	userID := "test-user"
	refreshToken := "new-refresh-token"

	mock.ExpectExec("INSERT INTO onedrive_integrations").
		WithArgs(ownerID, userID, refreshToken).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.SaveOneDriveRefreshToken(ownerID, userID, refreshToken)

	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveOneDriveRefreshToken_Error(t *testing.T) {
	pool, mock := setupMockDB(t)
	defer pool.Close()

	repo := NewPostgresRepository(pool)

	ownerID := int64(123)
	userID := "test-user"
	refreshToken := "new-refresh-token"

	expectedErr := errors.New("database constraint violation")
	mock.ExpectExec("INSERT INTO onedrive_integrations").
		WithArgs(ownerID, userID, refreshToken).
		WillReturnError(expectedErr)

	err := repo.SaveOneDriveRefreshToken(ownerID, userID, refreshToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save refresh token")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOneDriveRefreshToken_Success(t *testing.T) {
	pool, mock := setupMockDB(t)
	defer pool.Close()

	repo := NewPostgresRepository(pool)

	ownerID := int64(123)
	expectedToken := "test-refresh-token"

	rows := sqlmock.NewRows([]string{"refresh_token"}).AddRow(expectedToken)
	mock.ExpectQuery("SELECT refresh_token FROM onedrive_integrations WHERE owner_id = \\$1").
		WithArgs(ownerID).
		WillReturnRows(rows)

	token, err := repo.GetOneDriveRefreshToken(ownerID)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOneDriveRefreshToken_NotFound(t *testing.T) {
	pool, mock := setupMockDB(t)
	defer pool.Close()

	repo := NewPostgresRepository(pool)

	ownerID := int64(123)

	mock.ExpectQuery("SELECT refresh_token FROM onedrive_integrations WHERE owner_id = \\$1").
		WithArgs(ownerID).
		WillReturnError(sql.ErrNoRows)

	token, err := repo.GetOneDriveRefreshToken(ownerID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active OneDrive integration found")
	assert.Equal(t, "", token)

	assert.NoError(t, mock.ExpectationsWereMet())
}

