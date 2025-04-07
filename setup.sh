#!/bin/bash
set -e

# Start the Docker services
docker-compose up -d db localstack

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 5

# Create the SQS queue
echo "Creating SQS queue..."
aws --endpoint-url=http://localhost:4566 sqs create-queue \
    --queue-name file-sync-queue \
    --region us-east-1 \
    --attributes '{"VisibilityTimeout": "60"}'

# Create the S3 bucket
echo "Creating S3 bucket..."
aws --endpoint-url=http://localhost:4566 s3 mb s3://file-sync-bucket \
    --region us-east-1

echo "Setup complete. Environment is ready!"