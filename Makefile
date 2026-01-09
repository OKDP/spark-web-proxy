.PHONY: help
help:
	@echo "Available commands:"
	@echo "  format      Format the project code"
	@echo "  lint        Run the configured linters (before pushing)"
	@echo "  compile     Compile the project code"
	@echo "  test        Run the unit tests"
	@echo "  build       Build a binary"
	@echo "  run         Run the project locally"
	@echo "  update      Update all the project dependencies"

format: tools gofmt
lint: format golint
compile: lint gocompile
test: compile gotest
build: test gobuild
run: test gorun
rundev: gocompile gotest gobuild gorun
update: build goupdate

.PHONY: tools
tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0
	go install golang.org/x/tools/cmd/goimports@v0.40.0

.PHONY: gofmt
gofmt:
	gofmt -s -w ./internal/ ./cmd/ && goimports -w -local ./internal/ ./cmd/

.PHONY: golint
golint:
	golangci-lint run

.PHONY: gobuild
gobuild:
	mkdir -p .bin/
	CGO_ENABLED=0 go build -a -o .bin/spark-web-proxy main.go

.PHONY: gocompile
gocompile:
	CGO_ENABLED=0 go build -a -o /dev/null main.go

.PHONY: gotest
gotest:
	go test ./... -v

.PHONY: gorun
gorun:
	go run *.go --config=.local/application-local.yaml

.PHONY: goupdate
goupdate:
	go get -u all
	go mod tidy

