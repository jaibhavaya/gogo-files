package onedrive

import (
	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
)

type Service struct {
	dbPool *db.Pool
	client *client
}

func NewService(onedriveIntegration *db.OneDriveIntegration, dbPool *db.Pool, cfg config.Config) *Service {
	return &Service{
		dbPool: dbPool,
		client: newClient(onedriveIntegration, cfg.OnedriveClientID, cfg.OnedriveClientSecret),
	}
}
