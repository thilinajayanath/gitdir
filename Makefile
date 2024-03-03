.PHONY: run
run: build
	./bin/app $(file)

.PHONY: build
build:
	go build -o ./bin/app cmd/gitdir/main.go

.PHONY: test
test:
	go test -v -race ./...
