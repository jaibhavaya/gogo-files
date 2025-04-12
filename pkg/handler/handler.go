package handler

import (
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/jaibhavaya/gogo-files/pkg/onedrive"
	"github.com/jaibhavaya/gogo-files/pkg/queue"
)

type Handler interface {
	Handle() error
}

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

type FileSyncHandler struct {
	message        queue.FileSyncMessage
	dbPool         db.Pool
	onedriveClient onedrive.Client
}

func (h *FileSyncHandler) Handle() error {
	messagePayload := h.message.Payload
	ownerID := messagePayload.OwnerID

	fmt.Printf("Handling file sync request for owner: %d\n", ownerID)
	fmt.Printf("  - Source: s3://%s/%s\n", messagePayload.Bucket, messagePayload.Key)
	fmt.Printf("  - Destination: %s\n", messagePayload.Destination)

	_, err := h.onedriveClient.GetAccessToken(messagePayload.OwnerID)
	if err != nil {
		return fmt.Errorf("failed to get OneDrive access token: %w", err)
	}

	fmt.Println("Successfully obtained access token")
	// TODO: Implement file sync logic once token refresh is working
	// 1. Download the file from S3
	// 2. Upload the file to OneDrive
	// 3. Update the file status in the database
	// probably use streaming though...

	return err
}

func NewHandler(msg queue.Message, dbPool db.Pool, cfg *config.Config, onedriveClient onedrive.Client) (Handler, error) {
	// TODO: possibly add validation here since dbPool and onedriveClient are required for the handlers created below
	switch msg := msg.(type) {
	case *queue.OneDriveAuthorizationMessage:
		return &OneDriveAuthHandler{
			message:        *msg,
			dbPool:         dbPool,
			onedriveClient: onedriveClient,
		}, nil

	case *queue.FileSyncMessage:
		return &FileSyncHandler{
			message:        *msg,
			dbPool:         dbPool,
			onedriveClient: onedriveClient,
		}, nil
	}

	return nil
}
