package onedrive

import (
	"fmt"

	"github.com/jaibhavaya/gogo-files/pkg/db"
)

func (s *Service) GetRefreshToken(ownerID int64) (string, error) {
	refreshToken, err := db.GetOneDriveRefreshToken(s.dbPool, ownerID)
	if err != nil {
		return "", fmt.Errorf("failed to get Refresh Token: %w", err)
	}

	return refreshToken, nil
}

func (s *Service) GetAccessToken(refreshToken string) (string, error) {
	accessToken, err := s.onedriveClient.getAccessToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to validate Onedrive Refresh Token %w", err)
	}

	return accessToken, nil
}

func (s *Service) SaveRefreshToken(ownerID int64, userID, refreshToken string) error {
	err := db.SaveOneDriveRefreshToken(
		s.dbPool,
		ownerID,
		userID,
		refreshToken,
	)
	if err != nil {
		return fmt.Errorf("failed to save OneDrive Refresh Token: %w", err)
	}

	return nil
}

func (s *Service) ValidateRefreshToken(refreshToken string) error {
	_, err := s.onedriveClient.getAccessToken(refreshToken)
	if err != nil {
		return fmt.Errorf("failed to validate Onedrive Refresh Token %w", err)
	}

	return nil
}
