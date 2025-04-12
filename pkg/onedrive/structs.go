package onedrive

import (
	"net/http"
	"time"
)

type Client struct {
	onedriveClientID     string
	onedriveClientSecret string
	httpClient           *http.Client
}

func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		onedriveClientID:     clientID,
		onedriveClientSecret: clientSecret,
		httpClient:           &http.Client{Timeout: 30 * time.Second},
	}
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}
