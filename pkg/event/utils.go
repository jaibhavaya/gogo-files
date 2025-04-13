package event

import (
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

func logStart(msg *message.Message) (*message.Message, time.Time) {
	startTime := time.Now()
	log.Printf("STARTED processing message %s at %v",
		msg.UUID, startTime.Format(time.RFC3339))

	return msg, startTime
}

func logEnd(msg *message.Message, startTime time.Time) {
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("FINISHED processing message %s at %v (took %v)",
		msg.UUID, endTime.Format(time.RFC3339), duration)
}
