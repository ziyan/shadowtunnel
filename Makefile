.PHONY: all
all: build

.PHONY: build
build: shadowtunnel

shadowtunnel: $(shell find . -iname '*.go' -print)
	CGO_ENABLED=0 go build github.com/ziyan/shadowtunnel
	objcopy --strip-all shadowtunnel

.PHONY: test
test: build
	go test ./...

.PHONY: docker
docker: test
	docker build -t ziyan/shadowtunnel .

.PHONY: format
format:
	gofmt -l -w client server config secure compress cli
