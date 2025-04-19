package onedrive

import (
	"fmt"
	"io"
	"net/url"
	"strings"
)

func (s *Service) GetRefreshToken(ownerID int64) (string, error) {
	refreshToken, err := s.repository.GetOneDriveRefreshToken(ownerID)
	if err != nil {
		return "", fmt.Errorf("failed to get Refresh Token: %w", err)
	}

	return refreshToken, nil
}

func (s *Service) SaveRefreshToken(ownerID int64, userID, refreshToken string) error {
	err := s.repository.SaveOneDriveRefreshToken(
		ownerID,
		userID,
		refreshToken,
	)
	if err != nil {
		return fmt.Errorf("failed to save OneDrive Refresh Token: %w", err)
	}

	return nil
}

func (s *Service) UploadSmallFile(driveID, folderPath, fileName string, fileContent io.Reader, fileSize int64) error {
	folderPath = strings.TrimPrefix(folderPath, "/")
	folderPath = strings.TrimSuffix(folderPath, "/")

	apiPath := fmt.Sprintf(
		"/drives/%s/root:/%s/%s:/content",
		driveID, folderPath, url.PathEscape(fileName),
	)

	headers := map[string]string{
		"Content-Type":   "application/octet-stream",
		"Content-Length": fmt.Sprintf("%d", fileSize),
	}

	resp, err := s.client.DoRequest("PUT", apiPath, fileContent, headers)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

