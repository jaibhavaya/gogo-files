package main

import (
	"log"

	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/jaibhavaya/gogo-files/pkg/event"
	"github.com/jaibhavaya/gogo-files/pkg/service"
)

func main() {
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

	// TODO: ability to pull these from somewhere external
	numSubscribers := 1
	numWorkers := 5

	processor := event.NewSQSProcessor(
		cfg.QueueURL,
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
