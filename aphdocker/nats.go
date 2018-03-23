package aphdocker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/nats-io/go-nats"
)

type NatsDocker struct {
	Client   *client.Client
	Image    string
	Debug    bool
	ContJSON types.ContainerJSON
}

func NewNatsDockerWithImage(image string) (*NatsDocker, error) {
	nats := &NatsDocker{}
	if len(os.Getenv("DOCKER_HOST")) == 0 {
		return nats, errors.New("DOCKER_HOST is not set")
	}
	if len(os.Getenv("DOCKER_API_VERSION")) == 0 {
		return nats, errors.New("DOCKER_API is not set")
	}
	cl, err := client.NewEnvClient()
	if err != nil {
		return nats, err
	}
	nats.Client = cl
	nats.Image = image
	return nats, nil
}

func NewNatsDocker() (*NatsDocker, error) {
	return NewNatsDockerWithImage("nats:1.0.6")
}

func (d *NatsDocker) GetNatsConnString() string {
	return fmt.Sprintf("nats://%s:%s", d.GetIP(), d.GetPort())
}

func (d *NatsDocker) RetryNatsConnection() (*nats.Conn, error) {
	nc, err := nats.Connect(d.GetNatsConnString())
	if err != nil {
		return nc, err
	}
	timeout, err := time.ParseDuration("28s")
	t1 := time.Now()
	for {
		if !nc.IsConnected() {
			if time.Now().Sub(t1).Seconds() > timeout.Seconds() {
				return nc, errors.New("timed out not connects to nats server")
			}
			continue
		}
		break
	}
	return nc, nil
}

func (d *NatsDocker) Run() (container.ContainerCreateCreatedBody, error) {
	cli := d.Client
	out, err := cli.ImagePull(context.Background(), d.Image, types.ImagePullOptions{})
	if err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}
	if d.Debug {
		io.Copy(os.Stdout, out)
	}
	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image: d.Image,
	}, nil, nil, "")
	if err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}
	if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}
	cjson, err := cli.ContainerInspect(context.Background(), resp.ID)
	if err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}
	d.ContJSON = cjson
	return resp, nil
}

func (d *NatsDocker) GetIP() string {
	return d.ContJSON.NetworkSettings.IPAddress
}

func (d *NatsDocker) GetPort() string {
	return "4222"
}

func (d *NatsDocker) Purge(resp container.ContainerCreateCreatedBody) error {
	cli := d.Client
	if err := cli.ContainerStop(context.Background(), resp.ID, nil); err != nil {
		return err
	}
	if err := cli.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}
	return nil
}
