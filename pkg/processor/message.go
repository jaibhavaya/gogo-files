package processor

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ThreeDotsLabs/watermill/message"
)

type Message interface {
	Type() string
}

type MessageWrapper struct {
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

// Onedrive

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

// S3
// these structs are used for ease of unmarshalling
// and then are flattened into an S3OpsMessage for cleaner usage
type RawS3Event struct {
	Records []RawRecord `json:"Records"`
}

// for unmarshalling
type RawRecord struct {
	EventName string    `json:"eventName"`
	S3        RawS3Data `json:"s3"`
}

// for unmarshalling
type RawS3Data struct {
	Bucket struct {
		Name string `json:"name"`
	} `json:"bucket"`
	Object struct {
		Key      string        `json:"key"`
		Size     int64         `json:"size"`
		Metadata RawS3Metadata `json:"metadata"`
	} `json:"object"`
}

// for unmarshalling
// TODO find a way to make firm_id generic
// if we have control over the metadata for this event, have it changed to owner_id
type RawS3Metadata struct {
	OwnerID  int64  `json:"owner-id"`
	UserID   string `json:"user-id"`
	FilePath string `json:"file-path"`
	UploadTo string `json:"upload-to"`
}

type Record struct {
	EventName  string
	BucketName string
	ObjectKey  string
	ObjectSize int64
	OwnerID    int64
	UserID     string
	FilePath   string
	UploadTo   string
}

type S3OpsMessage struct {
	EventType string
	Records   []Record
}

// this function will flatten our s3 event
func ParseS3Event(data []byte) (S3OpsMessage, error) {
	var event RawS3Event
	if err := json.Unmarshal(data, &event); err != nil {
		return S3OpsMessage{}, err
	}

	records := make([]Record, len(event.Records))
	for _, record := range event.Records {
		records = append(records, Record{
			EventName:  record.EventName,
			BucketName: record.S3.Bucket.Name,
			ObjectKey:  record.S3.Object.Key,
			OwnerID:    record.S3.Object.Metadata.OwnerID,
			UserID:     record.S3.Object.Metadata.UserID,
			FilePath:   record.S3.Object.Metadata.FilePath,
			UploadTo:   record.S3.Object.Metadata.UploadTo,
		})
	}

	return S3OpsMessage{
		EventType: "s3_op",
		Records:   records,
	}, nil
}

func (m *S3OpsMessage) Type() string {
	return m.EventType
}

const (
	ONEDRIVE_AUTH_MESSAGE_TYPE = "onedrive_authorization"
	S3_OP_MESSAGE_TYPE         = "s3_op"
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

	case S3_OP_MESSAGE_TYPE:
		message, err := ParseS3Event([]byte(msg.Payload))
		if err != nil {
			return nil, err
		}
		return &message, nil

	default:
		return nil, fmt.Errorf("unknown message type: %s", wrapper.EventType)
	}
}
