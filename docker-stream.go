package main

import (
	"github.com/urfave/cli"
	"os"
	"github.com/cpollet/docker-stream/commands"
)

func main() {
	app := cli.NewApp()

	app.Name = "docker-stream"
	app.Usage = "chain docker containers execution as a stream"
	app.Version = "0.0.0"

	app.Commands = []cli.Command{
		{
			Name:  "run",
			Usage: "Run a stream",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "keep, k",
					Usage: "keep containers and volumes after execution",
				},
				cli.StringFlag{
					Name:  "file, f",
					Usage: "Specify an alternate stream `FILE`",
					Value: "docker-stream.yml",
				},
			},
			Action: func(c *cli.Context) error {
				workDir, err := os.Getwd()
				if err != nil {
					panic(err)
				}

				commands.Run(&commands.RunConfiguration{
					Keep:     c.Bool("keep"),
					WorkDir:  workDir,
					Filename: c.String("file"),
				})

				return nil
			},
		},
	}

	app.Run(os.Args)
}
