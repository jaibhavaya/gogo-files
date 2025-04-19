package file

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

const FOUR_MB int64 = 4 * 1024 * 1024

type SyncFileParams struct {
	Bucket     string
	Key        string
	DriveID    string
	FolderPath string
	FileName   string
}

func (s *Service) SyncFile(params SyncFileParams) error {
	file, err := s.s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(params.Bucket),
		Key:    aws.String(params.Key),
	})
	if err != nil {
		return fmt.Errorf("couldn't get object: %v", err)
	}

	defer file.Body.Close()

	size := *file.ContentLength

	if size < FOUR_MB {
		fmt.Println("Under four mb! sync normally")
		err = s.onedriveService.UploadSmallFile(
			params.DriveID,
			params.FolderPath,
			params.FileName,
			file.Body,
			*file.ContentLength,
		)
		if err != nil {
			return fmt.Errorf("failed to upload small file: %w", err)
		}
	} else {
		fmt.Println("Over 4mb! stream sync this bad boy!")
		// TODO: Implement large file upload
	}

	return nil
}