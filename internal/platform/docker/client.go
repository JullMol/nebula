package docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerClient interface {
	RunContainer(ctx context.Context, image string, cmd []string, env []string) (string, error)
	StopContainer(ctx context.Context, containerID string) error
	GetLogs(ctx context.Context, containerID string) (string, error)
}

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &Client{cli: cli}, nil
}

func (c *Client) RunContainer(ctx context.Context, image string, cmd []string, env []string) (string, error) {
	reader, err := c.cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pull image %s: %w", image, err)
	}
	io.Copy(io.Discard, reader)
	reader.Close()

	resp, err := c.cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd: cmd,
		Env: env,
		Tty: false,
	}, nil, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}
	return resp.ID, nil
}

func (c *Client) StopContainer(ctx context.Context, containerID string) error {
	timeout := 5
	stopOptions := container.StopOptions{
		Timeout: &timeout,
	}
	
	if err := c.cli.ContainerStop(ctx, containerID, stopOptions); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}
	return nil
}

func (c *Client) GetLogs(ctx context.Context, containerID string) (string, error) {
	out, err := c.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow: true,
	})
	if err != nil {
		return "", err
	}
	defer out.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, out)
	return buf.String(), err
}