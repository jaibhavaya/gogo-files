package handler

import (
	"fmt"

	"github.com/jaibhavaya/gogo-files/pkg/config"
	"github.com/jaibhavaya/gogo-files/pkg/db"
	"github.com/jaibhavaya/gogo-files/pkg/onedrive"
	"github.com/jaibhavaya/gogo-files/pkg/queue"
)

type Handler interface {
	Handle() error
}

func NewHandler(msg queue.Message, dbPool db.Pool, cfg *config.Config, onedriveClient onedrive.Client) (Handler, error) {
	// TODO: possibly add validation here since dbPool and onedriveClient are required for the handlers created below
	switch msg := msg.(type) {
	case *queue.OneDriveAuthorizationMessage:
		return &OneDriveAuthHandler{
			message:        *msg,
			dbPool:         dbPool,
			onedriveClient: onedriveClient,
		}, nil

	case *queue.FileSyncMessage:
		return &FileSyncHandler{
			message:        *msg,
			dbPool:         dbPool,
			onedriveClient: onedriveClient,
		}, nil
	}

	return nil, fmt.Errorf("Unknown Message Type %T", msg)
}
