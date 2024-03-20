package git

import (
	"context"
	"errors"
	"fmt"
	"io"
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
		fmt.Println("Cloning repo", repo.URL)

		r, fs, err := cloneRepo(repo.URL, repo.Auth)
		if err != nil {
			fmt.Println("Error with cloning the git repo", repo.URL)
			fmt.Println("Error:", err.Error())
			continue
		}

		w, err := r.Worktree()
		if err != nil {
			fmt.Println("Error with retriving the git worktree for the repo", repo.URL)
			fmt.Println("Error:", err.Error())
			continue
		}

		fmt.Println("Cloned repo", repo.URL)

		for _, dir := range repo.Directories {
			copyDir(dir.Target, repo.URL, dir.Revision, dir.Source, fs, w)
		}
	}
}

// cloneRepo clones a git repo and returns a billy.Filesystem of that represents
// all the files of the git repo
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

// getDomain retrieves the domain of the git repo URL. Returns error if the
// URL is not valid
func getDomain(repo string) (string, error) {
	if strings.HasPrefix(repo, "git") {
		return strings.Split(strings.Split(repo, "git@")[1], ":")[0], nil
	} else if strings.HasPrefix(repo, "https") {
		return strings.Split(strings.Split(repo, "https://")[1], "/")[0], nil
	}

	return "", errors.New("git repo url is invalid")
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

func copyDir(dst, repo, rev, src string, fs billy.Filesystem, wt *git.Worktree) {
	err := wt.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(rev),
	})
	if err != nil {
		fmt.Printf("Error with checking out the %s commit in %s repo\n", rev, repo)
		fmt.Println("Error:", err.Error())
		return
	}

	srcFs, err := fs.Chroot(src)
	if err != nil {
		fmt.Printf("Error with changing to source directory %s in %s commit in %s repo\n", src, rev, repo)
		fmt.Println("Error:", err.Error())
		return
	}

	fmt.Println("Copying files from", src, "in", repo)

	dirTreeChan := make(chan fileInfo)

	ctxCreateFs, cancelCreateFs := context.WithCancelCause(context.Background())
	ctxWalkFS, cancelWalkFS := context.WithCancel(context.Background())

	go walkFS(ctxWalkFS, cancelCreateFs, srcFs, "/", dirTreeChan)

	err = createFS(ctxCreateFs, cancelWalkFS, dst, fs, dirTreeChan, src)
	if err != nil {
		_, ok := <-dirTreeChan
		if ok {
			fmt.Println("closing channel")
			close(dirTreeChan)
		}

		fmt.Println("Error with copying the files")
		fmt.Println("Error:", err.Error())
		return
	}

	fmt.Println("Copied files")
}

// walkFS goes through all the directories in a given file system and sends the
// full path from the root to the given channel.
func walkFS(ctx context.Context, cancel context.CancelCauseFunc, srcFs billy.Filesystem, parent string, fileName chan<- fileInfo) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			files, err := srcFs.ReadDir("/")
			if err != nil {
				cancel(err)
				return
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
						cancel(err)
						return
					}

					walkFS(ctx, cancel, newSrcFs, filePath, fileName)
				}
			}

			if parent == "/" {
				close(fileName)
			}
			return
		}
	}
}

// createFS retrieve directories and file names from the channel given channel
// and recreates a folder structure
func createFS(ctx context.Context, cancel context.CancelFunc, dst string, fs billy.Filesystem, fileName <-chan fileInfo, src string) error {
	err := os.RemoveAll(dst)
	if err != nil {
		cancel()
		return err
	}

	err = os.MkdirAll(dst, 0755)
	if err != nil {
		cancel()
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			x, ok := <-fileName

			if !ok {
				return nil
			}

			filePath := path.FulltPath(dst, x.path)
			if x.isDir {
				err = os.Mkdir(filePath, 0755)
				if err != nil {
					fmt.Println("Dir", filePath, "creation falied")
					cancel()
					return err
				}
			} else {
				srcFilePath := path.FulltPath(src, x.path)
				srcFile, err := fs.Open(srcFilePath)

				if err != nil {
					fmt.Println("Unable to open file", srcFilePath)
					cancel()
					return err
				}
				defer srcFile.Close()

				dstFile, err := os.Create(filePath)
				if err != nil {
					fmt.Println("Unable to create the file", dstFile)
					cancel()
					return err
				}
				defer dstFile.Close()

				_, err = io.Copy(dstFile, srcFile)
				if err != nil {
					fmt.Printf("Unable to copy file from %s to %s", srcFile.Name(), dstFile.Name())
					cancel()
					return err
				}

				err = dstFile.Sync()
				if err != nil {
					fmt.Printf("Unable to commit the %s to disk", dstFile.Name())
					cancel()
					return err
				}
			}
		}
	}
}
