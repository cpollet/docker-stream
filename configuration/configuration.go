package configuration

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
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

func Read(filename string) (*Config, error) {
	source, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	var c Config
	err = yaml.Unmarshal(source, &c)
	return &c, err
}

