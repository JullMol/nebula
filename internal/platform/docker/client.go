package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Client{cli: cli}, nil
}

func (c *Client) RunContainer(ctx context.Context, imageName string, command string, code string) (string, error) {
	reader, err := c.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return "", fmt.Errorf("gagal pull image: %w", err)
	}
	io.Copy(io.Discard, reader)
	reader.Close()

	var hostConfig *container.HostConfig
	if code != "" {
		cwd, _ := os.Getwd()
		tempDir := filepath.Join(cwd, "temp_jobs", uuid.New().String())
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return "", fmt.Errorf("gagal bikin folder temp: %w", err)
		}

		fileName := "main.txt"
		runCommand := ""

		if strings.Contains(imageName, "python") {
			fileName = "main.py"
			runCommand = "python -u"
		} else if strings.Contains(imageName, "node") {
			fileName = "main.js"
			runCommand = "node"
		}

		filePath := filepath.Join(tempDir, fileName)
		if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
			return "", fmt.Errorf("gagal tulis file: %w", err)
		}

		fmt.Printf("📂 Script (%s) dibuat di Host: %s\n", fileName, filePath)

		hostConfig = &container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:/app", tempDir),
			},
		}

		if command == "" {
			command = fmt.Sprintf("%s /app/%s", runCommand, fileName)
		}
	}
	resp, err := c.cli.ContainerCreate(ctx, 
		&container.Config{
			Image: imageName,
			Cmd:   []string{"sh", "-c", command},
			Tty:   false,
		}, 
		hostConfig,
		nil, nil, "",
	)
	if err != nil {
		return "", fmt.Errorf("gagal create container: %w", err)
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("gagal start container: %w", err)
	}

	return resp.ID, nil
}

func (c *Client) StopContainer(ctx context.Context, containerID string) error {
	return c.cli.ContainerStop(ctx, containerID, container.StopOptions{})
}

func (c *Client) WaitContainer(ctx context.Context, containerID string) error {
	statusCh, errCh := c.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return err
	case <-statusCh:
		return nil
	}
}

func (c *Client) GetLogs(ctx context.Context, containerID string) (string, error) {
	out, err := c.cli.ContainerLogs(ctx, containerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", err
	}
	defer out.Close()

	logs, err := io.ReadAll(out)
	if err != nil {
		return "", err
	}
	return string(logs), nil
}