NAME=lion
BUILDDIR=build
VERSION ?=Unknown
BuildTime:=$(shell date -u '+%Y-%m-%d %I:%M:%S%p')
COMMIT:=$(shell git rev-parse HEAD)
GOVERSION:=$(shell go version)
GOLDFLAGS=-X 'main.Version=$(VERSION)'
GOLDFLAGS+=-X 'main.Buildstamp=$(BuildTime)'
GOLDFLAGS+=-X 'main.Githash=$(COMMIT)'
GOLDFLAGS+=-X 'main.Goversion=$(GOVERSION)'

GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags "$(GOLDFLAGS)"

UIDIR=ui
NPMINSTALL=npm i
NPMBUILD=npm run-script build

PLATFORM_LIST = \
	darwin-amd64 \
	linux-amd64 \
	linux-arm64

WINDOWS_ARCH_LIST = \
	windows-amd64

all-arch: $(PLATFORM_LIST) $(WINDOWS_ARCH_LIST)

darwin-amd64:lion-ui
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@
	mkdir -p $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/lion/
	cp $(BUILDDIR)/$(NAME)-$@ $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(NAME)
	-cp config_example.yml $(BUILDDIR)/$(NAME)-$(VERSION)-$@/config_example.yml
	cp -r $(UIDIR)/lion/ $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/lion/
	cd $(BUILDDIR) && tar -czvf $(NAME)-$(VERSION)-$@.tar.gz $(NAME)-$(VERSION)-$@
	rm -rf $(BUILDDIR)/$(NAME)-$(VERSION)-$@ $(BUILDDIR)/$(NAME)-$@

linux-amd64:lion-ui
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@
	mkdir -p $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/lion/
	cp $(BUILDDIR)/$(NAME)-$@ $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(NAME)
	-cp config_example.yml $(BUILDDIR)/$(NAME)-$(VERSION)-$@/config_example.yml
	cp -r $(UIDIR)/lion/ $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/lion/
	cd $(BUILDDIR) && tar -czvf $(NAME)-$(VERSION)-$@.tar.gz $(NAME)-$(VERSION)-$@
	rm -rf $(BUILDDIR)/$(NAME)-$(VERSION)-$@ $(BUILDDIR)/$(NAME)-$@

linux-arm64:lion-ui
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@
	mkdir -p $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/lion/
	cp $(BUILDDIR)/$(NAME)-$@ $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(NAME)
	-cp config_example.yml $(BUILDDIR)/$(NAME)-$(VERSION)-$@/config_example.yml
	cp -r $(UIDIR)/lion/ $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/lion/
	cd $(BUILDDIR) && tar -czvf $(NAME)-$(VERSION)-$@.tar.gz $(NAME)-$(VERSION)-$@
	rm -rf $(BUILDDIR)/$(NAME)-$(VERSION)-$@ $(BUILDDIR)/$(NAME)-$@

windows-amd64:lion-ui
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@.exe
	mkdir -p $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/lion/
	cp $(BUILDDIR)/$(NAME)-$@.exe $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(NAME).exe
	-cp config_example.yml $(BUILDDIR)/$(NAME)-$(VERSION)-$@/config_example.yml
	cp -r $(UIDIR)/lion/ $(BUILDDIR)/$(NAME)-$(VERSION)-$@/$(UIDIR)/lion/
	cd $(BUILDDIR) && tar -czvf $(NAME)-$(VERSION)-$@.tar.gz $(NAME)-$(VERSION)-$@
	rm -rf $(BUILDDIR)/$(NAME)-$(VERSION)-$@ $(BUILDDIR)/$(NAME)-$@.exe

lion-ui:
	@echo "build ui"
	@cd $(UIDIR) && $(NPMINSTALL) && $(NPMBUILD)

clean:
	rm -rf $(BUILDDIR)
	-rm -rf $(UIDIR)/lion