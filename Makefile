.PHONY: run
run: build
	./bin/gitdir $(file)

.PHONY: build
build:
	go build -o ./bin/gitdir cmd/gitdir/main.go

.PHONY: test
test:
	go test -v -race ./...
