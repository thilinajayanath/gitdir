package config

import (
	"fmt"
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
	Type        string            `yaml:"type"`
	Credentials map[string]string `yaml:"credentials"`
}

type Directory struct {
	Source   string `yaml:"source"`
	Target   string `yaml:"target"`
	Revision string `yaml:"revision"`
}

// GetConfig return the unmarshaled yaml configuration
func GetConfig(cf string) (Config, error) {
	content, err := os.ReadFile(cf)
	if err != nil {
		fmt.Println("Error with opening the configuraiton file")
		return Config{}, err
	}

	c := Config{}

	err = yaml.Unmarshal(content, &c)
	if err != nil {
		fmt.Println("Error with parsing the configuraiton file")
		return Config{}, err
	}

	fmt.Println("Parsed the configuration file successfully")
	return c, nil
}
