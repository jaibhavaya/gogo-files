package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL          string
	QueueURL             string
	AWSRegion            string
	S3Bucket             string
	S3Endpoint           string
	Environment          string // "development" or "production"
	EncryptionKey        string
	OneDriveClientID     string
	OneDriveClientSecret string
}

func FromEnv() (*Config, error) {
	databaseURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	queueURL, ok := os.LookupEnv("QUEUE_URL")
	if !ok {
		return nil, fmt.Errorf("QUEUE_URL is not set")
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-west-1"
	}

	s3Bucket, ok := os.LookupEnv("S3_BUCKET")
	if !ok {
		return nil, fmt.Errorf("S3_BUCKET is not set")
	}

	s3Endpoint := os.Getenv("S3_ENDPOINT")

	// Check environment
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		encryptionKey = "default-dev-key-please-change-in-production"
	}

	onedriveClientID := os.Getenv("ONEDRIVE_CLIENT_ID")
	if onedriveClientID == "" {
		onedriveClientID = "your-client-id"
	}

	onedriveClientSecret := os.Getenv("ONEDRIVE_CLIENT_SECRET")
	if onedriveClientSecret == "" {
		onedriveClientSecret = "your-client-secret"
	}

	return &Config{
		DatabaseURL:          databaseURL,
		QueueURL:             queueURL,
		AWSRegion:            awsRegion,
		S3Bucket:             s3Bucket,
		S3Endpoint:           s3Endpoint,
		Environment:          environment,
		EncryptionKey:        encryptionKey,
		OneDriveClientID:     onedriveClientID,
		OneDriveClientSecret: onedriveClientSecret,
	}, nil
}
