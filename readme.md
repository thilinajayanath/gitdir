# gitdir

A cli tool to download a directory from a git repo.

- Works on Linux and macOS
- Supports built in [git credential store](https://git-scm.com/docs/git-credential-store)

## Installing and running gitdir

Installing gitdir require [golang tool chain](https://go.dev/dl/).

There are two ways to install gitdir.
1. Install with go tool chain
2. Build from source

### Install with go tool chain
1. Install with go tool chain.
```bash
go install github.com/thilinajayanath/gitdir/cmd/gitdir@latest
```

2. Run the tool with a config file. An example config is shown below.

```bash
<go-path>/bin/gitdir -config config.yaml
```

### Build from source
1. Clone the repo and switch to the repo.

```bash
git clone https://github.com/thilinajayanath/gitdir.git
cd gitdir
```

2. Build the tool with go.

```bash
make build
```

3. Run the tool with a config file. An example config is shown below.

```bash
./bin/gitdir -config config.yaml
```

## Config file

Configuration file is a yaml structure containing the following information.

- One or more git repo(s) from where the files are copied from under the `repos` key.
- Each git repo has the following configurations.
  - URL of the git repo (SSH or HTTPS).
  - One of the following authentication methods.
    - `none` - If no authentication is required to access the gir repo (For example, for public git repos).
    - `ssh` - For accessing a git repo using the ssh method. This require the ssh key path to be specified.
    - `credential-store` - For accessing a git repo using an access token. Credentials are retrieved from the default locations specified [here](https://git-scm.com/docs/git-credential-store)
  - List of source directory of the repo, local destination path and the git revision.

An example configuration file is shown below.

### Example Configuration File

```yaml
repos:
  - url: git@github.com:example-user/example.git
    auth:
      type: ssh
      credentials:
        key: /home/example/.ssh/id_rsa
    directories:
      - source: /
        target: /tmp/example
        revision: aaaabbbbccccddddeeeeffffgggghhhhiiiikkkk
  - url: https://github.com/example-user/example.git
    auth:
      type: none
    directories:
      - source: /
        target: /tmp/example
        revision: aaaabbbbccccddddeeeeffffgggghhhhiiiikkkk
  - url: https://github.com/example-user/private-repo.git
    auth:
      type: credential-store
    directories:
      - source: /
        target: /tmp/example
        revision: aaaabbbbccccddddeeeeffffgggghhhhiiiikkkk
      - source: /test/data
        target: /tmp/data
        revision: aaaabbbbccccddddeeeeyyyygggghhhhiiiikkkk
```
