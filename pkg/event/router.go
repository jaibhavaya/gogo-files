package event

import (
	"fmt"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/jaibhavaya/gogo-files/pkg/handler"
)

func (p *SQSProcessor) setup() error {
	err := p.createRouter()
	if err != nil {
		return fmt.Errorf("failed to start Queue Processor: %v", err)
	}

	p.addMiddleware()
	p.addAuthHandler()
	p.addSyncHandler()
	p.addOpsHandler()

	return nil
}

func (p *SQSProcessor) createRouter() error {
	var err error
	p.router, err = message.NewRouter(p.routerConfig, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create router: %v", err)
	}
	return nil
}

func (p *SQSProcessor) addMiddleware() {
	p.router.AddMiddleware(
		middleware.NewThrottle(10, time.Second).Middleware,
		middleware.Recoverer,
		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: time.Second,
			Logger:          p.logger,
		}.Middleware,
		concurrencyLimiter(5),
	)
}

func (p *SQSProcessor) addSyncHandler() error {
	publisher, err := sqs.NewPublisher(p.publisherConfig, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create publisher: %w", err)
	}

	subscriber, err := sqs.NewSubscriber(p.subscriberConfig, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %w", err)
	}

	p.router.AddHandler(
		"Sync",
		"one-drive-sync",
		subscriber,
		"one-drive-status",
		publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			log.Printf("Processing message: %s", msg.UUID)

			// handle initial syncs
			// p.processMessage(msg)

			notificationMsg := message.NewMessage(
				watermill.NewUUID(),
				[]byte("Processing completed successfully"),
			)

			// TODO: publish status to publish queue
			return []*message.Message{notificationMsg}, nil
		},
	)

	return nil
}

func (p *SQSProcessor) addOpsHandler() error {
	publisher, err := sqs.NewPublisher(p.publisherConfig, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create publisher: %w", err)
	}

	subscriber, err := sqs.NewSubscriber(p.subscriberConfig, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %w", err)
	}

	p.router.AddHandler(
		"Ops",
		"one-drive-ops",
		subscriber,
		"one-drive-status",
		publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			log.Printf("Processing message: %s", msg.UUID)

			// handle file/folder ops
			// p.processMessage(msg)

			notificationMsg := message.NewMessage(
				watermill.NewUUID(),
				[]byte("Processing completed successfully"),
			)

			// TODO: publish status to publish queue
			return []*message.Message{notificationMsg}, nil
		},
	)

	return nil
}

func (p *SQSProcessor) addAuthHandler() error {
	subscriber, err := sqs.NewSubscriber(p.subscriberConfig, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %w", err)
	}

	p.router.AddNoPublisherHandler(
		"Auth",
		"one-drive-auth",
		subscriber,
		func(msg *message.Message) error {
			log.Printf("Processing message: %s", msg.UUID)

			err := p.processMessage(msg)
			if err != nil {
				// Figure out what to do on error here
				log.Printf("failed to process auth message: %v\n", err)
			}

			return nil
		},
	)

	return nil
}

func concurrencyLimiter(maxConcurrent int) message.HandlerMiddleware {
	semaphore := make(chan struct{}, maxConcurrent)

	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			return h(msg)
		}
	}
}

func (p *SQSProcessor) processMessage(msg *message.Message) error {
	defer logEnd(logStart(msg))

	// TODO error handling in terms of what to do with the event
	// requeue? depends on type of error

	message, err := parseMessage(msg)
	if err != nil {
		return fmt.Errorf("error Parsing Message: %v", err)
	}

	handler, err := p.handlerForMessage(message)
	if err != nil {
		return fmt.Errorf("error retrieving handler for message: %v", err)
	}

	err = handler.Handle()
	if err != nil {
		return fmt.Errorf("failed to handle message %v", err)
	}

	return nil
}

func (p *SQSProcessor) handlerForMessage(msg Message) (handler.Handler, error) {
	switch msg := msg.(type) {
	case *OneDriveAuthorizationMessage:
		return handler.NewOnedriveAuthHandler(
			msg.Payload.OwnerID,
			msg.Payload.UserID,
			msg.Payload.RefreshToken,
			p.onedriveService,
		), nil

	case *FileSyncMessage:
		return handler.NewFileSyncHandler(
			msg.Payload.OwnerID,
			msg.Payload.Key,
			msg.Payload.Bucket,
			msg.Payload.Destination,
			p.onedriveService,
			p.fileService,
		), nil
	}

	return nil, fmt.Errorf("unknown Message Type %T", msg)
}
