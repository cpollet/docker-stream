package stream

import (
	"fmt"
	"github.com/cpollet/docker-stream/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/cpollet/docker-stream/context"
)

type Step struct {
	ContainerName string
	context       *context.Context
	container     docker.Container
}
type StepConfiguration struct {
	Image         string
	ContainerName string
	Command       []string
	Environment   []string
	Volumes       map[string]string
	Index         int
	First         bool
	Last          bool
}

func CreateStep(ctx *context.Context, configuration StepConfiguration) Step {
	containerConfig := &container.Config{
		Image:        configuration.Image,
		Cmd:          append([]string{"sh", "-c"}, configuration.Command...),
		Env:          configuration.Environment,
		AttachStdout: true,
		AttachStderr: true,
		Volumes:      map[string]struct{}{},
	}

	hostConfig := &container.HostConfig{}

	if !configuration.First {
		containerConfig.Volumes["/stream_in"] = struct{}{}
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s_%d:/stream_in", ctx.Stream, configuration.Index-1))
	}
	if !configuration.Last {
		containerConfig.Volumes["/stream_out"] = struct{}{}
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s_%d:/stream_out", ctx.Stream, configuration.Index))
	}
	for hostPath, containerPath := range configuration.Volumes {
		containerConfig.Volumes[containerPath] = struct{}{}
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s", hostPath, containerPath))
	}

	return Step{
		ContainerName: configuration.ContainerName,
		context:       ctx,
		container:     docker.CreateContainer(ctx, configuration.ContainerName, containerConfig, hostConfig, nil),
	}
}

func (s *Step) RunSync() int {
	return s.container.StartSync()
}

func (s *Step) Destroy() {
	s.container.Destroy()
}
