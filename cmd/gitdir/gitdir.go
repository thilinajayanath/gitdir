package gitdir

import (
	"flag"
	"fmt"
	"os"

	"github.com/thilinajayanath/gitdir/internal/config"
	"github.com/thilinajayanath/gitdir/internal/git"
)

func parseFlags() string {
	configFile := flag.String("config", "", "Configuration file for gitdir")
	flag.Parse()
	return *configFile
}

func Run() {
	configFile := parseFlags()

	config, err := config.GetConfig(configFile)
	if err != nil {
		fmt.Println("Unable to load the configuration file")
		fmt.Println("Error: ", err.Error())
		os.Exit(1)
	}

	git.CopyGitDir(config)
}
