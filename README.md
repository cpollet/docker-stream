# docker-stream

The goal of docker-stream is to be able to chain containers just like one chain linux/unix commands using pipes (like `ps -aux | grep -v grep | whatever`). Each container's output becomes the next one's input. At the moment, the stream is described in a yaml file. You can refer to [test.yml](https://github.com/cpollet/docker-stream/blob/master/test.yml) as an example of such a file.

The idea behind it is to be able to build complex processes while keeping each docker 
images involved as simple as possible. It is an attempt to apply the
[single responsibility principle](https://en.m.wikipedia.org/wiki/Single_responsibility_principle)
to docker images.

This project is a poc to validate this idea.
