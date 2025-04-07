#!/bin/bash
set -e

docker-compose up -d db localstack

echo "Waiting for services to start..."
sleep 5

echo "Creating SQS queue..."
aws --endpoint-url=http://localhost:4566 sqs create-queue \
    --queue-name gogo-files-queue \
    --region us-west-1 \
    --attributes '{"VisibilityTimeout": "60"}'

echo "Creating S3 bucket..."
aws --endpoint-url=http://localhost:4566 s3 mb s3://gogo-files \
    --region us-east-1

echo "Setup complete. Environment is ready!"
