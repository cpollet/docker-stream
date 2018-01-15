package main

import (
	"github.com/cpollet/docker-stream/docker"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"regexp"
	"sync"
	"strings"
	apiContext "context"
	"github.com/cpollet/docker-stream/stream"
	"github.com/cpollet/docker-stream/context"
	"github.com/docker/docker/client"
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

func main() {
	workDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	config, err := readConfig(os.Args[1])
	if err != nil {
		panic(err)
	}

	if config.Version != "0" {
		panic(fmt.Sprintf("Invalid version: %v", config.Version))
	}

	fmt.Printf("Starting stream %#v\n", config.Name)

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")

	ctx := &context.Context{
		Stream:       reg.ReplaceAllString(config.Name, "-"),
		WorkDir:      workDir,
		APIContext:   apiContext.Background(),
		DockerClient: dockerClient,
	}

	volumes := createVolumes(ctx, len(config.Steps)-1)

	steps := createSteps(ctx, config, reg)

	var wg sync.WaitGroup
	for _, step := range steps {
		wg.Add(1)

		fmt.Printf("\n--\n-- Starting %s...\n--\n", step.ContainerName)
		status := step.RunSync()
		fmt.Printf("-- Status: %d\n", status)

		wg.Done()
	}
	wg.Wait()

	for _, s := range steps {
		s.Destroy()
		fmt.Printf("Container destroyed: %s\n", s.ContainerName)
	}
	for _, v := range volumes {
		v.Destroy()
		fmt.Printf("Volume destroyed: %s\n", v.Name)
	}
}

func createSteps(ctx *context.Context, config *configuration.Config, reg *regexp.Regexp) []stream.Step {
	var steps []stream.Step
	for i, stepConfig := range config.Steps {
		containerName := ctx.Stream + "_" + reg.ReplaceAllString(stepConfig.Name, "-")

		step := stream.CreateStep(
			ctx,
			stream.StepConfiguration{
				Image:         stepConfig.Image,
				ContainerName: containerName,
				Command:       stepConfig.Command,
				Environment:   stepConfig.Environment,
				Volumes:       parseMounts(ctx, stepConfig.Volumes),
				Index:         i,
				First:         i == 0,
				Last:          i == len(config.Steps)-1,
			},
		)
		steps = append(steps, step)

		fmt.Printf("Container created: %s\n", containerName)
	}
	return steps
}

func parseMounts(ctx *context.Context, mounts []string) map[string]string {
	mountsMap := make(map[string]string)

	for _, mount := range mounts {
		paths := strings.Split(mount, ":")
		mountsMap[resolveWorkDir(paths[0], ctx.WorkDir)] = paths[1]
	}

	return mountsMap
}

func resolveWorkDir(path string, workDir string) string {
	if !strings.HasPrefix(path, ".") {
		return path
	}

	return strings.Replace(path, ".", workDir, 1)
}

func createVolumes(ctx *context.Context, count int) []docker.Volume {
	var volumes []docker.Volume

	for i := 0; i < count; i++ {
		volumeName := fmt.Sprintf("%s_%d", ctx.Stream, i)
		volumes = append(volumes, docker.CreateVolume(ctx, volumeName))
		fmt.Printf("Volume created: %s\n", volumeName)
	}

	return volumes
}

func readConfig(filename string) (*Config, error) {
	source, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(source, &config)
	return &config, err
}
