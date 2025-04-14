package main

import (
	"log"

	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/jaibhavaya/gogo-files/pkg/event"
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

	processor := event.NewSQSProcessor(
		cfg,
		dbPool,
	)

	if err := processor.Start(); err != nil {
		panic(err)
	}

	select {}
}
