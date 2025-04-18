package onedrive

import (
	"fmt"

	"github.com/jaibhavaya/gogo-files/pkg/db"
)

type OneDriveAuthHandler struct {
	RefreshToken string
	OwnerID      int64
	UserID       string
	DbPool       *db.Pool
}

func (h *OneDriveAuthHandler) Handle() error {
	fmt.Printf("Handling OneDrive authorization for owner: %d, user: %s\n", h.OwnerID, h.UserID)

	err := db.SaveOneDriveRefreshToken(h.DbPool, h.OwnerID, h.UserID, h.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to save OneDrive refresh token: %w", err)
	}

	fmt.Printf("OneDrive refresh token saved for owner: %d\n", h.OwnerID)

	return nil
}
