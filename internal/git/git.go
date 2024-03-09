package git

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/thilinajayanath/gitdir/internal/config"
	"github.com/thilinajayanath/gitdir/internal/path"
)

type fileInfo struct {
	isDir bool
	mode  os.FileMode
	path  string
}

func CopyGitDir(c config.Config) {
	for _, repo := range c.Repos {
		log.Println("cloning repo ", repo.URL)
		r, fs, err := cloneRepo(repo.URL, repo.Auth)
		if err != nil {
			log.Println("error with cloning the git repo:", repo.URL)
			log.Println("error: ", err.Error())
			continue
		}

		w, err := r.Worktree()
		if err != nil {
			log.Println("error with retriving the git worktree for the repo: ", repo.URL)
			log.Println("error: ", err.Error())
			continue
		}
		log.Println(repo.URL, "cloned")

		for _, dir := range repo.Directories {
			cloneDir(dir.Target, repo.URL, dir.Revision, dir.Source, fs, w)
		}
	}
}

func cloneRepo(repo string, auth config.Auth) (*git.Repository, billy.Filesystem, error) {
	// Filesystem abstraction based on memory
	fs := memfs.New()
	// Git objects storer based on memory
	storer := memory.NewStorage()

	co := &git.CloneOptions{URL: repo}

	if auth.Type != "none" {
		domain, err := getDomain(repo)
		if err != nil {
			return &git.Repository{}, fs, err
		}

		authMethod, err := setupAuth(auth, domain)
		if err != nil {
			return &git.Repository{}, fs, err
		}

		co.Auth = authMethod
	}

	r, err := git.Clone(storer, fs, co)
	if err != nil {
		return &git.Repository{}, fs, err
	}

	return r, fs, nil
}

func getDomain(repo string) (string, error) {
	if strings.HasPrefix(repo, "git") {
		return strings.Split(strings.Split(repo, "git@")[0], ":")[0], nil
	} else if strings.HasPrefix(repo, "https") {
		return strings.Split(strings.Split(repo, "https://")[0], "/")[0], nil
	}

	return "", errors.New("git repo url is invalid")
}

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

func cloneDir(dst, repo, rev, src string, fs billy.Filesystem, wt *git.Worktree) {
	err := wt.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(rev),
	})
	if err != nil {
		log.Println("error with checking out the commit in repo ", repo)
		log.Println("error: ", err.Error())
		return
	}

	srcFs, err := fs.Chroot(src)
	if err != nil {
		log.Println("error with changing to source directory to ", src)
		log.Println("error: ", err.Error())
		return
	}

	log.Println("copying files")
	dirTreeChan := make(chan fileInfo)
	go walk(srcFs, "/", dirTreeChan)
	createFS(dst, fs, dirTreeChan, src)

}

func walk(srcFs billy.Filesystem, parent string, fileName chan<- fileInfo) {
	files, err := srcFs.ReadDir("/")
	if err != nil {
		log.Println(err.Error())
	}

	for _, file := range files {
		filePath := ""
		if parent != "/" {
			filePath = path.FulltPath(parent, file.Name())
		} else {
			filePath = path.FulltPath("/", file.Name())
		}

		fileName <- fileInfo{
			isDir: file.IsDir(),
			mode:  file.Mode(),
			path:  filePath,
		}

		if file.IsDir() {
			newSrcFs, err := srcFs.Chroot(file.Name())
			if err != nil {
				log.Println(err.Error())
			}

			walk(newSrcFs, filePath, fileName)
		}
	}

	if parent == "/" {
		close(fileName)
	}
}

func createFS(dst string, fs billy.Filesystem, fileName <-chan fileInfo, src string) {
	err := os.RemoveAll(dst)
	if err != nil {
		log.Println(err.Error())
	}

	os.MkdirAll(dst, 0755)
	if err != nil {
		log.Println(err.Error())
	}

	for x := range fileName {
		filePath := path.FulltPath(dst, x.path)
		if x.isDir {
			err = os.Mkdir(filePath, 0755)
			if err != nil {
				log.Println("dir ", filePath, "creation falied", err.Error())
			}
		} else {
			srcFilePath := path.FulltPath(src, x.path)
			srcFile, err := fs.Open(srcFilePath)
			if err != nil {
				log.Println("src", err.Error())
			}

			dstFile, err := os.Create(filePath)
			if err != nil {
				log.Println("dst", err.Error())
			}

			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				log.Println(err.Error())
			}

			err = dstFile.Sync()
			if err != nil {
				log.Println(err.Error())
			}

			srcFile.Close()
			dstFile.Close()
		}
	}
}
