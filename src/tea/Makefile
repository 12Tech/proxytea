VERSION := $(shell git describe)

.PHONY: clean build

build: clean
	GOOS=darwin GOARCH=amd64 go build  -o bin/tea.darwin -ldflags="-s -w -X 'main.queueUrl=${VERSION}'" ./main.go
	GOOS=linux GOARCH=amd64 go build  -o bin/tea.linux -ldflags="-s -w -X 'main.queueUrl=${VERSION}'" ./main.go

clean:
	rm -rf ./bin

run: build
	AWS_REGION=us-east-1 ./bin/tea.darwin \
		--queueUrl https://sqs.us-east-1.amazonaws.com/558941102362/558941102362-ProxyTeaQueue-nonlive.fifo \
		--service http://localhost:8000