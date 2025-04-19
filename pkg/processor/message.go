package processor

import (
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
)

type Message interface {
	Type() string
}

type MessageWrapper struct {
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

type OneDriveAuthorizationMessage struct {
	EventType string                       `json:"event_type"`
	Payload   OneDriveAuthorizationPayload `json:"payload"`
}

type OneDriveAuthorizationPayload struct {
	OwnerID      int64  `json:"owner_id"`
	UserID       string `json:"user_id"`
	RefreshToken string `json:"refresh_token"`
}

func (m *OneDriveAuthorizationMessage) Type() string {
	return m.EventType
}

type FileSyncMessage struct {
	EventType string          `json:"event_type"`
	Payload   FileSyncPayload `json:"payload"`
}

type FileSyncPayload struct {
	OwnerID int64          `json:"owner_id"`
	UserID  string         `json:"user_id"`
	Items   []FileSyncItem `json:"items"`
}

type FileSyncItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Size   int    `json:"size"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

func (m *FileSyncMessage) Type() string {
	return m.EventType
}

const (
	ONEDRIVE_AUTH_MESSAGE_TYPE = "onedrive_authorization"
	FILE_SYNC_MESSAGE_TYPE     = "file_sync"
)

func parseMessage(msg *message.Message) (Message, error) {
	var wrapper MessageWrapper
	if err := json.Unmarshal([]byte(msg.Payload), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	switch wrapper.EventType {
	case ONEDRIVE_AUTH_MESSAGE_TYPE:
		var message OneDriveAuthorizationMessage
		message.EventType = wrapper.EventType
		if err := json.Unmarshal(wrapper.Payload, &message.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal onedrive authorization payload: %w", err)
		}
		return &message, nil

	case FILE_SYNC_MESSAGE_TYPE:
		var message FileSyncMessage
		message.EventType = wrapper.EventType
		if err := json.Unmarshal(wrapper.Payload, &message.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal onedrive authorization payload: %w", err)
		}
		return &message, nil

	default:
		return nil, fmt.Errorf("unknown message type: %s", wrapper.EventType)
	}
}
