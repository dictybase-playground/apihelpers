package nats

import (
	"github.com/dictyBase/apihelpers/pubsub"
	"github.com/nats-io/go-nats"
)

type natsMessage struct {
	msg *nats.Msg
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
	conn         *nats.Conn
	subscription string
	err          error
	output       chan pubsub.SubscriberMessage
}

func NewSubscriber(url string, sub string) (pubsub.Subscriber, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return &natsSubscriber{}, err
	}
	return &natsSubscriber{
		conn:         nc,
		subscription: sub,
		output:       make(chan pubsub.SubscriberMessage),
	}, err
}

func (s *natsSubscriber) Start() <-chan pubsub.SubscriberMessage {
	nm := &natsMessage{}
	_, err := s.conn.Subscribe(s.subscription, func(msg *nats.Msg) {
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
