.PHONY: all
all: build

.PHONY: build
build: shadowtunnel

shadowtunnel: $(shell find . -iname '*.go' -print)
	godep go build github.com/ziyan/shadowtunnel/cmd/shadowtunnel

.PHONY: test
test: build
	godep go test ./...

.PHONY: docker
docker: test
	docker build -t ziyan/shadowtunnel .

.PHONY: save
save:
	godep save ./...

.PHONY: format
format:
	gofmt -l -w cmd pkg

.PHONY: doc
doc:
	@godoc -http=:6060 -index=true

