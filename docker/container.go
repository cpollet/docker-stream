package docker

import (
	"os"
	"github.com/cpollet/docker-stream/context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/api/types/network"
)

type Container struct {
	ID      string
	context *context.Context
}

func CreateContainer(ctx *context.Context, name string,
	containerConfig *container.Config, hostConfig *container.HostConfig, networkingConfix *network.NetworkingConfig) Container {

	containerCreateResponse, err := ctx.DockerClient.ContainerCreate(ctx.APIContext, containerConfig, hostConfig, nil, name)
	if err != nil {
		panic(err)
	}

	return Container{
		ID:      containerCreateResponse.ID,
		context: ctx,
	}
}

func (c *Container) StartSync() int {
	closeStreamFunc := attach(c.context, c.ID)
	defer closeStreamFunc()

	if err := c.context.DockerClient.ContainerStart(c.context.APIContext, c.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return syncWaitExit(c.context, c.ID)
}

func (c *Container) Destroy() {
	c.context.DockerClient.ContainerRemove(c.context.APIContext, c.ID, types.ContainerRemoveOptions{})
}

func attach(ctx *context.Context, id string) func() {
	attachOptions := types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	}

	attachResponse, err := ctx.DockerClient.ContainerAttach(ctx.APIContext, id, attachOptions)

	if err != nil {
		panic(err)
	}

	go func() {
		_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, attachResponse.Reader)
		if err != nil {
			panic(err)
		}
	}()

	return attachResponse.Close
}

func syncWaitExit(ctx *context.Context, containerId string) int {
	return <-waitExit(ctx, containerId)
}

func waitExit(ctx *context.Context, containerId string) chan int {
	statusChan := make(chan int)

	resultChan, errChan := ctx.DockerClient.ContainerWait(ctx.APIContext, containerId, container.WaitConditionNextExit)
	go func() {
		select {
		case err := <-errChan:
			if err != nil {
				panic(err)
			}
		case result := <-resultChan:
			statusChan <- int(result.StatusCode)
		}
	}()

	return statusChan
}
