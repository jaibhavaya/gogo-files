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

func (m *FileSyncMessage) Type() string {
	return "file_sync"
}

type MessageWrapper struct {
	MessageType string          `json:"message_type"`
	Payload     json.RawMessage `json:"payload"`
}

func ParseMessage(messageBody string) (Message, error) {
	var wrapper MessageWrapper
	if err := json.Unmarshal([]byte(messageBody), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	switch wrapper.MessageType {
	case "onedrive_authorization":
		var message OneDriveAuthorizationMessage
		message.MessageType = wrapper.MessageType
		if err := json.Unmarshal(wrapper.Payload, &message.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal onedrive authorization payload: %w", err)
		}
		return &message, nil

	case "file_sync":
		var message FileSyncMessage
		message.MessageType = wrapper.MessageType
		if err := json.Unmarshal(wrapper.Payload, &message.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal file sync payload: %w", err)
		}
		return &message, nil

	default:
		return nil, fmt.Errorf("unknown message type: %s", wrapper.MessageType)
	}
}
