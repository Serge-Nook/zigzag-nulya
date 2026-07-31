BINARY := korob

.PHONY: all build build-arm64 test vet fmt clean

all: build

build:
	go build -o dist/$(BINARY) .

build-arm64:
	CGO_ENABLED=1 GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -o dist/$(BINARY)-arm64 .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l .

clean:
	rm -rf dist
