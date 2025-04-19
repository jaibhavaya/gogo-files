package file

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Creates a new S3 client that implements S3ClientInterface
func newS3Client(region, endpoint, key, secret, session string) S3ClientInterface {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		// TODO investigate what to use for production config
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			key, secret, session),
		),
	)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // special for localstack, check what's needed for production
	})
}