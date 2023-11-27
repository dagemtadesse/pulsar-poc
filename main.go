// dockerize.go
package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"pulsar/poc/containers"
)

func main() {
	manager, err := containers.NewManager()
	if err != nil {
		panic(err)
	}

	server := echo.New()
	bgCtx := context.Background()

	server.POST("project/:name", func(ctx echo.Context) error {
		imageName := ctx.Param("name")
		contextPath := "./app" // path to your GoLang application

		fmt.Println("Creating Docker image for the project")
		err = manager.BuildImage(bgCtx, contextPath, imageName)
		if err != nil {
			return ctx.String(http.StatusInternalServerError,
				fmt.Sprintf("Failed to build Docker image: %v\n", err))
		}

		containerId, err := manager.CreateContainer(bgCtx, imageName)
		if err != nil {
			return ctx.String(http.StatusInternalServerError,
				fmt.Sprintf("Failed to create Docker container: %v\n", err))
		}
		return ctx.String(http.StatusOK, fmt.Sprintf("Project build with conatiner id: %s", containerId))
	})

	server.POST("project/:projectId/start", func(ctx echo.Context) error {
		containerId := ctx.Param("projectId")

		err = manager.StartContainer(bgCtx, containerId)
		if err != nil {
			return ctx.String(http.StatusInternalServerError,
				fmt.Sprintf("Failed to start Docker container: %v\n", err))
		}
		return ctx.String(http.StatusOK, "Project started successfully")
	})

	server.POST("project/:projectId/stop", func(ctx echo.Context) error {
		containerId := ctx.Param("projectId")
		err = manager.StopContainer(bgCtx, containerId)
		if err != nil {
			return ctx.String(http.StatusInternalServerError,
				fmt.Sprintf("Failed to stop project container: %v", err))
		}
		return ctx.String(http.StatusOK, "Project stopped successfully")
	})

	server.DELETE("project/:projectId", func(ctx echo.Context) error {
		containerId := ctx.Param("projectId")
		err = manager.DeleteContainer(bgCtx, containerId)
		if err != nil {
			return ctx.String(http.StatusInternalServerError,
				fmt.Sprintf("Failed to remove project container: %v", err))
		}
		return ctx.String(http.StatusOK, "Project removed successfully")
	})

	server.Logger.Fatal(server.Start(":3001"))
}
