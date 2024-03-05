package git

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/thilinajayanath/gitdir/internal/config"
)

type fileInfo struct {
	isDir bool
	mode  os.FileMode
	path  string
}

func CopyGitDir(c config.Config) {
	for _, repo := range c.Repos {
		for _, dir := range repo.Directories {
			cloneDir(repo.Auth, dir.Target, repo.URL, dir.Revision, dir.Source)
		}
	}
}

func cloneDir(auth config.Auth, dst, repo, rev, src string) {
	log.Println("cloning repo ", repo)

	r, fs, err := cloneRepo(auth, repo)
	if err != nil {
		log.Println("error with cloning the git repo:", repo)
		log.Println("error: ", err.Error())
		return
	}

	w, err := r.Worktree()
	if err != nil {
		log.Println("error with retriving the git worktree for the repo: ", repo)
		log.Println("error: ", err.Error())
		return
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(rev),
	})
	if err != nil {
		log.Println("error with checking out the commit in repo: ", repo)
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

func cloneRepo(auth config.Auth, repo string) (*git.Repository, billy.Filesystem, error) {
	// Filesystem abstraction based on memory
	fs := memfs.New()
	// Git objects storer based on memory
	storer := memory.NewStorage()

	co := &git.CloneOptions{URL: repo}

	if auth.Type != "none" {
		authMethod, err := ssh.NewPublicKeysFromFile("git", auth.Credentials["key"], "")
		if err != nil {
			return &git.Repository{}, fs, err
		}
		co.Auth = authMethod
	}

	r, err := git.Clone(
		storer,
		fs,
		co,
	)
	if err != nil {
		return &git.Repository{}, fs, err
	}

	return r, fs, nil
}

func walk(srcFs billy.Filesystem, parent string, fileName chan<- fileInfo) {
	files, err := srcFs.ReadDir("/")
	if err != nil {
		log.Println(err.Error())
	}

	for _, file := range files {

		filePath := ""
		if parent != "/" {
			filePath = createFulltPath(parent, file.Name())
		} else {
			filePath = createFulltPath("/", file.Name())
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
		filePath := createFulltPath(dst, x.path)
		if x.isDir {
			err = os.Mkdir(filePath, 0755)
			if err != nil {
				log.Println("dir ", filePath, "creation falied", err.Error())
			}
		} else {
			srcFilePath := createFulltPath(src, x.path)
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

func createFulltPath(start, end string) string {
	splitPath := strings.Split(fmt.Sprintf("%s/%s", start, end), "/")

	path := []string{}

	for _, v := range splitPath {
		if v != "" {
			path = append(path, v)
		}
	}

	return fmt.Sprintf("/%s", strings.Join(path, "/"))
}
