package nats

import (
	"github.com/dictyBase/apihelpers/pubsub"
	gnats "github.com/nats-io/go-nats"
)

type natsReply struct {
	conn *gnats.Conn
}

func NewReply(url string, options ...gnats.Option) (pubsub.Reply, error) {
	nc, err := gnats.Connect(url, options...)
	if err != nil {
		return &natsReply{}, err
	}
	return &natsReply{
		conn: nc,
	}, err
}

func (r *natsReply) Publish(subj string, data []byte) error {
	return r.conn.Publish(subj, data)
}

func (r *natsReply) Start(fn pubsub.ReplyFn, subj string) error {
	_, err := r.conn.Subscribe(subj, func(msg *gnats.Msg) {
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
