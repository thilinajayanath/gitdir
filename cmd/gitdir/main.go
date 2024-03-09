package main

import (
	"flag"
	"log"

	"github.com/thilinajayanath/gitdir/internal/config"
	"github.com/thilinajayanath/gitdir/internal/git"
)

func parseFlags() string {
	configFile := flag.String("config", "", "Configuration file for gitdir")
	flag.Parse()
	return *configFile
}

func main() {

	configFile := parseFlags()

	config, err := config.GetConfig(configFile)
	if err != nil {
		log.Fatalf("error with loading the configuration from file: %s\n", err.Error())
	}

	git.CopyGitDir(config)
}
