package main

import (
	"github.com/cpollet/docker-stream/docker"
	"fmt"
	"os"
	"regexp"
	"strings"
	apiContext "context"
	"github.com/cpollet/docker-stream/stream"
	"github.com/cpollet/docker-stream/context"
	"github.com/docker/docker/client"
	"github.com/cpollet/docker-stream/configuration"
	"path"
	"log"
)

func main() {
	executionContext := context.HandleSigint()

	workDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	configFilename := "docker-stream.yml"

	if len(os.Args) > 1 {
		configFilename = os.Args[1]
	}

	config, err := configuration.Read(configFilename)
	if err != nil {
		panic(err)
	}

	if config.Version != "0" {
		panic(fmt.Sprintf("Invalid version: %v", config.Version))
	}

	if config.Name == "" {
		config.Name = path.Base(workDir)
	}

	log.Printf("Starting stream %#v\n", config.Name)

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")

	ctx := &context.Context{
		Stream:           reg.ReplaceAllString(config.Name, "-"),
		WorkDir:          workDir,
		APIContext:       apiContext.Background(),
		DockerClient:     dockerClient,
		ExecutionContext: executionContext,
	}

	volumes := createVolumes(ctx, len(config.Steps)-1)
	steps := createSteps(ctx, config, reg)

	for _, step := range steps {
		if !ctx.ExecutionContext.WorkerStart() {
			break
		}

		log.Printf("Executing %s...\n", step.ContainerName)
		status := step.RunSync()
		log.Printf("Exit status: %d\n", status)

		ctx.ExecutionContext.WorkerStop()
	}

	ctx.ExecutionContext.Wait()
	destroySteps(steps)
	destroyVolumes(volumes)
}

func createVolumes(ctx *context.Context, count int) []docker.Volume {
	var volumes []docker.Volume

	for i := 0; i < count; i++ {
		if !ctx.ExecutionContext.WorkerStart() {
			break
		}

		volumeName := fmt.Sprintf("%s_%d", ctx.Stream, i)
		volumes = append(volumes, docker.CreateVolume(ctx, volumeName))
		log.Printf("Volume created: %s\n", volumeName)
		ctx.ExecutionContext.WorkerStop()
	}

	return volumes
}

func createSteps(ctx *context.Context, config *configuration.Config, reg *regexp.Regexp) []stream.Step {
	var steps []stream.Step

	for i, stepConfig := range config.Steps {
		if !ctx.ExecutionContext.WorkerStart() {
			break
		}

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

		log.Printf("Container created: %s\n", containerName)
		ctx.ExecutionContext.WorkerStop()
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

func destroySteps(steps []stream.Step) {
	for _, s := range steps {
		s.Destroy()
		log.Printf("Container destroyed: %s\n", s.ContainerName)
	}
}

func destroyVolumes(volumes []docker.Volume) {
	for _, v := range volumes {
		v.Destroy()
		log.Printf("Volume destroyed: %s\n", v.Name)
	}
}
