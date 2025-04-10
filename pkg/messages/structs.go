package messages

import (
	"encoding/json"
	"fmt"
)

type Message interface {
	Type() string
}

type OneDriveAuthorizationPayload struct {
	OwnerID      int64  `json:"owner_id"`
	UserID       string `json:"user_id"`
	RefreshToken string `json:"refresh_token"`
}

type FileSyncPayload struct {
	OwnerID     int64  `json:"owner_id"`
	Bucket      string `json:"bucket"`
	Key         string `json:"key"`
	Destination string `json:"destination"`
}

type OneDriveAuthorizationMessage struct {
	MessageType string                       `json:"message_type"`
	Payload     OneDriveAuthorizationPayload `json:"payload"`
}

func (m *OneDriveAuthorizationMessage) Type() string {
	return "onedrive_authorization"
}

type FileSyncMessage struct {
	MessageType string          `json:"message_type"`
	Payload     FileSyncPayload `json:"payload"`
}

type MessageWrapper struct {
	MessageType string          `json:"message_type"`
	Payload     json.RawMessage `json:"payload"`
}
