package config

import (
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Repos []Repo `yaml:"repos"`
}

type Repo struct {
	Auth        Auth        `yaml:"auth"`
	Directories []Directory `yaml:"directories"`
	URL         string      `yaml:"url"`
}

type Auth struct {
	Credentials map[string]string `yaml:"credentials"`
	Type        string            `yaml:"type"`
}

type Directory struct {
	Source   string `yaml:"source"`
	Target   string `yaml:"target"`
	Revision string `yaml:"revision"`
}

func GetConfig(cf string) (Config, error) {
	content, err := os.ReadFile(cf)
	if err != nil {
		slog.Error("error with opening the configuraiton file")
		return Config{}, err

	}

	c := Config{}

	err = yaml.Unmarshal(content, &c)
	if err != nil {
		slog.Error("error with parsing the configuraiton file")
		return Config{}, err
	}

	slog.Info("parsed the configuration file successfully")
	return c, nil
}
