package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"github.com/docker/docker/client"
	"io/ioutil"
	"os"
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types"
	"io"
)

type Config struct {
	Name string
	Steps []struct {
		Name        string
		Image       string
		Input       string
		Output      string
		Command     []string
		Environment []string
	}
}

func main() {
	filename := os.Args[1]

	source, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	var config Config
	err = yaml.Unmarshal(source, &config)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Starting stream %#v\n", config.Name)

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	for _, step := range config.Steps {
		containerConfig := &container.Config{
			Image: step.Image,
			Cmd:   step.Command,
			Env:   step.Environment,
		}
		var hostConfig *container.HostConfig = nil
		var networkConfig *network.NetworkingConfig = nil
		containerName := ""

		resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, containerName)

		if err != nil {
			panic(err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}

		statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				panic(err)
			}
		case <-statusCh:
		}

		out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
		if err != nil {
			panic(err)
		}

		io.Copy(os.Stdout, out)

		// TODO delete container
	}
}
