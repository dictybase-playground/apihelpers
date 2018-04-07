package aphdocker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/vst"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const (
	charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var seedRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	var b []byte
	for i := 0; i < length; i++ {
		b = append(
			b,
			charset[seedRand.Intn(len(charset))],
		)
	}
	return string(b)
}
func RandString(length int) string {
	return stringWithCharset(length, charSet)
}

type ArangoDocker struct {
	Client   *client.Client
	Image    string
	Debug    bool
	ContJSON types.ContainerJSON
	user     string
	password string
}

func NewArangoDockerWithImage(image string) (*ArangoDocker, error) {
	ag := &ArangoDocker{}
	if len(os.Getenv("DOCKER_HOST")) == 0 {
		return ag, errors.New("DOCKER_HOST is not set")
	}
	if len(os.Getenv("DOCKER_API_VERSION")) == 0 {
		return ag, errors.New("DOCKER_API is not set")
	}
	cl, err := client.NewEnvClient()
	if err != nil {
		return ag, err
	}
	ag.Client = cl
	ag.Image = image
	ag.user = "root"
	ag.password = RandString(10)
	return ag, nil
}

func NewArangoDocker() (*ArangoDocker, error) {
	return NewArangoDockerWithImage("arangodb:3.3.5")
}

func (d *ArangoDocker) Run() (container.ContainerCreateCreatedBody, error) {
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
		Env: []string{
			"ARANGO_ROOT_PASSWORD=" + d.password,
			"ARANGO_STORAGE_ENGINE=rocksdb",
		},
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

func (d *ArangoDocker) GetUser() string {
	return d.user
}

func (d *ArangoDocker) GetPassword() string {
	return d.password
}

func (d *ArangoDocker) GetIP() string {
	return d.ContJSON.NetworkSettings.IPAddress
}

func (d *ArangoDocker) GetPort() string {
	return "8529"
}

func (d *ArangoDocker) Purge(resp container.ContainerCreateCreatedBody) error {
	cli := d.Client
	if err := cli.ContainerStop(context.Background(), resp.ID, nil); err != nil {
		return err
	}
	if err := cli.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}
	return nil
}

func (d *ArangoDocker) RetryConnection() (driver.Client, error) {
	conn, err := vst.NewConnection(
		vst.ConnectionConfig{
			Endpoints: []string{
				fmt.Sprintf("vst://%s:%s", d.GetIP(), d.GetPort()),
			},
		})
	if err != nil {
		log.Fatalf("cannot connect to arangodb server %s", err)
	}
	client, err := driver.NewClient(
		driver.ClientConfig{
			Connection: conn,
			Authentication: driver.BasicAuthentication(
				d.GetUser(),
				d.GetPassword(),
			),
		})
	if err != nil {
		return client, fmt.Errorf("could not get a client %s\n", err)
	}
	timeout, err := time.ParseDuration("15s")
	t1 := time.Now()
	for {
		_, err := client.Version(context.Background())
		if err != nil {
			if time.Now().Sub(t1).Seconds() > timeout.Seconds() {
				log.Fatalf("client error %s\n", err)
			}
			continue
		}
		break
	}
	return client, nil
}
