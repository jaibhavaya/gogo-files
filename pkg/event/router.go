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
	p.addHandler()

	return nil
}

func (p *SQSProcessor) createRouter() error {
	var err error
	p.publisher, err = sqs.NewPublisher(p.publisherConfig, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create publisher: %w", err)
	}

	p.subscriber, err = sqs.NewSubscriber(p.subscriberConfig, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %w", err)
	}

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

func (p *SQSProcessor) addHandler() {
	p.router.AddHandler(
		"Dwayne-TheRock-Johnson",
		"gogo-files-queue",
		p.subscriber,
		"notification-queue",
		p.publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			log.Printf("Processing message: %s", msg.UUID)

			p.processMessage(msg)

			// Create a notification message to publish
			notificationMsg := message.NewMessage(
				watermill.NewUUID(),
				[]byte("Processing completed successfully"),
			)

			// Return the message to be published
			return []*message.Message{notificationMsg}, nil
		},
	)
}

func concurrencyLimiter(maxConcurrent int) message.HandlerMiddleware {
	semaphore := make(chan struct{}, maxConcurrent)

	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			semaphore <- struct{}{} // Acquire a slot
			defer func() {
				<-semaphore // Release the slot when done
			}()

			return h(msg)
		}
	}
}

func (p *SQSProcessor) processMessage(msg *message.Message) {
	defer logEnd(logStart(msg))

	// TODO error handling in terms of what to do with the event
	// requeue? depends on type of error

	message, err := parseMessage(msg)
	if err != nil {
		log.Printf("Error Parsing Message: %v", err)
	}

	handler, err := p.handlerForMessage(message)
	if err != nil {
		log.Printf("Failed to get handler for message: %v", err)
	}

	err = handler.Handle()
	if err != nil {
		log.Printf("Failed to handle message %v", err)
	}
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
