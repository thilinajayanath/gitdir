# gitdir - Work in progress

A cli tool to download a directory from a git repo.

## Running the application

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

### Example

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
```
