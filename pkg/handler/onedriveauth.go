package handler

import (
	"fmt"

	"github.com/jaibhavaya/gogo-files/pkg/db"
)

type oneDriveAuthHandler struct {
	refreshToken string
	ownerID      int64
	userID       string
	dbPool       *db.Pool
}

func (h *oneDriveAuthHandler) Handle() error {
	fmt.Printf("Handling OneDrive authorization for owner: %d, user: %s\n", h.ownerID, h.userID)

	err := db.SaveOneDriveRefreshToken(h.dbPool, h.ownerID, h.userID, h.refreshToken)
	if err != nil {
		return fmt.Errorf("failed to save OneDrive refresh token: %w", err)
	}

	fmt.Printf("OneDrive refresh token saved for owner: %d\n", h.ownerID)

	return nil
}

func NewOnedriveAuthHandler(
	ownerID int64,
	userID,
	refreshToken string,
	dbPool *db.Pool,
) *oneDriveAuthHandler {
	return &oneDriveAuthHandler{
		refreshToken: refreshToken,
		ownerID:      ownerID,
		userID:       userID,
		dbPool:       dbPool,
	}
}
