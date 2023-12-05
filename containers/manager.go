package containers

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
)

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

type ContainerManager struct {
	cli *client.Client
}

func NewManager() (*ContainerManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return &ContainerManager{cli}, err
}

func (cm *ContainerManager) BuildImage(ctx context.Context, contextPath string, projectName string) error {
	currentWD, err := os.Getwd()

	err = os.Chdir(contextPath)
	if err != nil {
		return err
	}

	buildOptions := types.ImageBuildOptions{
		Tags:           []string{projectName},
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
		Dockerfile:     "dockerfile",
	}

	buildContext, err := archive.TarWithOptions(".", &archive.TarOptions{})
	if err != nil {
		return err
	}

	buildResponse, err := cm.cli.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return err
	}

	defer buildResponse.Body.Close()
	defer os.Chdir(currentWD)

	// Print build output
	return checkError(buildResponse.Body)
}

func (cm *ContainerManager) CreateContainer(ctx context.Context, imageName string) (string, error) {
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"3000/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "3000",
				},
			},
		},
	}

	containerConfig := &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			"3000/tcp": struct{}{},
		},
	}

	resp, err := cm.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")

	return resp.ID, err
}

func (cm *ContainerManager) StartContainer(ctx context.Context, containerId string) error {
	err := cm.cli.ContainerStart(ctx, containerId, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	return err
}

func (cm *ContainerManager) StopContainer(ctx context.Context, containerId string) error {
	err := cm.cli.ContainerStop(ctx, containerId, container.StopOptions{})
	return err
}

func (cm *ContainerManager) DeleteContainer(ctx context.Context, containerId string) error {
	err := cm.cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{})
	return err
}

func (cm *ContainerManager) IsRunning(ctx context.Context, containerId string) bool {
	container, err := cm.cli.ContainerInspect(ctx, containerId)
	if err != nil {
		return false
	}

	return container.State.Running
}

func (cm *ContainerManager) GetStatus(ctx context.Context, containerId string) string {
	container, err := cm.cli.ContainerInspect(ctx, containerId)
	if err != nil {
		return ""
	}

	return container.State.Status
}

func checkError(reader io.Reader) error {
	var lastLine string

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lastLine = scanner.Text()
		fmt.Println(scanner.Text())
	}

	errLine := &ErrorLine{}
	err := json.Unmarshal([]byte(lastLine), errLine)
	if err != nil {
		return errors.New("unable to parse build output")
	}

	if errLine.Error != "" {
		return errors.New(errLine.Error)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
