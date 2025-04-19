package onedrive

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) DoRequest(method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	args := m.Called(method, path, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
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

func TestGetRefreshToken_Success(t *testing.T) {
	mockClient := new(MockHTTPClient)
	mockRepository := new(MockDBRepository)

	expectedToken := "test-refresh-token"
	mockRepository.On("GetOneDriveRefreshToken", int64(123)).Return(expectedToken, nil)

	service := NewServiceWithDependencies(
		nil,
		mockClient,
		mockRepository,
	)

	token, err := service.GetRefreshToken(123)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	mockRepository.AssertExpectations(t)
}

func TestGetRefreshToken_Error(t *testing.T) {
	mockClient := new(MockHTTPClient)
	mockRepository := new(MockDBRepository)

	expectedErr := errors.New("database error")
	mockRepository.On("GetOneDriveRefreshToken", int64(123)).Return("", expectedErr)

	service := NewServiceWithDependencies(
		nil,
		mockClient,
		mockRepository,
	)

	token, err := service.GetRefreshToken(123)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get Refresh Token")
	assert.Equal(t, "", token)
	mockRepository.AssertExpectations(t)
}

func TestUploadSmallFile_Success(t *testing.T) {
	mockClient := new(MockHTTPClient)
	mockRepository := new(MockDBRepository)

	fileContent := []byte("test file content")
	fileSize := int64(len(fileContent))
	reader := bytes.NewReader(fileContent)

	responseBody := io.NopCloser(strings.NewReader(`{"id": "123"}`))
	response := &http.Response{
		StatusCode: 200,
		Body:       responseBody,
	}

	expectedPath := "/drives/test-drive/root:/Documents/Reports/test-file.txt:/content"

	mockClient.On("DoRequest", "PUT", expectedPath, mock.Anything).Return(response, nil)

	service := NewServiceWithDependencies(
		nil,
		mockClient,
		mockRepository,
	)

	err := service.UploadSmallFile("test-drive", "/Documents/Reports", "test-file.txt", reader, fileSize)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestUploadSmallFile_RequestError(t *testing.T) {
	mockClient := new(MockHTTPClient)
	mockRepository := new(MockDBRepository)

	fileContent := []byte("test file content")
	fileSize := int64(len(fileContent))
	reader := bytes.NewReader(fileContent)

	expectedErr := errors.New("network error")
	mockClient.On("DoRequest", "PUT", mock.Anything, mock.Anything).Return(nil, expectedErr)

	service := NewServiceWithDependencies(
		nil,
		mockClient,
		mockRepository,
	)

	err := service.UploadSmallFile("test-drive", "/Documents/Reports", "test-file.txt", reader, fileSize)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error sending request")
	mockClient.AssertExpectations(t)
}

func TestUploadSmallFile_BadResponse(t *testing.T) {
	mockClient := new(MockHTTPClient)
	mockRepository := new(MockDBRepository)

	fileContent := []byte("test file content")
	fileSize := int64(len(fileContent))
	reader := bytes.NewReader(fileContent)

	responseBody := io.NopCloser(strings.NewReader(`{"error": "Bad request"}`))
	response := &http.Response{
		StatusCode: 400,
		Body:       responseBody,
	}

	mockClient.On("DoRequest", "PUT", mock.Anything, mock.Anything).Return(response, nil)

	service := NewServiceWithDependencies(
		nil,
		mockClient,
		mockRepository,
	)

	err := service.UploadSmallFile("test-drive", "/Documents/Reports", "test-file.txt", reader, fileSize)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload failed with status 400")
	mockClient.AssertExpectations(t)
}

