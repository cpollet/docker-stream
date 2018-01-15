package main

import (
	"github.com/cpollet/docker-stream/math"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"context"
	"regexp"
	"sync"
	"strings"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/fatih/color"
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
	Volumes     []string
}

type RunContext struct {
	StepIndex     int
	Step          Step
	First         bool
	Last          bool
	StreamName    string
	ContainerName ContainerName
	Volumes       map[string]string
}

type ContainerName struct {
	Formatted string
	Raw       string
}

func main() {
	workDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

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
	for i, step := range config.Steps {
		containerName := streamName + "_" + reg.ReplaceAllString(step.Name, "-")
		stdoutContainerName := fgBlue("%s%s|", containerName, strings.Repeat(" ", math.Max(1, math.Abs(20-len(containerName)))))

		runContext := &RunContext{
			StepIndex:  i,
			Step:       step,
			First:      i == 0,
			Last:       i == len(config.Steps)-1,
			StreamName: streamName,
			ContainerName: ContainerName{
				Formatted: stdoutContainerName,
				Raw:       containerName,
			},
			Volumes: parseMounts(step.Volumes, workDir),
		}

		wg.Add(1)
		runStep(ctx, runContext, dockerClient)
		wg.Done()
	}
	wg.Wait()

	for _, v := range volumes {
		v.Close()
	}
}

func parseMounts(mounts []string, workDir string) map[string]string {
	mountsMap := make(map[string]string)

	for _, mount := range mounts {
		paths := strings.Split(mount, ":")
		mountsMap[resolveWorkDir(paths[0], workDir)] = paths[1]
	}

	return mountsMap
}

func resolveWorkDir(path string, workDir string) string {
	if !strings.HasPrefix(path, ".") {
		return path
	}

	return strings.Replace(path, ".", workDir, 1)
}

type Volume struct {
	Name          string
	DockerClient  *client.Client
	DockerContext context.Context
}

func (v *Volume) Close() {
	err := v.DockerClient.VolumeRemove(v.DockerContext, v.Name, true)
	if err != nil {
		panic(err)
	}
}

func createVolumes(ctx context.Context, dockerClient *client.Client, config *Config, streamName string) []Volume {
	var volumes []Volume

	for i := 0; i < len(config.Steps)-1; i++ {
		volumeCreate := volume.VolumesCreateBody{
			Driver: "local",
			Name:   fmt.Sprintf("%s_%d", streamName, i),
		}

		volumeCreateResponse, err := dockerClient.VolumeCreate(ctx, volumeCreate)
		if err != nil {
			panic(err)
		}

		volumes = append(volumes, Volume{
			Name:          volumeCreateResponse.Name,
			DockerClient:  dockerClient,
			DockerContext: ctx,
		})
	}

	return volumes
}

func runStep(ctx context.Context, runContext *RunContext, dockerClient *client.Client) {
	containerConfig := &container.Config{
		Image:        runContext.Step.Image,
		Cmd:          append([]string{"sh", "-c"}, runContext.Step.Command...),
		Env:          runContext.Step.Environment,
		AttachStdout: true,
		AttachStderr: true,
		Volumes:      map[string]struct{}{},
	}

	hostConfig := &container.HostConfig{}

	if !runContext.First {
		containerConfig.Volumes["/stream_in"] = struct{}{}
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("stream_%d:/stream_in", runContext.StepIndex-1))
	}
	if !runContext.Last {
		containerConfig.Volumes["/stream_out"] = struct{}{}
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("stream_%d:/stream_out", runContext.StepIndex))
	}
	for hostPath, containerPath := range runContext.Volumes {
		containerConfig.Volumes[containerPath] = struct{}{}
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s", hostPath, containerPath))
	}

	fmt.Printf("%s create\n", runContext.ContainerName.Formatted)
	containerCreateResponse, err := dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, runContext.ContainerName.Raw)

	if err != nil {
		panic(err)
	}

	closeStreamFunc := attach(ctx, dockerClient, containerCreateResponse.ID)
	defer closeStreamFunc()

	fmt.Printf("%s start\n", runContext.ContainerName.Formatted)

	if err := dockerClient.ContainerStart(ctx, containerCreateResponse.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	status := syncWaitExit(dockerClient, ctx, containerCreateResponse)
	fmt.Printf("%s exited with status %#v\n", runContext.ContainerName.Formatted, status)

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
