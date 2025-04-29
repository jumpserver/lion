NAME=lion
BUILDDIR=build

VERSION ?= Unknown
BuildTime := $(shell date -u '+%Y-%m-%d %I:%M:%S%p')
COMMIT := $(shell git rev-parse HEAD)
GOVERSION := $(shell go version)

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

LDFLAGS=-w -s

GOLDFLAGS=-X 'main.Version=$(VERSION)'
GOLDFLAGS+=-X 'main.Buildstamp=$(BuildTime)'
GOLDFLAGS+=-X 'main.Githash=$(COMMIT)'
GOLDFLAGS+=-X 'main.Goversion=$(GOVERSION)'

GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags "$(GOLDFLAGS) ${LDFLAGS}"

UIDIR=ui

define make_artifact_full
	GOOS=$(1) GOARCH=$(2) $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$(1)-$(2)
	mkdir -p $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/$(UIDIR)/dist/
	cp $(BUILDDIR)/$(NAME)-$(1)-$(2) $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/$(NAME)
	cp README.md $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/README.md
	cp LICENSE $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/LICENSE
	cp config_example.yml $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/config_example.yml
	cp entrypoint.sh $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/entrypoint.sh
	cp supervisord.conf $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/supervisord.conf
	cp -r $(UIDIR)/dist/* $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/$(UIDIR)/dist/

	cd $(BUILDDIR) && tar -czvf $(NAME)-$(VERSION)-$(1)-$(2).tar.gz $(NAME)-$(VERSION)-$(1)-$(2)
	rm -rf $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2) $(BUILDDIR)/$(NAME)-$(1)-$(2)
endef

build:
	GOARCH=$(GOARCH) GOOS=$(GOOS) $(GOBUILD) -o $(BUILDDIR)/$(NAME) .

all: lion-ui
	$(call make_artifact_full,darwin,amd64)
	$(call make_artifact_full,darwin,arm64)
	$(call make_artifact_full,linux,amd64)
	$(call make_artifact_full,linux,arm64)
	$(call make_artifact_full,linux,mips64le)
	$(call make_artifact_full,linux,ppc64le)
	$(call make_artifact_full,linux,s390x)
	$(call make_artifact_full,linux,riscv64)
	$(call make_artifact_full,linux,loong64)

local: lion-ui
	$(call make_artifact_full,$(shell go env GOOS),$(shell go env GOARCH))

darwin-amd64: lion-ui
	$(call make_artifact_full,darwin,amd64)

darwin-arm64: lion-ui
	$(call make_artifact_full,darwin,arm64)

linux-amd64: lion-ui
	$(call make_artifact_full,linux,amd64)

linux-arm64: lion-ui
	$(call make_artifact_full,linux,arm64)

linux-loong64: lion-ui
	$(call make_artifact_full,linux,loong64)

linux-ppc64le: lion-ui
	$(call make_artifact_full,linux,ppc64le)

linux-mips64le: lion-ui
	$(call make_artifact_full,linux,mips64le)

linux-s390x: lion-ui
	$(call make_artifact_full,linux,s390x)

linux-riscv64: lion-ui
	$(call make_artifact_full,linux,riscv64)

.PHONY: docker
docker:
	docker buildx build --build-arg VERSION=$(VERSION) -t jumpserver/lion:$(VERSION) .

lion-ui:
	@echo "build ui"
	@cd $(UIDIR) && yarn install && yarn build

clean:
	rm -rf $(BUILDDIR)
	-rm -rf $(UIDIR)/dist
