package onedrive

import (
	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
)

type Service struct {
	dbPool         *db.Pool
	onedriveClient *client
}

func NewService(dbPool *db.Pool, cfg *config.Config) *Service {
	return &Service{
		dbPool:         dbPool,
		onedriveClient: newClient(cfg.OnedriveClientID, cfg.OnedriveClientSecret),
	}
}
