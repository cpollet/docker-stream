# docker-stream

The goal of docker-stream is to be able to chain containers just like one chain linux/unix commands using pipes
(like `ps -aux | grep -v grep | whatever`). Each container's output becomes the next one's input. At the moment, the
stream is described in a yaml file. You can refer to [docker-stream.yml](https://github.com/cpollet/docker-stream/blob/master/docker-stream.yml)
as an example of such a file.

The idea behind it is to be able to build complex processes while keeping each docker images involved as simple as
possible. It is an attempt to apply the [single responsibility principle](https://en.m.wikipedia.org/wiki/Single_responsibility_principle)
to docker images.

This project is a POC to validate this idea.

## Run
```
go get ./...
go run -ldflags "-X main.gitHash=$(git rev-parse HEAD)" docker-stream.go
go install -ldflags "-X main.gitHash=$(git rev-parse HEAD)"
```
