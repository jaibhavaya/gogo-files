package file

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

const FOUR_MB int64 = 4 * 1024 * 1024

func (f *Service) SyncFile(bucket, key string) error {
	file, err := f.s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("couldn't get object: %v", err)
	}

	defer file.Body.Close()

	size := *file.ContentLength

	if size < FOUR_MB {
		fmt.Println("Under four mb! sync normally")
		// TODO: get these from the sqs message
		driveID := "123"
		folderID := "456"
		fileName := "something.txt"
		f.onedriveService.UploadSmallFile(
			driveID, folderID, fileName,
			file.Body, *file.ContentLength,
		)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Over 4mb! stream sync this bad boy!")
	}

	return nil
}
