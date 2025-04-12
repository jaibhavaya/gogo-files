package main

import (
	"log"

	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
	_ "github.com/jaibhavaya/gogo-files/pkg/messages"
	"github.com/jaibhavaya/gogo-files/pkg/onedrive"
	"github.com/jaibhavaya/gogo-files/pkg/queue"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbPool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	_ = onedrive.NewClient(
		dbPool,
		cfg.EncryptionKey,
		cfg.OneDriveClientID,
		cfg.OneDriveClientSecret,
	)

	numSubscribers := 10
	numWorkers := 50

	processor := queue.NewSQSProcessor("gogo-files-queue", numSubscribers, numWorkers)

	if err := processor.Start(); err != nil {
		panic(err)
	}

	processor.StartPublishing()

	select {}
}

// TODO: should pull this into separate file
// func processMessage(messageBody string, dbPool *db.Pool, cfg *config.Config, onedriveClient *onedrive.Client) error {
// 	message, err := messages.ParseMessage(messageBody)
// 	if err != nil {
// 		return fmt.Errorf("failed to parse message: %w", err)
// 	}

// 	switch msg := message.(type) {
// 	case *messages.OneDriveAuthorizationMessage:
// 		fmt.Printf("Handling OneDrive authorization for owner: %d, user: %s\n",
// 			msg.Payload.OwnerID, msg.Payload.UserID)

// 		err := db.SaveOneDriveRefreshToken(
// 			dbPool,
// 			msg.Payload.OwnerID,
// 			msg.Payload.UserID,
// 			msg.Payload.RefreshToken,
// 		)
// 		if err != nil {
// 			return fmt.Errorf("failed to save OneDrive refresh token: %w", err)
// 		}

// 		fmt.Printf("OneDrive refresh token saved for owner: %d\n", msg.Payload.OwnerID)

// 		accessToken, err := onedriveClient.GetAccessToken(msg.Payload.OwnerID)
// 		if err != nil {
// 			fmt.Printf("Warning: Saved refresh token, but token validation failed: %v\n", err)
// 			fmt.Println("The refresh token may be invalid or expired")
// 		} else {
// 			fmt.Println("Successfully validated refresh token and obtained access token")
// 			fmt.Printf("Access token: %s...\n", accessToken[:min(20, len(accessToken))])
// 			fmt.Println("OneDrive integration is now ready for use")
// 		}

// 	case *messages.FileSyncMessage:
// 		fmt.Printf("Handling file sync request for owner: %d\n", msg.Payload.OwnerID)
// 		fmt.Printf("  - Source: s3://%s/%s\n", msg.Payload.Bucket, msg.Payload.Key)
// 		fmt.Printf("  - Destination: %s\n", msg.Payload.Destination)

// 		accessToken, err := onedriveClient.GetAccessToken(msg.Payload.OwnerID)
// 		if err != nil {
// 			return fmt.Errorf("failed to get OneDrive access token: %w", err)
// 		}

// 		fmt.Printf("Successfully obtained access token: %s...\n", accessToken[:min(20, len(accessToken))])
// 		// TODO: Implement file sync logic once token refresh is working
// 		// 1. Download the file from S3
// 		// 2. Upload the file to OneDrive
// 		// 3. Update the file status in the database
// 	}

// 	return nil
// }
