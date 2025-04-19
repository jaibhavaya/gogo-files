package file

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

type MockOneDriveService struct {
	mock.Mock
}

func (m *MockOneDriveService) UploadSmallFile(driveID, folderID, fileName string, fileContent io.Reader, fileSize int64) error {
	args := m.Called(driveID, folderID, fileName, fileSize)
	return args.Error(0)
}

type MockDBRepository struct {
	mock.Mock
}

func (m *MockDBRepository) GetOneDriveIntegration(ownerID int64) (*db.OneDriveIntegration, error) {
	args := m.Called(ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.OneDriveIntegration), args.Error(1)
}

func (m *MockDBRepository) SaveOneDriveRefreshToken(ownerID int64, userID, refreshToken string) error {
	args := m.Called(ownerID, userID, refreshToken)
	return args.Error(0)
}

func (m *MockDBRepository) GetOneDriveRefreshToken(ownerID int64) (string, error) {
	args := m.Called(ownerID)
	return args.String(0), args.Error(1)
}

func TestSyncFile_SmallFile_Success(t *testing.T) {
	mockS3Client := new(MockS3Client)
	mockOneDriveService := new(MockOneDriveService)
	mockDBRepo := new(MockDBRepository)

	testContent := []byte("test file content")
	testContentReader := io.NopCloser(bytes.NewReader(testContent))
	contentLength := int64(len(testContent))

	mockS3Client.On(
		"GetObject",
		mock.Anything,
		mock.MatchedBy(func(input *s3.GetObjectInput) bool {
			return *input.Bucket == "test-bucket" && *input.Key == "test-key"
		}),
	).Return(&s3.GetObjectOutput{
		Body:          testContentReader,
		ContentLength: aws.Int64(contentLength),
	}, nil)

	mockOneDriveService.On(
		"UploadSmallFile",
		"test-drive",
		"test-folder",
		"test-file.txt",
		contentLength,
	).Return(nil)

	service := NewServiceWithDependencies(
		nil,
		mockS3Client,
		mockOneDriveService,
		mockDBRepo,
	)

	params := SyncFileParams{
		Bucket:   "test-bucket",
		Key:      "test-key",
		DriveID:  "test-drive",
		FolderID: "test-folder",
		FileName: "test-file.txt",
	}

	err := service.SyncFile(params)

	assert.NoError(t, err)
	mockS3Client.AssertExpectations(t)
	mockOneDriveService.AssertExpectations(t)
}

func TestSyncFile_S3Error(t *testing.T) {
	mockS3Client := new(MockS3Client)
	mockOneDriveService := new(MockOneDriveService)
	mockDBRepo := new(MockDBRepository)

	expectedErr := errors.New("s3 error")
	mockS3Client.On("GetObject", mock.Anything, mock.Anything).Return(nil, expectedErr)

	service := NewServiceWithDependencies(
		nil,
		mockS3Client,
		mockOneDriveService,
		mockDBRepo,
	)

	params := SyncFileParams{
		Bucket:   "test-bucket",
		Key:      "test-key",
		DriveID:  "test-drive",
		FolderID: "test-folder",
		FileName: "test-file.txt",
	}

	err := service.SyncFile(params)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "couldn't get object")
	mockS3Client.AssertExpectations(t)
	mockOneDriveService.AssertNotCalled(t, "UploadSmallFile")
}

func TestSyncFile_OneDriveError(t *testing.T) {
	mockS3Client := new(MockS3Client)
	mockOneDriveService := new(MockOneDriveService)
	mockDBRepo := new(MockDBRepository)

	testContent := []byte("test file content")
	testContentReader := io.NopCloser(bytes.NewReader(testContent))
	contentLength := int64(len(testContent))

	mockS3Client.On("GetObject", mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{
		Body:          testContentReader,
		ContentLength: aws.Int64(contentLength),
	}, nil)

	expectedErr := errors.New("upload failed")
	mockOneDriveService.On("UploadSmallFile",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedErr)

	service := NewServiceWithDependencies(
		nil,
		mockS3Client,
		mockOneDriveService,
		mockDBRepo,
	)

	params := SyncFileParams{
		Bucket:   "test-bucket",
		Key:      "test-key",
		DriveID:  "test-drive",
		FolderID: "test-folder",
		FileName: "test-file.txt",
	}

	err := service.SyncFile(params)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upload small file")
	mockS3Client.AssertExpectations(t)
	mockOneDriveService.AssertExpectations(t)
}

