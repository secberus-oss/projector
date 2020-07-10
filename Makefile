GOPATH:=$(shell go env GOPATH)

build:
	go build -o projector *.go

test:
	go test -v ./... -cover

docker:
	docker build -t feelobot/projector:${IMAGE_VERSION} .

admin:org_hook
admin:repo_hook
