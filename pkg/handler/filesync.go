package handler

import (
	"fmt"

	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/jaibhavaya/gogo-files/pkg/onedrive"
	"github.com/jaibhavaya/gogo-files/pkg/queue"
)

type fileSyncHandler struct {
	message        queue.FileSyncMessage
	dbPool         db.Pool
	onedriveClient onedrive.Client
}

func (h *fileSyncHandler) Handle() error {
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
