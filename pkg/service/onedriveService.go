package service

import (
	"fmt"

	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/jaibhavaya/gogo-files/pkg/onedrive"
)

type OnedriveService struct {
	dbPool         *db.Pool
	onedriveClient *onedrive.Client
}

func NewOnedriveService(dbPool *db.Pool, cfg config.Config) *OnedriveService {
	return &OnedriveService{
		dbPool:         dbPool,
		onedriveClient: onedrive.NewClient(cfg.OnedriveClientID, cfg.OnedriveClientSecret),
	}
}

func (s *OnedriveService) GetRefreshToken(ownerID int64) (string, error) {
	refreshToken, err := db.GetOneDriveRefreshToken(s.dbPool, ownerID)
	if err != nil {
		return "", fmt.Errorf("failed to get Refresh Token: %w", err)
	}

	return refreshToken, nil
}

func (s *OnedriveService) GetAccessToken(refreshToken string) (string, error) {
	accessToken, err := s.onedriveClient.GetAccessToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to validate Onedrive Refresh Token %w", err)
	}

	return accessToken, nil
}

func (s *OnedriveService) SaveRefreshToken(ownerID int64, userID, refreshToken string) error {
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

func (s *OnedriveService) ValidateRefreshToken(refreshToken string) error {
	_, err := s.onedriveClient.GetAccessToken(refreshToken)
	if err != nil {
		return fmt.Errorf("Failed to validate Onedrive Refresh Token %w", err)
	}

	return nil
}
