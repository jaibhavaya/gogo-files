package onedrive

import (
	"io"
	"net/http"

	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
)

type HTTPInteractor interface {
	DoRequest(method, path string, body io.Reader, headers map[string]string) (*http.Response, error)
}

type DBInteractor interface {
	GetOneDriveIntegration(ownerID int64) (*db.OneDriveIntegration, error)
	GetOneDriveRefreshToken(ownerID int64) (string, error)
	SaveOneDriveRefreshToken(ownerID int64, userID, refreshToken string) error
}

type Service struct {
	dbPool     *db.Pool
	client     HTTPInteractor
	repository DBInteractor
}

func NewService(onedriveIntegration *db.OneDriveIntegration, dbPool *db.Pool, cfg config.Config) *Service {
	return &Service{
		dbPool:     dbPool,
		client:     newClient(onedriveIntegration, cfg.OnedriveClientID, cfg.OnedriveClientSecret),
		repository: db.NewPostgresRepository(dbPool),
	}
}

func NewServiceWithDependencies(
	dbPool *db.Pool,
	client HTTPInteractor,
	repository DBInteractor,
) *Service {
	return &Service{
		dbPool:     dbPool,
		client:     client,
		repository: repository,
	}
}

