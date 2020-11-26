VERSION=0.0.10
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"
GO111MODULE=on

all: mackerel-plugin-log-incr-rate

.PHONY: mackerel-plugin-log-incr-rate

mackerel-plugin-log-incr-rate: main.go
	go build $(LDFLAGS) -o mackerel-plugin-log-incr-rate

linux: main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o mackerel-plugin-log-incr-rate

deps:
	go get -d
	go mod tidy

deps-update:
	go get -u -d
	go mod tidy

check:
	go test ./...

clean:
	rm -rf mackerel-plugin-log-incr-rate

tag:
	git tag v${VERSION}
	git push origin v${VERSION}
	git push origin master
