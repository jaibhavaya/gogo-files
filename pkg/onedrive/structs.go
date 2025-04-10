package onedrive

import (
	"net/http"
	"time"

	"github.com/jaibhavaya/gogo-files/pkg/db"
)

type Client struct {
	dbPool               *db.Pool
	encryptionKey        string
	onedriveClientID     string
	onedriveClientSecret string
	httpClient           *http.Client
}

func NewClient(dbPool *db.Pool, encryptionKey, clientID, clientSecret string) *Client {
	return &Client{
		dbPool:               dbPool,
		encryptionKey:        encryptionKey,
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
