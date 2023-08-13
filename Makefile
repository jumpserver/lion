NAME=lion
BUILDDIR=build
VERSION ?=Unknown
BuildTime:=$(shell date -u '+%Y-%m-%d %I:%M:%S%p')
COMMIT:=$(shell git rev-parse HEAD)
GOVERSION:=$(shell go version)

LDFLAGS=-w -s

GOLDFLAGS=-X 'main.Version=$(VERSION)'
GOLDFLAGS+=-X 'main.Buildstamp=$(BuildTime)'
GOLDFLAGS+=-X 'main.Githash=$(COMMIT)'
GOLDFLAGS+=-X 'main.Goversion=$(GOVERSION)'

GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags "$(GOLDFLAGS) ${LDFLAGS}"

UIDIR=ui
NPMINSTALL=yarn
NPMBUILD=yarn build

PLATFORM_LIST = \
	linux-amd64 \
	linux-arm64 \
	linux-s390x \
	linux-ppc64le

CURRENT_OS_ARCH = $(shell go env GOOS)-$(shell go env GOARCH)

all-arch: $(PLATFORM_LIST)

build:
	$(GOBUILD) -o $(BUILDDIR)/$(NAME)-$(CURRENT_OS_ARCH)

linux-amd64:lion-ui
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@
	mkdir -p $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/dist/
	cp $(BUILDDIR)/$(NAME)-$@ $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(NAME)
	-cp config_example.yml $(BUILDDIR)/$(NAME)-$(VERSION)-$@/config_example.yml
	cp -r $(UIDIR)/dist/* $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/dist/
	cd $(BUILDDIR) && tar -czvf $(NAME)-$(VERSION)-$@.tar.gz $(NAME)-$(VERSION)-$@
	rm -rf $(BUILDDIR)/$(NAME)-$(VERSION)-$@ $(BUILDDIR)/$(NAME)-$@

linux-arm64:lion-ui
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@
	mkdir -p $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/dist/
	cp $(BUILDDIR)/$(NAME)-$@ $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(NAME)
	-cp config_example.yml $(BUILDDIR)/$(NAME)-$(VERSION)-$@/config_example.yml
	cp -r $(UIDIR)/dist/* $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/dist/
	cd $(BUILDDIR) && tar -czvf $(NAME)-$(VERSION)-$@.tar.gz $(NAME)-$(VERSION)-$@
	rm -rf $(BUILDDIR)/$(NAME)-$(VERSION)-$@ $(BUILDDIR)/$(NAME)-$@

linux-s390x:lion-ui
	GOARCH=s390x GOOS=linux $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@
	mkdir -p $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/dist/
	cp $(BUILDDIR)/$(NAME)-$@ $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(NAME)
	-cp config_example.yml $(BUILDDIR)/$(NAME)-$(VERSION)-$@/config_example.yml
	cp -r $(UIDIR)/dist/* $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/dist/
	cd $(BUILDDIR) && tar -czvf $(NAME)-$(VERSION)-$@.tar.gz $(NAME)-$(VERSION)-$@
	rm -rf $(BUILDDIR)/$(NAME)-$(VERSION)-$@ $(BUILDDIR)/$(NAME)-$@

linux-ppc64le:lion-ui
	GOARCH=ppc64le GOOS=linux $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@
	mkdir -p $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/dist/
	cp $(BUILDDIR)/$(NAME)-$@ $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(NAME)
	-cp config_example.yml $(BUILDDIR)/$(NAME)-$(VERSION)-$@/config_example.yml
	cp -r $(UIDIR)/dist/* $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/dist/
	cd $(BUILDDIR) && tar -czvf $(NAME)-$(VERSION)-$@.tar.gz $(NAME)-$(VERSION)-$@
	rm -rf $(BUILDDIR)/$(NAME)-$(VERSION)-$@ $(BUILDDIR)/$(NAME)-$@

linux-loong64:lion-ui
	GOARCH=loong64 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@
	mkdir -p $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/dist/
	cp $(BUILDDIR)/$(NAME)-$@ $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(NAME)
	-cp config_example.yml $(BUILDDIR)/$(NAME)-$(VERSION)-$@/config_example.yml
	cp -r $(UIDIR)/dist/* $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/dist/
	cd $(BUILDDIR) && tar -czvf $(NAME)-$(VERSION)-$@.tar.gz $(NAME)-$(VERSION)-$@
	rm -rf $(BUILDDIR)/$(NAME)-$(VERSION)-$@ $(BUILDDIR)/$(NAME)-$@

.PHONY: docker
docker:
	docker buildx build --build-arg VERSION=$(VERSION) -t jumpserver/lion:$(VERSION) .

lion-ui:
	@echo "build ui"
	@cd $(UIDIR) && $(NPMINSTALL) && $(NPMBUILD)

clean:
	rm -rf $(BUILDDIR)
	-rm -rf $(UIDIR)/dist
