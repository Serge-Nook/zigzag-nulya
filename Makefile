VERSION ?= 0.1.0
LDFLAGS := -s -w -X github.com/Serge-Nook/zigzag-nulya/internal/ui.Version=$(VERSION)
DIST := dist

.PHONY: all build test lint linux windows deb appimage dist clean

all: build

build:
	go build -ldflags '$(LDFLAGS)' -o $(DIST)/topor .

test:
	go test ./...

lint:
	gofmt -l . | tee /dev/stderr | (! read)
	go vet ./...

linux:
	CGO_ENABLED=1 go build -ldflags '$(LDFLAGS)' -o $(DIST)/topor-linux-amd64 .

# Requires the mingw-w64 cross compiler (apt install gcc-mingw-w64).
windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
		go build -ldflags '$(LDFLAGS) -H windowsgui' -o $(DIST)/T0P0R-$(VERSION)-windows-amd64.exe .

deb: linux
	VERSION=$(VERSION) packaging/build-deb.sh

appimage: linux
	VERSION=$(VERSION) packaging/build-appimage.sh

dist: deb appimage windows

clean:
	rm -rf $(DIST)
