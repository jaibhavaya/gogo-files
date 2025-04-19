package file

import (
	"fmt"
	"sync"

	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
)

type Item interface {
	Bucket() string
	Key() string
	ID() string
	Name() string
	Path() string
	Size() int
}

type SyncHandler struct {
	OwnerID int64
	UserID  string
	Items   []Item

	DbPool *db.Pool
	Config config.Config
}

type FileResult struct {
	msg string
	// definitely will have more, just a placeholder
}

func processItem(
	item Item,
	service Service,
	results chan<- FileResult,
) {
	bucket := item.Bucket()
	key := item.Key()
	path := item.Path()

	fmt.Printf("Syncing File\nbucket: %v\n  key: %v\n    path: %v\n", bucket, key, path)

	err := service.SyncFile(SyncFileParams{Bucket: bucket, Key: key})
	if err != nil {
		results <- FileResult{fmt.Sprintf("failed to sync file: %v in bucket: %v because: %v", key, bucket, err)}
	}

	results <- FileResult{fmt.Sprintf("successfully synced file: %v in bucket: %v", key, bucket)}
}

func (h SyncHandler) Handle() error {
	fmt.Printf("Handling file sync request for owner: %d\n", h.OwnerID)

	onedriveIntegration, err := db.GetOneDriveIntegration(h.DbPool, h.OwnerID)
	if err != nil {
		return fmt.Errorf("failed to get onedrive integration: %v", err)
	}

	fileService := NewService(onedriveIntegration, h.DbPool, h.Config)

	results := make(chan FileResult, len(h.Items))
	wg := sync.WaitGroup{}
	for _, item := range h.Items {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processItem(item, *fileService, results)
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		fmt.Println("Got result: ", result)
	}

	return err
}
