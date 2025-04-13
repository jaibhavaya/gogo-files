package main

import (
	"log"

	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/jaibhavaya/gogo-files/pkg/event"
	"github.com/jaibhavaya/gogo-files/pkg/service"
	"github.com/joho/godotenv"
)

func main() {
	// TODO: pull other configuration options from .env as well
	// numWorkers
	// numSubscribers
	// queue name
	// others...
	// should also restructure config struct to organize these better

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

	onedriveService := service.NewOnedriveService(dbPool, *cfg)
	fileService := service.NewFileService()

	numSubscribers := 1
	numWorkers := 5

	processor := event.NewSQSProcessor(
		"gogo-files-queue",
		numSubscribers,
		numWorkers,
		onedriveService,
		fileService,
	)

	if err := processor.Start(); err != nil {
		panic(err)
	}

	select {}
}

// TODO: should pull this into separate file
