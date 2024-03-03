package main

import (
	"flag"
	"log/slog"

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
		slog.Error(err.Error())
	}

	// fmt.Println(config)

	git.CopyGitDir(config)
}
