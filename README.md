# GoGo Files

A Go-based microservice to sync files from AWS S3 to Microsoft OneDrive.

## Overview

GoGo Files is a port of the Rust-based Ferris File Sync service to Go. It provides an API and background worker for synchronizing files between S3 and OneDrive.

## Features

- SQS message consumer
- OneDrive OAuth integration
- File synchronization from S3 to OneDrive
- Database persistence with PostgreSQL
- Token encryption for secure storage

## Prerequisites

- Go 1.24+
- PostgreSQL database
- AWS account with S3 and SQS services
- Microsoft Azure AD application for OneDrive integration

## Configuration

Configuration is handled through environment variables:

```
DATABASE_URL=postgres://username:password@localhost:5432/gogo_files
QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789012/gogo-files-queue
AWS_REGION=us-east-1
S3_BUCKET=your-s3-bucket
S3_ENDPOINT=http://localhost:4566 # Optional, for local development with LocalStack
ENCRYPTION_KEY=your-encryption-key
ONEDRIVE_CLIENT_ID=your-client-id
ONEDRIVE_CLIENT_SECRET=your-client-secret
```

## Setup

1. Clone the repository
2. Install dependencies:
   ```
   go mod download
   ```
3. Set up the database:
   ```
   psql -U postgres -c "CREATE DATABASE gogo_files"
   ```
4. Run the service:
   ```
   go run main.go
   ```

## Development

The migrations directory contains SQL files for database setup.

## Building

```
go build -o gogo-files
```

## Deployment

The service can be containerized using Docker:

```
docker build -t gogo-files .
docker run -p 8080:8080 gogo-files
```

## Message Format

The service processes two types of SQS messages:

1. OneDrive Authorization:
```json
{
  "message_type": "onedrive_authorization",
  "payload": {
    "owner_id": 123,
    "user_id": 456,
    "refresh_token": "your-refresh-token"
  }
}
```

2. File Sync:
```json
{
  "message_type": "file_sync",
  "payload": {
    "owner_id": 123,
    "bucket": "your-s3-bucket",
    "key": "path/to/file.pdf",
    "destination": "OneDrive/Documents/file.pdf"
  }
}
```
