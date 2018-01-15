package context

import (
	"context"
	"github.com/docker/docker/client"
)

type Context struct {
	Stream       string
	WorkDir      string
	DockerClient *client.Client
	APIContext   context.Context
}
