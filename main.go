// dockerize.go
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
)

func buildDockerImage(ctx context.Context, cli *client.Client, imageName, contextPath string) error {
	err := os.Chdir(contextPath)
	if err != nil {
		panic(err)
	}

	buildOptions := types.ImageBuildOptions{
		Tags:           []string{imageName},
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

	buildResponse, err := cli.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return err
	}
	defer buildResponse.Body.Close()

	// Print build output
	_, err = io.Copy(os.Stdout, buildResponse.Body)
	if err != nil {
		return err
	}

	return nil
}

func createAndStartContainer(ctx context.Context, cli *client.Client, imageName string) (string, error) {

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

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")

	if err != nil {
		return "", err
	}

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func main() {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	imageName := "your-serverless-app:latest"
	contextPath := "./app" // path to your GoLang application

	err = buildDockerImage(ctx, cli, imageName, contextPath)
	if err != nil {
		fmt.Printf("Failed to build Docker image: %v\n", err)
		return
	}

	createAndStartContainer(ctx, cli, imageName)

	fmt.Printf("Docker image '%s' built successfully!\n", imageName)
}
