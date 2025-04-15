package main

import (
	"log"

	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/jaibhavaya/gogo-files/pkg/processor"
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

	processor := processor.NewSQSProcessor(
		cfg,
		dbPool,
	)

	if err := processor.Start(); err != nil {
		log.Fatalf("Failed to start SQS processor: %v", err)
	}

	select {}
}
