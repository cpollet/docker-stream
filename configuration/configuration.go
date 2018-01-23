package configuration

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type config struct {
	Version string
	Name    string
	Steps   yaml.MapSlice
}

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

	var c config
	err = yaml.Unmarshal(source, &c)
	return convert(c), err
}

func convert(c config) *Config {
	config := &Config{
		Version: c.Version,
		Name:    c.Name,
	}

	convertSteps(c.Steps, config)

	return config
}

func convertSteps(stepSlice yaml.MapSlice, config *Config) {
	config.Steps = make([]Step, len(stepSlice))

	if len(stepSlice) == 0 {
		return
	}

	for stepIndex, stepRaw := range stepSlice {
		config.Steps[stepIndex] = Step{}
		step := &config.Steps[stepIndex]

		step.Name = stepRaw.Key.(string)
		for _, v := range stepRaw.Value.(yaml.MapSlice) {
			switch v.Key.(string) {
			case "image":
				step.Image = v.Value.(string)
			case "command":
				step.Command = make([]string, len(v.Value.([]interface{})))
				for i, v := range v.Value.([]interface{}) {
					step.Command[i] = v.(string)
				}
			case "environment":
				step.Environment = make([]string, len(v.Value.([]interface{})))
				for i, v := range v.Value.([]interface{}) {
					step.Environment[i] = v.(string)
				}
			case "volumes":
				step.Volumes = make([]string, len(v.Value.([]interface{})))
				for i, v := range v.Value.([]interface{}) {
					step.Volumes[i] = v.(string)
				}
			}
		}
	}
}
