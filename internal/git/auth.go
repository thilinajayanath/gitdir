package git

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/thilinajayanath/gitdir/internal/path"
)

type cred struct {
	username string
	password string
}

// https://git-scm.com/docs/git-credential-store

func getCredentials(domain string) ([]cred, error) {
	f, err := getCredFileLocation()
	if err != nil {
		return []cred{}, err
	}

	content := getAuthFileContent(f)
	// fileContent := getAuthFileContent(f)
	return parseAuth(content, domain), nil
}

func getCredFileLocation() (string, error) {
	userHome := os.Getenv("HOME")
	xdgConfHome := os.Getenv("XDG_CONFIG_HOME")

	if userHome == "" && xdgConfHome == "" {
		return "", errors.New("HOME and XDG_CONFIG_HOME environment variables are not set")
	}

	credFile := ""

	if userHome != "" {
		credFile = path.FulltPath(userHome, ".git-credentials")
		if _, err := os.Stat(credFile); err == nil {
			return credFile, nil
		} else if errors.Is(err, os.ErrNotExist) {
			log.Printf("git credential file %s does not exist\n", credFile)
		} else {
			log.Printf("%s file information cannot be retrieved \n", credFile)
			log.Println("error: ", err.Error())
		}
	}

	if xdgConfHome != "" {
		credFile = path.FulltPath(xdgConfHome, ".config/git/credentials")

		if _, err := os.Stat(credFile); err == nil {
			return credFile, nil
		} else if errors.Is(err, os.ErrNotExist) {
			log.Printf("git credential file %s does not exist\n", credFile)
		} else {
			log.Printf("%s file information cannot be retrieved \n", credFile)
			log.Println("error: ", err.Error())
		}
	}

	credFile = path.FulltPath(userHome, ".config/git/credentials")
	if _, err := os.Stat(credFile); err == nil {
		return credFile, nil
	} else if errors.Is(err, os.ErrNotExist) {
		log.Printf("default git credential file %s does not exist\n", credFile)
	} else {
		log.Printf("%s file information cannot be retrieved \n", credFile)
		log.Println("error: ", err.Error())
	}

	return "", errors.New("default git credential file cannot be found")
}

func getAuthFileContent(f string) []string {
	readFile, err := os.Open(f)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}

	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)

	lines := []string{}
	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}

	return lines
}

func parseAuth(arr []string, domain string) []cred {
	rgx := fmt.Sprintf(`(?:https\:\/\/)?(?P<username>[[:alnum:]-]+)\:(?P<pw>\S+)\@%s`, regexp.QuoteMeta(domain))
	r := regexp.MustCompile(rgx)

	c := []cred{}

	for _, v := range arr {
		x := r.FindSubmatch([]byte(v))
		c = append(c, cred{username: string(x[1]), password: string(x[2])})
	}

	return c
}
