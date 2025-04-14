package handler

import (
	"fmt"

	"github.com/jaibhavaya/gogo-files/pkg/onedrive"
)

type oneDriveAuthHandler struct {
	refreshToken    string
	ownerID         int64
	userID          string
	onedriveService onedrive.Service
}

func (h *oneDriveAuthHandler) Handle() error {
	fmt.Printf("Handling OneDrive authorization for owner: %d, user: %s\n", h.ownerID, h.userID)

	err := h.onedriveService.SaveRefreshToken(h.ownerID, h.userID, h.refreshToken)
	if err != nil {
		return fmt.Errorf("failed to save OneDrive refresh token: %w", err)
	}

	fmt.Printf("OneDrive refresh token saved for owner: %d\n", h.ownerID)

	err = h.onedriveService.ValidateRefreshToken(h.refreshToken)
	if err != nil {
		return fmt.Errorf("failed to validate Onedrive Refresh Token %w", err)
	}

	fmt.Println("Successfully validated refresh token")
	fmt.Println("OneDrive integration is now ready for use")

	return nil
}

func NewOnedriveAuthHandler(
	ownerID int64,
	userID,
	refreshToken string,
	onedriveService *onedrive.Service,
) *oneDriveAuthHandler {
	return &oneDriveAuthHandler{
		refreshToken:    refreshToken,
		ownerID:         ownerID,
		userID:          userID,
		onedriveService: *onedriveService,
	}
}
