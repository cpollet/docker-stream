package docker

import (
	"github.com/cpollet/docker-stream/context"
	"github.com/docker/docker/api/types/volume"
)

type Volume struct {
	Name    string
	context *context.Context
}

func CreateVolume(context *context.Context, name string) Volume {
	volumeCreate := volume.VolumesCreateBody{
		Driver: "local",
		Name:   name,
	}

	volumeCreateResponse, err := context.DockerClient.VolumeCreate(context.APIContext, volumeCreate)
	if err != nil {
		panic(err)
	}

	return Volume{
		Name:    volumeCreateResponse.Name,
		context: context,
	}
}

func (v *Volume) Destroy() {
	err := v.context.DockerClient.VolumeRemove(v.context.APIContext, v.Name, true)
	if err != nil {
		panic(err)
	}
}
