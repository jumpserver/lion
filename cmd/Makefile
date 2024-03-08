NAME=guacd-client-cli


GOOS:=$(shell go env GOOS)
GOARCH:=$(shell go env GOARCH)

LDFLAGS=-w -s


GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags "$(GOLDFLAGS) ${LDFLAGS}"


define make_artifact_full
	GOOS=$(1) GOARCH=$(2) $(GOBUILD) -o $(NAME)-$(1)-$(2)
endef

build:
	GOARCH=$(GOARCH) GOOS=$(GOOS) $(GO_BUILD) -o $(NAME) .

all: 
	$(call make_artifact_full,darwin,amd64)
	$(call make_artifact_full,darwin,arm64)
	$(call make_artifact_full,linux,amd64)
	$(call make_artifact_full,linux,arm64)
	$(call make_artifact_full,linux,mips64le)
	$(call make_artifact_full,linux,ppc64le)
	$(call make_artifact_full,linux,s390x)
	$(call make_artifact_full,linux,riscv64)

local: 
	$(call make_artifact_full,$(shell go env GOOS),$(shell go env GOARCH))

darwin-amd64: 
	$(call make_artifact_full,darwin,amd64)

darwin-arm64: 
	$(call make_artifact_full,darwin,arm64)

linux-amd64: 
	$(call make_artifact_full,linux,amd64)

linux-arm64: 
	$(call make_artifact_full,linux,arm64)

linux-loong64: 
	$(call make_artifact_full,linux,loong64)

linux-ppc64le: 
	$(call make_artifact_full,linux,ppc64le)

linux-mips64le: 
	$(call make_artifact_full,linux,mips64le)

linux-s390x: 
	$(call make_artifact_full,linux,s390x)

linux-riscv64: 
	$(call make_artifact_full,linux,riscv64)


clean:
	rm -f $(NAME)-*-*

# root/Lay-131132