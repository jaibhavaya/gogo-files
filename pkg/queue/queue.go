package queue

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/jaibhavaya/gogo-files/pkg/messages"

	amazonsqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	transport "github.com/aws/smithy-go/endpoints"
	"github.com/samber/lo"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
)

// SQSProcessor manages the SQS message processing architecture
type SQSProcessor struct {
	logger           watermill.LoggerAdapter
	subscriberConfig sqs.SubscriberConfig
	publisherConfig  sqs.PublisherConfig
	queueName        string
	numSubscribers   int
	numWorkers       int
	messageChan      chan *message.Message
	publisher        message.Publisher
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
}

// NewSQSProcessor creates a new SQS processor
func NewSQSProcessor(queueName string, numSubscribers, numWorkers int) *SQSProcessor {
	logger := watermill.NewStdLogger(false, false)
	ctx, cancel := context.WithCancel(context.Background())
	sqsOpts := []func(*amazonsqs.Options){
		amazonsqs.WithEndpointResolverV2(sqs.OverrideEndpointResolver{
			Endpoint: transport.Endpoint{
				URI: *lo.Must(url.Parse("http://localhost:4566")),
			},
		}),
	}

	subscriberConfig := sqs.SubscriberConfig{
		AWSConfig: aws.Config{
			Credentials: aws.AnonymousCredentials{},
			Region:      "us-west-1",
		},
		OptFns: sqsOpts,
	}

	publisherConfig := sqs.PublisherConfig{
		AWSConfig: aws.Config{
			Credentials: aws.AnonymousCredentials{},
			Region:      "us-west-1",
		},
		OptFns: sqsOpts,
	}

	return &SQSProcessor{
		logger:           logger,
		subscriberConfig: subscriberConfig,
		publisherConfig:  publisherConfig,
		queueName:        queueName,
		numSubscribers:   numSubscribers,
		numWorkers:       numWorkers,
		messageChan:      make(chan *message.Message, 100),
		ctx:              ctx,
		cancel:           cancel,
	}
}

// Start begins SQS message processing
func (p *SQSProcessor) Start() error {
	// Initialize publisher
	var err error
	p.publisher, err = sqs.NewPublisher(p.publisherConfig, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create publisher: %w", err)
	}

	// Start multiple subscribers
	for i := range p.numSubscribers {
		if err := p.startSubscriber(i); err != nil {
			return err
		}
	}

	// Start multiple workers
	p.startWorkers()

	return nil
}

// StartPublishing begins publishing test messages
func (p *SQSProcessor) StartPublishing() {
	go p.publishMessages()
}

// Stop gracefully shuts down processing
func (p *SQSProcessor) Stop() {
	p.cancel()
	p.wg.Wait()
	close(p.messageChan)
}

// startSubscriber creates and starts a single subscriber
func (p *SQSProcessor) startSubscriber(subscriberID int) error {
	subscriber, err := sqs.NewSubscriber(p.subscriberConfig, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create subscriber %d: %w", subscriberID, err)
	}

	messages, err := subscriber.Subscribe(p.ctx, p.queueName)
	if err != nil {
		return fmt.Errorf("failed to subscribe to queue %s: %w", p.queueName, err)
	}

	p.wg.Add(1)
	go func(subID int) {
		defer p.wg.Done()
		for {
			select {
			case msg, ok := <-messages:
				if !ok {
					log.Printf("Subscriber %d channel closed", subID)
					return
				}
				log.Printf("Subscriber %d received message: %s", subID, msg.UUID)
				p.messageChan <- msg
			case <-p.ctx.Done():
				log.Printf("Subscriber %d shutting down", subID)
				return
			}
		}
	}(subscriberID)

	return nil
}

// startWorkers creates and starts worker goroutines
func (p *SQSProcessor) startWorkers() {
	for i := range p.numWorkers {
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()

			log.Printf("Starting worker %d", workerID)
			for {
				select {
				case msg, ok := <-p.messageChan:
					if !ok {
						log.Printf("Worker %d shutting down - channel closed", workerID)
						return
					}
					log.Printf("Worker %d processing message: %s, payload: %s",
						workerID, msg.UUID, string(msg.Payload))

					p.processMessage(msg, workerID)
					msg.Ack()
				case <-p.ctx.Done():
					log.Printf("Worker %d shutting down - context canceled", workerID)
					return
				}
			}
		}(i)
	}
}

// processMessage handles the processing of a single message
func (p *SQSProcessor) processMessage(msg *message.Message, workerID int) {
	startTime := time.Now()
	log.Printf("Worker %d STARTED processing message %s at %v",
		workerID, msg.UUID, startTime.Format(time.RFC3339))

	messageHandler, err := messages.ParseMessage(msg)
	if err != nil {
		log.Printf("Error Parsing Message: %v", err)
	}

	log.Printf("messageHandler: %v", messageHandler)

	// Simulate actual work
	time.Sleep(10 * time.Second)

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("Worker %d FINISHED processing message %s at %v (took %v)",
		workerID, msg.UUID, endTime.Format(time.RFC3339), duration)
}

// publishMessages continuously publishes test messages
func (p *SQSProcessor) publishMessages() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			msg := message.NewMessage(watermill.NewUUID(), []byte("Hello, world!"))
			if err := p.publisher.Publish("example-topic", msg); err != nil {
				log.Printf("Error publishing message: %v", err)
			}
			time.Sleep(time.Second)
		}
	}
}

// PublishMessage publishes a single message
func (p *SQSProcessor) PublishMessage(payload []byte) error {
	msg := message.NewMessage(watermill.NewUUID(), payload)
	return p.publisher.Publish("example-topic", msg)
}
