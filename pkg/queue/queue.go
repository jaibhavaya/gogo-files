package queue

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/jaibhavaya/gogo-files/pkg/handler"
	"github.com/jaibhavaya/gogo-files/pkg/service"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	transport "github.com/aws/smithy-go/endpoints"
	"github.com/samber/lo"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

type SQSProcessor struct {
	logger           watermill.LoggerAdapter
	queueName        string
	numSubscribers   int
	numWorkers       int
	messageChan      chan *message.Message
	router           *message.Router
	routerConfig     message.RouterConfig
	subscriber       message.Subscriber
	subscriberConfig sqs.SubscriberConfig
	publisher        message.Publisher
	publisherConfig  sqs.PublisherConfig
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	fileService      *service.FileService
	onedriveService  *service.OnedriveService
}

func NewSQSProcessor(
	queueName string,
	numSubscribers, numWorkers int,
	onedriveService *service.OnedriveService,
	fileService *service.FileService,
) *SQSProcessor {
	logger := watermill.NewStdLogger(false, false)
	ctx, cancel := context.WithCancel(context.Background())
	sqsOpts := []func(*awssqs.Options){
		awssqs.WithEndpointResolverV2(sqs.OverrideEndpointResolver{
			Endpoint: transport.Endpoint{
				URI: *lo.Must(url.Parse("http://localhost:4566")),
			},
		}),
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	subscriberConfig := sqs.SubscriberConfig{
		AWSConfig: awsCfg,
		GenerateReceiveMessageInput: func(ctx context.Context, queueURL sqs.QueueURL) (*awssqs.ReceiveMessageInput, error) {
			return &awssqs.ReceiveMessageInput{
				QueueUrl:            aws.String(string(queueURL)),
				MaxNumberOfMessages: int32(10),
				WaitTimeSeconds:     int32(20),
			}, nil
		},
		OptFns: sqsOpts,
	}

	publisherConfig := sqs.PublisherConfig{
		AWSConfig: awsCfg,
		OptFns:    sqsOpts,
	}

	routerConfig := message.RouterConfig{
		CloseTimeout: time.Second * 30,
	}

	return &SQSProcessor{
		logger:           logger,
		subscriberConfig: subscriberConfig,
		publisherConfig:  publisherConfig,
		routerConfig:     routerConfig,
		queueName:        queueName,
		numSubscribers:   numSubscribers,
		numWorkers:       numWorkers,
		messageChan:      make(chan *message.Message, 100),
		ctx:              ctx,
		cancel:           cancel,
		onedriveService:  onedriveService,
		fileService:      fileService,
	}
}

func (p *SQSProcessor) Start() error {
	err := p.setup()
	if err != nil {
		log.Fatalf("Failed to start Queue Processor: %v", err)
	}

	log.Println("Starting SQS message router...")
	if err := p.router.Run(p.ctx); err != nil {
		log.Fatalf("Router error: %v", err)
	}

	return nil
}

func (p *SQSProcessor) StartPublishing() {
	go p.publishMessages()
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
