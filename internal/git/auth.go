package git

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/thilinajayanath/gitdir/internal/config"
	"github.com/thilinajayanath/gitdir/internal/path"
)

type cred struct {
	username string
	password string
}

// If git is configured to store credentials, there are several places where
// they can be stored as mentioned in https://git-scm.com/docs/git-credential-store
func getCredentials(domain string) ([]cred, error) {
	f, err := getCredFileLocation()
	if err != nil {
		return []cred{}, err
	}

	content := getAuthFileContent(f)
	// fileContent := getAuthFileContent(f)
	return parseAuth(content, domain)
}

// getCredFileLocation checks the default locations for the stored git
// credential file and returns the first file location that exist.
// Returns an error if none of the default credential files exist.
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
			fmt.Printf("Git credential file %s does not exist\n", credFile)
		} else {
			fmt.Printf("%s file information cannot be retrieved\n", credFile)
			fmt.Println("Error: ", err.Error())
		}
	}

	if xdgConfHome != "" {
		credFile = path.FulltPath(xdgConfHome, ".config/git/credentials")

		if _, err := os.Stat(credFile); err == nil {
			return credFile, nil
		} else if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Git credential file %s does not exist\n", credFile)
		} else {
			fmt.Printf("%s file information cannot be retrieved\n", credFile)
			fmt.Println("Error: ", err.Error())
		}
	}

	credFile = path.FulltPath(userHome, ".config/git/credentials")
	if _, err := os.Stat(credFile); err == nil {
		return credFile, nil
	} else if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Default git credential file %s does not exist\n", credFile)
	} else {
		fmt.Printf("%s file information cannot be retrieved\n", credFile)
		fmt.Println("Error: ", err.Error())
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

func parseAuth(arr []string, domain string) ([]cred, error) {
	c := []cred{}
	rgx := fmt.Sprintf(`(?:https\:\/\/)?(?P<username>[[:alnum:]-]+)\:(?P<pw>\S+)\@%s`, regexp.QuoteMeta(domain))
	r, err := regexp.Compile(rgx)
	if err != nil {
		return c, err
	}

	for _, v := range arr {
		x := r.FindSubmatch([]byte(v))
		if len(x) == 3 {
			c = append(c, cred{username: string(x[1]), password: string(x[2])})
		}
	}

	return c, nil
}

// setupAuth creates the authentication parameters for git from the given user
// configuration
func setupAuth(auth config.Auth, domain string) (transport.AuthMethod, error) {
	switch auth.Type {
	case "ssh":
		authMethod, err := ssh.NewPublicKeysFromFile("git", auth.Credentials["key"], "")
		if err != nil {
			return nil, err
		}

		return authMethod, nil
	case "credential-store":
		cred, err := getCredentials(domain)
		if err != nil {
			return nil, err
		}

		autheMethod := http.BasicAuth{
			Username: cred[0].username,
			Password: cred[0].password,
		}

		return &autheMethod, nil
	default:
		return nil, errors.New("authentication method not found")
	}
}
