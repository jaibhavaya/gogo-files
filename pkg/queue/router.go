package queue

import (
	"fmt"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
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
		ConcurrencyLimiter(5),
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
