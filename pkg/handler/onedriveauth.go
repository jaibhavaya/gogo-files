package handler

import (
	"fmt"

	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/jaibhavaya/gogo-files/pkg/onedrive"
	"github.com/jaibhavaya/gogo-files/pkg/queue"
)

type OneDriveAuthHandler struct {
	message        queue.OneDriveAuthorizationMessage
	dbPool         db.Pool
	onedriveClient onedrive.Client
}

func (h *OneDriveAuthHandler) Handle() error {
	ownerID := h.message.Payload.OwnerID
	userID := h.message.Payload.UserID
	fmt.Printf("Handling OneDrive authorization for owner: %d, user: %s\n", ownerID, userID)

	err := db.SaveOneDriveRefreshToken(
		&h.dbPool,
		ownerID,
		userID,
		h.message.Payload.RefreshToken,
	)
	if err != nil {
		return fmt.Errorf("failed to save OneDrive refresh token: %w", err)
	}

	fmt.Printf("OneDrive refresh token saved for owner: %d\n", ownerID)

	// we just try to fetch an access token to validate the refresh token
	_, err = h.onedriveClient.GetAccessToken(ownerID)
	if err != nil {
		return fmt.Errorf("Failed to validate Onedrive Refresh Token %w", err)
	}

	fmt.Println("Successfully validated refresh token")
	fmt.Println("OneDrive integration is now ready for use")

	return nil
}
