我是光年实验室高级招聘经理。
我在github上访问了你的开源项目，你的代码超赞。你最近有没有在看工作机会，我们在招软件开发工程师，拉钩和BOSS等招聘网站也发布了相关岗位，有公司和职位的详细信息。
我们公司在杭州，业务主要做流量增长，是很多大型互联网公司的流量顾问。公司弹性工作制，福利齐全，发展潜力大，良好的办公环境和学习氛围。
公司官网是http://www.gnlab.com,公司地址是杭州市西湖区古墩路紫金广场B座，若你感兴趣，欢迎与我联系，
电话是0571-88839161，手机号：18668131388，微信号：echo 'bGhsaGxoMTEyNAo='|base64 -D ,静待佳音。如有打扰，还请见谅，祝生活愉快工作顺利。

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
