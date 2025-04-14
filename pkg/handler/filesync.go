package handler

import (
	"fmt"
	"time"

	"github.com/jaibhavaya/gogo-files/pkg/file"
	"github.com/jaibhavaya/gogo-files/pkg/onedrive"
)

type fileSyncHandler struct {
	bucket          string
	destination     string
	ownerID         int64
	key             string
	onedriveService *onedrive.Service
	fileService     *file.Service
}

func (h *fileSyncHandler) Handle() error {
	fmt.Printf("Handling file sync request for owner: %d\n", h.ownerID)
	fmt.Printf("  - Source: s3://%s/%s\n", h.bucket, h.key)
	fmt.Printf("  - Destination: %s\n", h.destination)

	refreshToken, err := h.onedriveService.GetRefreshToken(h.ownerID)
	if err != nil {
		return fmt.Errorf("failed to handle file sync: %w", err)
	}

	_, err = h.onedriveService.GetAccessToken(refreshToken)
	if err != nil {
		return fmt.Errorf("failed to get OneDrive access token: %w", err)
	}

	fmt.Println("Successfully obtained access token")

	// Simulate doing the sync
	time.Sleep(2 * time.Second)
	fmt.Printf("Syncing File!\nbucket: %v\n  key: %v\n    destination: %v\n", h.bucket, h.key, h.destination)

	// TODO: Implement file sync logic once token refresh is working
	// 1. Download the file from S3
	// 2. Upload the file to OneDrive
	// 3. Update the file status in the database
	// probably use streaming though...

	return err
}

func NewFileSyncHandler(
	ownerID int64,
	key, bucket, destination string,
	onedriveService *onedrive.Service,
	fileService *file.Service,
) *fileSyncHandler {
	return &fileSyncHandler{
		bucket:          bucket,
		destination:     destination,
		ownerID:         ownerID,
		key:             key,
		onedriveService: onedriveService,
		fileService:     fileService,
	}
}
