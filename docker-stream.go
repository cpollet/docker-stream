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
	"regexp"
	"io"
	"sync"
	"github.com/fatih/color"
	"strings"
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
	err, config := readConfig(os.Args[1])

	if err != nil {
		panic(err)
	}

	fmt.Printf("Starting stream %#v\n", config.Name)

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	fgBlue := color.New(color.FgBlue).SprintfFunc()

	for _, step := range config.Steps {
		wg.Add(1)

		containerConfig := &container.Config{
			Image:        step.Image,
			Cmd:          step.Command,
			Env:          step.Environment,
			AttachStdout: false,
		}
		var hostConfig *container.HostConfig = nil
		var networkConfig *network.NetworkingConfig = nil
		containerName := reg.ReplaceAllString(config.Name, "-") + "_" + reg.ReplaceAllString(step.Name, "-")
		stdoutContainerName := fgBlue("%s%s|", containerName, strings.Repeat(" ", 20-len(containerName)))

		fmt.Printf("%s create\n", stdoutContainerName)
		containerCreateResponse, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, containerName)
		if err != nil {
			panic(err)
		}

		// TODO attach container

		fmt.Printf("%s start\n", stdoutContainerName)
		if err := cli.ContainerStart(ctx, containerCreateResponse.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}

		status := syncWaitExit(cli, ctx, containerCreateResponse)
		fmt.Printf("%s exited with status %#v\n", stdoutContainerName, status)
		wg.Done()

		out, err := cli.ContainerLogs(ctx, containerCreateResponse.ID, types.ContainerLogsOptions{ShowStdout: true})
		if err != nil {
			panic(err)
		}

		io.Copy(os.Stdout, out)

		// TODO delete container
	}

	wg.Wait()
}
func readConfig(filename string) (error, *Config) {
	source, err := ioutil.ReadFile(filename)

	if err != nil {
		return err, nil
	}

	var config Config
	err = yaml.Unmarshal(source, &config)
	return err, &config
}

func syncWaitExit(cli *client.Client, ctx context.Context, containerCreateResponse container.ContainerCreateCreatedBody) int {
	return <-waitExit(cli, ctx, containerCreateResponse)
}

func waitExit(cli *client.Client, ctx context.Context, containerCreateResponse container.ContainerCreateCreatedBody) chan int {
	statusChan := make(chan int)

	resultChan, errChan := cli.ContainerWait(ctx, containerCreateResponse.ID, container.WaitConditionNextExit)
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
