package messages

import (
	"encoding/json"
	"fmt"
)

func (m *FileSyncMessage) Type() string {
	return "file_sync"
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
