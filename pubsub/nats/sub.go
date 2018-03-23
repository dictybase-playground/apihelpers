package nats

import (
	"github.com/dictyBase/apihelpers/pubsub"
	gnats "github.com/nats-io/go-nats"
)

type natsMessage struct {
	msg *gnats.Msg
}

func (m *natsMessage) Message() []byte {
	return m.msg.Data
}

func (m *natsMessage) Done() error {
	if err := m.msg.Sub.Unsubscribe(); err != nil {
		return err
	}
	return nil
}

type natsSubscriber struct {
	conn   *gnats.Conn
	err    error
	output chan pubsub.SubscriberMessage
}

func NewSubscriber(url string, options ...gnats.Option) (pubsub.Subscriber, error) {
	nc, err := gnats.Connect(url, options...)
	if err != nil {
		return &natsSubscriber{}, err
	}
	return &natsSubscriber{
		conn:   nc,
		output: make(chan pubsub.SubscriberMessage),
	}, err
}

func (s *natsSubscriber) Start(sub string) <-chan pubsub.SubscriberMessage {
	nm := &natsMessage{}
	_, err := s.conn.Subscribe(sub, func(msg *gnats.Msg) {
		nm.msg = msg
		s.output <- nm
	})
	if err != nil {
		s.err = err
		return s.output
	}
	if err := s.conn.Flush(); err != nil {
		s.err = err
		return s.output
	}
	if err := s.conn.LastError(); err != nil {
		s.err = err
		return s.output
	}
	return s.output
}

func (s *natsSubscriber) Err() error {
	return s.err
}

func (s *natsSubscriber) Stop() error {
	s.conn.Close()
	close(s.output)
	return nil
}
