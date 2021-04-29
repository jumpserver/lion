NAME=lion
BUILDDIR=build
GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags
VERSION ?=Unknown
BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S%p')
COMMIT=$(shell git rev-parse HEAD)
GOVERSION=$(shell go version)
GOLDFLAGS=-X 'main.Version=$(VERSION)'
GOLDFLAGS+=-X 'main.Buildstamp=$(BuildTime)'
GOLDFLAGS+=-X 'main.Githash=$(COMMIT)'
GOLDFLAGS+=-X 'main.Goversion=$(GOVERSION)'

GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags "$(GOLDFLAGS)"

PLATFORM_LIST = \
	darwin-amd64 \
	linux-amd64 \
	linux-arm64

WINDOWS_ARCH_LIST = \
	windows-amd64

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@.exe


all-arch: $(PLATFORM_LIST) $(WINDOWS_ARCH_LIST)

clean:
	rm -rf $(BUILDDIR)