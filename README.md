# docker-stream

The goal of docker-stream is to be able to chain cobtainers just like one chain linux/unix
commands using pipes (like `ps ps -aux | grep -v grep | whatever`). At the moment, the stream
is described in a yaml file. You can refer to
[test.yml](https://github.com/cpollet/docker-stream/blob/master/test.yml) as an example of
such file.

The idea behind it is to be able to build complex processes while keeping each docker 
images involved as sinole as possible. It is an attempt to apply the
[single responsibility principle](https://en.m.wikipedia.org/wiki/Single_responsibility_principle)
to docker images.

This project is a poc to validate this idea.
