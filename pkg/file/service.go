package file

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

const FOUR_MB int64 = 4 * 1024 * 1024

func (f *Service) getObject(bucket, key string) (*s3.GetObjectOutput, error) {
	result, err := f.s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't get object: %v", err)
	}

	return result, nil
}

func (f *Service) SyncFile(bucket, key string) error {
	fileReader, err := f.getObject(bucket, key)
	if err != nil {
		return fmt.Errorf("failed to get file reader: %v", err)
	}

	defer fileReader.Body.Close()

	size := *fileReader.ContentLength

	if size < FOUR_MB {
		fmt.Println("Under four mb! sync normally")
	} else {
		fmt.Println("Over 4mb! stream sync this bad boy!")
	}

	return nil
}
