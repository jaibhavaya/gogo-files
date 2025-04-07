package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/jaibhavaya/gogo-files/pkg/messages"
	"github.com/jaibhavaya/gogo-files/pkg/onedrive"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbPool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Run database migrations
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	var awsOptions []func(*awsconfig.LoadOptions) error
	awsOptions = append(awsOptions, awsconfig.WithRegion(cfg.AWSRegion))

	if cfg.Environment == "development" && cfg.S3Endpoint != "" {
		log.Printf("Using local endpoint for AWS services: %s", cfg.S3Endpoint)
		awsOptions = append(awsOptions,
			awsconfig.WithEndpointResolverWithOptions(
				aws.EndpointResolverWithOptionsFunc(
					func(service, region string, options ...interface{}) (aws.Endpoint, error) {
						return aws.Endpoint{
							URL: cfg.S3Endpoint,
						}, nil
					},
				),
			),
		)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsOptions...,
	)
	if err != nil {
		log.Fatalf("Failed to load AWS configuration: %v", err)
	}

	sqsClient := sqs.NewFromConfig(awsCfg)

	onedriveClient := onedrive.NewClient(
		dbPool,
		cfg.EncryptionKey,
		cfg.OneDriveClientID,
		cfg.OneDriveClientSecret,
	)

	fmt.Println("GoGo Files SQS Consumer starting...")
	fmt.Printf("Listening for messages on queue: %s\n", cfg.QueueURL)

	for {
		receiveOutput, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(cfg.QueueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20, // Long polling
		})
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if len(receiveOutput.Messages) == 0 {
			fmt.Println("No messages received. Waiting...")
			time.Sleep(time.Second)
			continue
		}

		for _, message := range receiveOutput.Messages {
			if message.MessageId != nil {
				fmt.Printf("Processing message ID: %s\n", *message.MessageId)
			}

			if message.Body != nil {
				err := processMessage(*message.Body, dbPool, cfg, onedriveClient)
				if err != nil {
					log.Printf("Error processing message: %v", err)
				} else {
					fmt.Println("Message processed successfully")
				}
			}

			// Delete the processed message
			if message.ReceiptHandle != nil {
				_, err := sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(cfg.QueueURL),
					ReceiptHandle: message.ReceiptHandle,
				})
				if err != nil {
					log.Printf("Error deleting message: %v", err)
				} else {
					fmt.Println("Message deleted from queue")
				}
			}
		}
	}
}

// TODO: should pull this into separate file
func processMessage(messageBody string, dbPool *db.Pool, cfg *config.Config, onedriveClient *onedrive.Client) error {
	message, err := messages.ParseMessage(messageBody)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	switch msg := message.(type) {
	case *messages.OneDriveAuthorizationMessage:
		fmt.Printf("Handling OneDrive authorization for owner: %d, user: %d\n",
			msg.Payload.OwnerID, msg.Payload.UserID)

		err := db.SaveOneDriveRefreshToken(
			dbPool,
			msg.Payload.OwnerID,
			msg.Payload.UserID,
			msg.Payload.RefreshToken,
			cfg.EncryptionKey,
		)
		if err != nil {
			return fmt.Errorf("failed to save OneDrive refresh token: %w", err)
		}

		fmt.Printf("OneDrive refresh token saved for owner: %d\n", msg.Payload.OwnerID)

		accessToken, err := onedriveClient.GetAccessToken(msg.Payload.OwnerID)
		if err != nil {
			fmt.Printf("Warning: Saved refresh token, but token validation failed: %v\n", err)
			fmt.Println("The refresh token may be invalid or expired")
		} else {
			fmt.Println("Successfully validated refresh token and obtained access token")
			fmt.Printf("Access token: %s...\n", accessToken[:min(20, len(accessToken))])
			fmt.Println("OneDrive integration is now ready for use")
		}

	case *messages.FileSyncMessage:
		fmt.Printf("Handling file sync request for owner: %d\n", msg.Payload.OwnerID)
		fmt.Printf("  - Source: s3://%s/%s\n", msg.Payload.Bucket, msg.Payload.Key)
		fmt.Printf("  - Destination: %s\n", msg.Payload.Destination)

		accessToken, err := onedriveClient.GetAccessToken(msg.Payload.OwnerID)
		if err != nil {
			return fmt.Errorf("failed to get OneDrive access token: %w", err)
		}

		fmt.Printf("Successfully obtained access token: %s...\n", accessToken[:min(20, len(accessToken))])
		// TODO: Implement file sync logic once token refresh is working
		// 1. Download the file from S3
		// 2. Upload the file to OneDrive
		// 3. Update the file status in the database
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
