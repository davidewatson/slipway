package mirror

import (
	"context"
	"fmt"
	//"net/url"

	docker "docker.io/go-docker"
	//"docker.io/go-docker/api/types"
	//"docker.io/go-docker/api/types/filters"
)

type Client struct {
	SourceURL         string
	SourceClient      *docker.Client
	DestinationURL    string
	DestinationClient *docker.Client
}

func NewClient(SourceURL, DestinationURL string) *Client {
	return &Client{
		SourceURL:      SourceURL,
		DestinationURL: DestinationURL,
	}
}

func (Client) ListImages() error {
	cli, err := docker.NewEnvClient()
	if err != nil {
		return err
	}

	image, _, err := cli.ImageInspectWithRaw(context.Background(), "centos")
	if err != nil {
		return err
	}

	for _, tag := range image.RepoTags {
		fmt.Printf(tag)
	}

	return nil
}
