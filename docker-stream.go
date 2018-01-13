package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"context"
	"regexp"
	"sync"
	"strings"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/api/types/volume"
)

type Config struct {
	Version string
	Name    string
	Steps   []Step
}

type Step struct {
	Name        string
	Image       string
	Command     []string
	Environment []string
}

func main() {
	err, config := readConfig(os.Args[1])

	if err != nil {
		panic(err)
	}

	if config.Version != "0" {
		panic(fmt.Sprintf("Invalid version: %v", config.Version))
	}

	fmt.Printf("Starting stream %#v\n", config.Name)

	ctx := context.Background()
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		panic(err)
	}

	fgBlue := color.New(color.FgBlue).SprintfFunc()

	streamName := reg.ReplaceAllString(config.Name, "-")

	volumes := createVolumes(ctx, dockerClient, config, streamName)

	var wg sync.WaitGroup
	for _, step := range config.Steps {
		containerName := streamName + "_" + reg.ReplaceAllString(step.Name, "-")
		stdoutContainerName := fgBlue("%s%s|", containerName, strings.Repeat(" ", 20-len(containerName)))

		wg.Add(1)
		runStep(ctx, dockerClient, step, stdoutContainerName, containerName)
		wg.Done()
	}
	wg.Wait()

	for _, volumeName := range volumes {
		err = dockerClient.VolumeRemove(ctx, volumeName, true)
		if err != nil {
			panic(err)
		}
	}
}
func createVolumes(ctx context.Context, dockerClient *client.Client, config *Config, streamName string) []string {
	var volumes []string

	for i := 0; i < len(config.Steps); i++ {
		volumeCreate := volume.VolumesCreateBody{
			Driver: "local",
			Name:   streamName + "_0",
		}

		volumeCreateResponse, err := dockerClient.VolumeCreate(ctx, volumeCreate)
		if err != nil {
			panic(err)
		}

		volumes = append(volumes, volumeCreateResponse.Name)
	}

	return volumes
}

func runStep(ctx context.Context, dockerClient *client.Client, step Step, stdoutContainerName string, containerName string) {
	containerConfig := &container.Config{
		Image:        step.Image,
		Cmd:          append([]string{"sh", "-c"}, step.Command...),
		Env:          step.Environment,
		AttachStdout: true,
		AttachStderr: true,
	}

	var hostConfig *container.HostConfig = nil
	var networkConfig *network.NetworkingConfig = nil

	fmt.Printf("%s create\n", stdoutContainerName)
	containerCreateResponse, err := dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, containerName)

	if err != nil {
		panic(err)
	}

	closeStreamFunc := attach(ctx, dockerClient, containerCreateResponse.ID)
	defer closeStreamFunc()

	fmt.Printf("%s start\n", stdoutContainerName)

	if err := dockerClient.ContainerStart(ctx, containerCreateResponse.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	status := syncWaitExit(dockerClient, ctx, containerCreateResponse)
	fmt.Printf("%s exited with status %#v\n", stdoutContainerName, status)

	dockerClient.ContainerRemove(ctx, containerCreateResponse.ID, types.ContainerRemoveOptions{})
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

func attach(ctx context.Context, client *client.Client, id string) func() {
	attachOptions := types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	}

	attachResponse, err := client.ContainerAttach(ctx, id, attachOptions)

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
