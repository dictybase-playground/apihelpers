package nats

import (
	"bytes"
	"log"
	"os"
	"testing"
	"time"

	"github.com/dictyBase/apihelpers/aphdocker"
)

var connString string

func TestMain(m *testing.M) {
	nats, err := aphdocker.NewNatsDocker()
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	resource, err := nats.Run()
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	_, err = nats.RetryNatsConnection()
	if err != nil {
		log.Fatal(err)
	}
	connString = nats.GetNatsConnString()
	code := m.Run()
	if err = nats.Purge(resource); err != nil {
		log.Fatalf("unable to remove container %s\n", err)
	}
	os.Exit(code)
}

func TestPublisher(t *testing.T) {
	p, err := NewPublisher(connString)
	if err != nil {
		t.Fatalf("cannot connect to nats %s\n", err)
	}
	if err := p.PublishRaw("test", []byte("testpub")); err != nil {
		t.Fatal("cannot publish raw byte array")
	}
}

func TestSubscriber(t *testing.T) {
	subj := "pubsubtest"
	p, err := NewPublisher(connString)
	if err != nil {
		t.Fatalf("cannot connect to nats for publishing %s\n", err)
	}
	s, err := NewSubscriber(connString)
	if err != nil {
		t.Fatalf("cannot connect to nats for subscription %s\n", err)
	}
	sc := s.Start(subj)
	if err := s.Err(); err != nil {
		t.Fatalf("could not start subscription %s\n", err)
	}
	pdata := []byte("yadayada")
	if err := p.PublishRaw(subj, pdata); err != nil {
		t.Fatal("cannot publish raw byte array")
	}
	msg := <-sc
	if !bytes.Equal(msg.Message(), pdata) {
		t.Fatalf("published message %s does not match %s\n", string(pdata), string(msg.Message()))
	}
	if err := s.Stop(); err != nil {
		t.Fatalf("could not stop the subscription %s\n", err)
	}
}

func TestReplyReq(t *testing.T) {
	subj := "replytest"
	req, err := NewRequest(connString)
	if err != nil {
		t.Fatalf("cannot connect to nats for request %s\n", err)
	}
	rep, err := NewReply(connString)
	if err != nil {
		t.Fatalf("cannot connect to nats for reply %s\n", err)
	}
	rdata := []byte("battery")
	fn := func(b []byte) []byte { return b }
	if err := rep.Start(fn, subj); err != nil {
		t.Fatalf("could not start reply subscription %s\n", err)
	}
	timeout, _ := time.ParseDuration("28s")
	msg := req.RawRequest(subj, rdata, timeout)
	if err := msg.Err(); err != nil {
		t.Fatalf("unable to get a reply %s\n", err)
	}
	if !bytes.Equal(rdata, msg.Message()) {
		t.Fatalf("request %s does not match reply %s\n", string(rdata), string(msg.Message()))
	}
	if err := rep.Stop(); err != nil {
		t.Fatalf("could not stop reply subscription\n", err)
	}
}
