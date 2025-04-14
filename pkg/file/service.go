package file

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

func (f *Service) GetObject(bucket, key string) (*s3.GetObjectOutput, error) {
	result, err := f.s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't get object: %v", err)
	}

	return result, nil
}
