package aphdocker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	_ "github.com/jackc/pgx/stdlib"
	"gopkg.in/src-d/go-git.v4"
)

type PgDocker struct {
	Client   *client.Client
	Image    string
	Pass     string
	User     string
	Database string
	Debug    bool
	ContJSON types.ContainerJSON
}

func NewPgDockerWithImage(image string) (*PgDocker, error) {
	pg := &PgDocker{}
	if len(os.Getenv("DOCKER_HOST")) == 0 {
		return pg, errors.New("DOCKER_HOST is not set")
	}
	if len(os.Getenv("DOCKER_API_VERSION")) == 0 {
		return pg, errors.New("DOCKER_API is not set")
	}
	cl, err := client.NewEnvClient()
	if err != nil {
		return pg, err
	}
	pg.Client = cl
	pg.Image = image
	pg.Pass = "pgdocker"
	pg.User = "pguser"
	pg.Database = "pgtest"
	return pg, nil
}

func NewPgDocker() (*PgDocker, error) {
	return NewPgDockerWithImage("postgres:9.6.6-alpine")
}

func (d *PgDocker) Run() (container.ContainerCreateCreatedBody, error) {
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
			"POSTGRES_PASSWORD=" + d.Pass,
			"POSTGRES_USER=" + d.User,
			"POSTGRES_DB=" + d.Database,
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

func (d *PgDocker) RetryDbConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		d.User, d.Pass, d.GetIP(), d.GetPort(), d.Database,
	)
	dbh, err := sql.Open("pgx", connStr)
	if err != nil {
		return dbh, err
	}
	timeout, err := time.ParseDuration("28s")
	t1 := time.Now()
	for {
		if err := dbh.Ping(); err != nil {
			if time.Now().Sub(t1).Seconds() > timeout.Seconds() {
				return dbh, errors.New("timed out, no connection retrieved")
			}
			continue
		}
		break
	}
	return dbh, nil
}

func (d *PgDocker) GetIP() string {
	return d.ContJSON.NetworkSettings.IPAddress
}

func (d *PgDocker) GetPort() string {
	return "5432"
}

func (d *PgDocker) Purge(resp container.ContainerCreateCreatedBody) error {
	cli := d.Client
	if err := cli.ContainerStop(context.Background(), resp.ID, nil); err != nil {
		return err
	}
	if err := cli.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}
	return nil
}

func CloneDbSchemaRepo(repo string) (string, error) {
	path, err := ioutil.TempDir("", "content")
	if err != nil {
		return path, err
	}
	_, err = git.PlainClone(path, false, &git.CloneOptions{URL: repo})
	return path, err
}
