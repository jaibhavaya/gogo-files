package event

import (
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
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
	EventType string                       `json:"event_type"`
	Payload   OneDriveAuthorizationPayload `json:"payload"`
}

func (m *OneDriveAuthorizationMessage) Type() string {
	return m.EventType
}

type FileSyncMessage struct {
	EventType string          `json:"event_type"`
	Payload   FileSyncPayload `json:"payload"`
}

func (m *FileSyncMessage) Type() string {
	return m.EventType
}

type MessageWrapper struct {
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

func parseMessage(msg *message.Message) (Message, error) {
	var wrapper MessageWrapper
	if err := json.Unmarshal([]byte(msg.Payload), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	switch wrapper.EventType {
	case "onedrive_authorization":
		var message OneDriveAuthorizationMessage
		message.EventType = wrapper.EventType
		if err := json.Unmarshal(wrapper.Payload, &message.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal onedrive authorization payload: %w", err)
		}
		return &message, nil

	case "file_sync":
		var message FileSyncMessage
		message.EventType = wrapper.EventType
		if err := json.Unmarshal(wrapper.Payload, &message.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal file sync payload: %w", err)
		}
		return &message, nil

	default:
		return nil, fmt.Errorf("unknown message type: %s", wrapper.EventType)
	}
}
