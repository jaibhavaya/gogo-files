package file

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/jaibhavaya/gogo-files/pkg/onedrive"
)

type Service struct {
	dbPool          *db.Pool
	s3Client        *s3.Client
	onedriveService *onedrive.Service
}

func NewService(dbPool *db.Pool, onedriveService *onedrive.Service, cfg *config.Config) *Service {
	return &Service{
		dbPool:          dbPool,
		s3Client:        newS3Client(cfg.AWSRegion, cfg.S3Endpoint, "test", "test", ""),
		onedriveService: onedriveService,
	}
}
