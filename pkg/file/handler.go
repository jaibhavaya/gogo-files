package file

import (
	"fmt"
	"time"

	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
)

type SyncHandler struct {
	Bucket      string
	Destination string
	OwnerID     int64
	Key         string
	DbPool      *db.Pool
	Config      config.Config
}

func (h SyncHandler) Handle() error {
	fmt.Printf("Handling file sync request for owner: %d\n", h.OwnerID)
	fmt.Printf("  - Source: s3://%s/%s\n", h.Bucket, h.Key)
	fmt.Printf("  - Destination: %s\n", h.Destination)

	onedriveIntegration, err := db.GetOneDriveIntegration(h.DbPool, h.OwnerID)
	if err != nil {
		return fmt.Errorf("failed to get onedrive integration: %v", err)
	}

	fileService := NewService(onedriveIntegration, h.DbPool, h.Config)

	fileService.SyncFile(h.Bucket, h.Key)

	// Simulate doing the sync
	time.Sleep(2 * time.Second)
	fmt.Printf("Syncing File!\nbucket: %v\n  key: %v\n    destination: %v\n", h.Bucket, h.Key, h.Destination)

	// TODO: Implement file sync logic once token refresh is working
	// 1. Download the file from S3
	// 2. Upload the file to OneDrive
	// 3. Update the file status in the database
	// probably use streaming though...

	return err
}
