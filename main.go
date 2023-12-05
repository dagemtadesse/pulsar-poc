// dockerize.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"pulsar/poc/builder"
	"pulsar/poc/containers"
	"pulsar/poc/services"
	"time"

	"github.com/labstack/echo/v4"
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

		projectFile, err := ctx.FormFile("project")
		if err != nil {
			return nil
		}

		installer, err := builder.Setup(projectFile)
		if err != nil {
			return nil
		}

		fmt.Println("Creating Docker image for the project")
		err = manager.BuildImage(bgCtx, installer.SrcDir, imageName)
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

	server.Any("*", echo.WrapHandler(services.NewProxy()), func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			req := ctx.Request()
			headers := req.Header
			appId := headers.Get("x-App-Id")

			containerManager, _ := containers.NewManager()
			if !containerManager.IsRunning(ctx.Request().Context(), appId) {
				err := containerManager.StartContainer(ctx.Request().Context(), appId)
				if err != nil {
					return ctx.String(http.StatusInternalServerError, fmt.Sprintf("Unable to start container %v", err))
				}
				fmt.Println("Container Started")
			}
			time.Sleep(500 * time.Millisecond)
			go func() {
				err = containerManager.StopContainer(context.Background(), appId)
				if err != nil {
					fmt.Printf("Unable to stop container %v\n", err)
				}
				fmt.Println("Container Stopped")
			}()

			time.Sleep(500 * time.Microsecond)
			return next(ctx)
		}
	})

	server.Logger.Fatal(server.Start(":3001"))
}
