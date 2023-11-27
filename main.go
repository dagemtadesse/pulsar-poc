// dockerize.go
package main

import (
	"context"
	"fmt"
	"pulsar/poc/containers"
)

func main() {
	ctx := context.Background()

	manager, err := containers.NewManager()
	if err != nil {
		panic(err)
	}

	imageName := "your-serverless-app:latest"
	contextPath := "./app" // path to your GoLang application

	fmt.Println("Creating Docker image for the project")
	err = manager.BuildImage(ctx, contextPath, imageName)
	if err != nil {
		fmt.Printf("Failed to build Docker image: %v\n", err)
		return
	}

	fmt.Println("Creating Docker Container for the project")
	containerId, err := manager.CreateContainer(ctx, imageName)
	if err != nil {
		fmt.Printf("Failed to create Docker container: %v\n", err)
		return
	}

	fmt.Println("Starting starting the project container")
	err = manager.StartContainer(ctx, containerId)
	if err != nil {
		fmt.Printf("Failed to start Docker container: %v\n", err)
	}

	fmt.Println("Stopping the project container")
	err = manager.StopContainer(ctx, containerId)
	if err != nil {
		fmt.Printf("Failed to start Docker container: %v\n", err)
	}

	fmt.Println("Deleting the project container")
	err = manager.DeleteContainer(ctx, containerId)
	if err != nil {
		fmt.Printf("Failed to start Docker container: %v\n", err)
	}
}
