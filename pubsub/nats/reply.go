package nats

import (
	"github.com/dictyBase/apihelpers/pubsub"
	gnats "github.com/nats-io/go-nats"
)

type natsReply struct {
	conn         *gnats.Conn
	subscription string
}

func NewReply(url string, sub string) (pubsub.Reply, error) {
	nc, err := gnats.Connect(url)
	if err != nil {
		return &natsReply{}, err
	}
	return &natsReply{
		conn:         nc,
		subscription: sub,
	}, err
}

func (r *natsReply) Publish(subj string, data []byte) error {
	return r.conn.Publish(subj, data)
}

func (r *natsReply) Start(fn pubsub.ReplyFn) error {
	_, err := r.conn.Subscribe(r.subscription, func(msg *gnats.Msg) {
		r.Publish(msg.Reply, fn(msg.Data))
	})
	if err != nil {
		return err
	}
	if err := r.conn.Flush(); err != nil {
		return err
	}
	if err := r.conn.LastError(); err != nil {
		return err
	}
	return nil
}

func (r *natsReply) Stop() error {
	r.conn.Close()
	return nil
}
