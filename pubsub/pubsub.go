package pubsub

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
)

// Publisher is a generic interface to encapsulate how we want our publishers
// to behave. Until we find reason to change, we're forcing all pubslishers
// to emit protobufs.
type Publisher interface {
	// Publish will publish a message.
	Publish(string, proto.Message) error
	// Publish will publish a raw byte array as a message.
	PublishRaw(string, []byte) error
}

// Subscriber is a generic interface to encapsulate how we want our subscribers
// to behave. For now the system will auto stop if it encounters any errors. If
// a user encounters a closed channel, they should check the Err() method to see
// what happened.
type Subscriber interface {
	// Start will return a channel of raw messages.
	Start(string) <-chan SubscriberMessage
	// Err will contain any errors returned from the consumer connection.
	Err() error
	// Stop will initiate a graceful shutdown of the subscriber connection.
	Stop() error
}

// SubscriberMessage is a struct to encapsulate subscriber messages and provide
// a mechanism for acknowledging messages _after_ they've been processed.
type SubscriberMessage interface {
	Message() []byte
	Done() error
}

// ReplyMessage is a struct to encapsulate a reply message from a request reply conversation
type ReplyMessage interface {
	// Message will be the reply message
	Message() []byte
	// Err will return any error that happened during reply
	Err() error
}

// Request is generic interface to encapsulate how the Request part of point to point
// request-reply communication will behave
type Request interface {
	// Request send protobuf encoded request with a specified timeout
	Request(string, proto.Message, time.Duration) ReplyMessage
	// Request send raw byte array request with a specified timeout
	RawRequest(string, []byte, time.Duration) ReplyMessage
	// RequestWithContext uses context to send protobuf encoded request
	RequestWithContext(context.Context, string, proto.Message) ReplyMessage
	// RawRequestWithContext uses context to send a raw byte array request
	RawRequestWithContext(context.Context, string, []byte) ReplyMessage
}

type ReplyFn func([]byte) []byte

// Reply is a generic interface to encapsulate how the reply part of request-reply
// communication will behave
type Reply interface {
	Publish(string, []byte) error
	// Start will start the subscription, get the reply message
	// through ReplyFn and then reply message through Publish
	Start(ReplyFn, string) error
	// Stop will initiate a graceful shutdown of the subscriber connection.
	Stop() error
}
