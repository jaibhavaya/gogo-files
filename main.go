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

	numSubscribers := 1
	numWorkers := 5

	processor := queue.NewSQSProcessor("gogo-files-queue", numSubscribers, numWorkers)

	if err := processor.Start(); err != nil {
		panic(err)
	}

	processor.StartPublishing()

	select {}
}

// TODO: should pull this into separate file
