package onedrive

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jaibhavaya/gogo-files/pkg/db"
)

type client struct {
	onedriveClientID     string
	onedriveClientSecret string
	refreshToken         string
	httpClient           *http.Client
}

func newClient(onedriveIntegration *db.OneDriveIntegration, clientID, clientSecret string) *client {
	return &client{
		onedriveClientID:     clientID,
		onedriveClientSecret: clientSecret,
		refreshToken:         onedriveIntegration.RefreshToken,
		httpClient:           &http.Client{Timeout: 30 * time.Second},
	}
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

func (c *client) getAccessToken() (string, error) {
	formData := url.Values{}
	formData.Set("grant_type", "refresh_token")
	formData.Set("refresh_token", c.refreshToken)
	formData.Set("client_id", c.onedriveClientID)
	formData.Set("client_secret", c.onedriveClientSecret)

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	resp, err := c.DoRequest(
		"POST",
		"https://login.microsoftonline.com/common/oauth2/v2.0/token",
		strings.NewReader(formData.Encode()),
		headers,
	)
	if err != nil {
		return "", fmt.Errorf("failed to fetch token")
	}

	var response tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return response.AccessToken, nil
}

func (c *client) DoRequest(method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	accessToken, err := c.getAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	fullURL := fmt.Sprintf("https://graph.microsoft.com/v1.0%s", path)

	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &http.Response{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &http.Response{}, fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	return resp, nil
}
