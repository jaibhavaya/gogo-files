package onedrive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jaibhavaya/gogo-files/pkg/db"
)

// Client represents a OneDrive API client
type Client struct {
	dbPool              *db.Pool
	encryptionKey       string
	onedriveClientID    string
	onedriveClientSecret string
	httpClient          *http.Client
}

// NewClient creates a new OneDrive client
func NewClient(dbPool *db.Pool, encryptionKey, clientID, clientSecret string) *Client {
	return &Client{
		dbPool:              dbPool,
		encryptionKey:       encryptionKey,
		onedriveClientID:    clientID,
		onedriveClientSecret: clientSecret,
		httpClient:          &http.Client{Timeout: 30 * time.Second},
	}
}

// TokenResponse represents the response from the OAuth token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"` 
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

// GetAccessToken retrieves an access token for a OneDrive integration
func (c *Client) GetAccessToken(ownerID int64) (string, error) {
	// Get refresh token from database
	refreshToken, err := db.GetOneDriveRefreshToken(c.dbPool, ownerID, c.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to get refresh token: %w", err)
	}

	// Exchange refresh token for access token
	formData := url.Values{}
	formData.Set("grant_type", "refresh_token")
	formData.Set("refresh_token", refreshToken)
	formData.Set("client_id", c.onedriveClientID)
	formData.Set("client_secret", c.onedriveClientSecret)

	req, err := http.NewRequest("POST", "https://login.microsoftonline.com/common/oauth2/v2.0/token", 
		strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// TODO: Store the new refresh token and access token in the database
	// This would involve implementing a method to update the tokens

	return tokenResponse.AccessToken, nil
}

// UploadFile uploads a file to OneDrive
func (c *Client) UploadFile(ownerID int64, fileData []byte, destination string) error {
	// Get access token
	accessToken, err := c.GetAccessToken(ownerID)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Create upload session for large files
	createSessionURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/root:/%s:/createUploadSession", 
		url.PathEscape(destination))

	req, err := http.NewRequest("POST", createSessionURL, bytes.NewBufferString(`{}`)) 
	if err != nil {
		return fmt.Errorf("failed to create upload session request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create upload session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create upload session failed with status: %d", resp.StatusCode)
	}

	var sessionResponse struct {
		UploadURL string `json:"uploadUrl"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&sessionResponse); err != nil {
		return fmt.Errorf("failed to decode upload session response: %w", err)
	}

	// Upload file content
	uploadReq, err := http.NewRequest("PUT", sessionResponse.UploadURL, bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	uploadReq.Header.Set("Content-Length", fmt.Sprintf("%d", len(fileData)))
	uploadReq.Header.Set("Content-Range", fmt.Sprintf("bytes 0-%d/%d", len(fileData)-1, len(fileData)))

	uploadResp, err := c.httpClient.Do(uploadReq)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("file upload failed with status: %d", uploadResp.StatusCode)
	}

	return nil
}
