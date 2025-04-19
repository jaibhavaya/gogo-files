package file

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/jaibhavaya/gogo-files/pkg/onedrive"
)

type S3ClientInterface interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type OneDriveServiceInterface interface {
	UploadSmallFile(driveID, folderID, fileName string, fileContent io.Reader, fileSize int64) error
	// Add other OneDrive methods as needed
}

type Service struct {
	dbPool          *db.Pool
	s3Client        S3ClientInterface
	onedriveService OneDriveServiceInterface
	dbRepository    db.Repository
}

func NewService(onedriveIntegration *db.OneDriveIntegration, dbPool *db.Pool, cfg config.Config) *Service {
	s3Client := newS3Client(cfg.AWSRegion, cfg.S3Endpoint, cfg.AWSAccessKey, cfg.AWSSecretKey, "")
	onedriveService := onedrive.NewService(onedriveIntegration, dbPool, cfg)
	dbRepository := db.NewPostgresRepository(dbPool)

	return &Service{
		dbPool:          dbPool,
		s3Client:        s3Client,
		onedriveService: onedriveService,
		dbRepository:    dbRepository,
	}
}

func NewServiceWithDependencies(
	dbPool *db.Pool,
	s3Client S3ClientInterface,
	onedriveService OneDriveServiceInterface,
	dbRepository db.Repository,
) *Service {
	return &Service{
		dbPool:          dbPool,
		s3Client:        s3Client,
		onedriveService: onedriveService,
		dbRepository:    dbRepository,
	}
}

