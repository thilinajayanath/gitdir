# gitdir

## Config file example

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
```
