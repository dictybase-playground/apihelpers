package nats

import (
	"github.com/dictyBase/apihelpers/pubsub"
	"github.com/golang/protobuf/proto"
	gnats "github.com/nats-io/go-nats"
)

type natsPublisher struct {
	conn *gnats.Conn
}

func NewPublisher(url string, options ...gnats.Option) (pubsub.Publisher, error) {
	nc, err := gnats.Connect(url, options...)
	if err != nil {
		return &natsPublisher{}, err
	}
	return &natsPublisher{conn: nc}, nil
}

// PublishRaw publishes a raw byte array as a message.
func (p *natsPublisher) PublishRaw(subj string, data []byte) error {
	return p.conn.Publish(subj, data)
}

// Publish will publish a protobuf encoded message.
func (p *natsPublisher) Publish(subj string, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return p.PublishRaw(subj, data)
}
