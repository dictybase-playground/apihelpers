package nats

import (
	"context"
	"time"

	"github.com/dictyBase/apihelpers/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
)

type natsReplyMessage struct {
	msg *nats.Msg
	err error
}

func (m *natsReplyMessage) Message() []byte {
	return m.msg.Data
}

func (m *natsReplyMessage) Err() error {
	return m.err
}

type natsRequest struct {
	conn *nats.Conn
}

func NewRequest(url string, options ...nats.Option) (pubsub.Request, error) {
	nc, err := nats.Connect(url, options...)
	if err != nil {
		return &natsRequest{}, err
	}
	return &natsRequest{conn: nc}, nil
}

func (r *natsRequest) RawRequest(subj string, data []byte, timeout time.Duration) pubsub.ReplyMessage {
	msg, err := r.conn.Request(subj, data, timeout)
	if err != nil {
		return &natsReplyMessage{err: err}
	}
	return &natsReplyMessage{msg: msg}
}

func (r *natsRequest) Request(subj string, p proto.Message, timeout time.Duration) pubsub.ReplyMessage {
	data, err := proto.Marshal(p)
	if err != nil {
		return &natsReplyMessage{err: err}
	}
	return r.RawRequest(subj, data, timeout)
}

func (r *natsRequest) RawRequestWithContext(ctx context.Context, subj string, data []byte) pubsub.ReplyMessage {
	msg, err := r.conn.RequestWithContext(ctx, subj, data)
	if err != nil {
		return &natsReplyMessage{err: err}
	}
	return &natsReplyMessage{msg: msg}
}

func (r *natsRequest) RequestWithContext(ctx context.Context, subj string, p proto.Message) pubsub.ReplyMessage {
	data, err := proto.Marshal(p)
	if err != nil {
		return &natsReplyMessage{err: err}
	}
	return r.RawRequestWithContext(ctx, subj, data)
}
