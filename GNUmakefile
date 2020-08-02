SHELL = bash
PROJECT_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_DIRTY := $(if $(shell git status --porcelain),+CHANGES)

THIS_OS := $(shell uname | cut -d- -f1)

SOURCE_FILES = $(shell find . -name '*.go')

GO_LDFLAGS := "-X github.com/tsarna/zone-update/version.GitCommit=$(GIT_COMMIT)$(GIT_DIRTY)"

ALL_TARGETS += linux_amd64 \
	linux_arm64

ifeq (Darwin,$(THIS_OS))
ALL_TARGETS += darwin_amd64
endif

default: release

pkg/darwin_amd64/zone-update: $(SOURCE_FILES) ## Build zone-update for darwin/amd64
	@echo "==> Building $@..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
		go build \
		-trimpath \
		-ldflags $(GO_LDFLAGS) \
		-o "$@"

pkg/linux_amd64/zone-update: $(SOURCE_FILES) ## Build zone-update for linux/amd64
	@echo "==> Building $@..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build \
		-trimpath \
		-ldflags $(GO_LDFLAGS) \
		-o "$@"

pkg/linux_arm64/zone-update: $(SOURCE_FILES) ## Build zone-update for linux/arm64
	@echo "==> Building $@..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
		go build \
		-trimpath \
		-ldflags $(GO_LDFLAGS) \
		-o "$@"

# Define targets for each supported platform

define makePackageTarget

pkg/$(1).zip: pkg/$(1)/zone-update
	@echo "==> Packaging for $(1)..."
	@zip -j pkg/$(1).zip pkg/$(1)/*

endef

$(foreach t,$(ALL_TARGETS),$(eval $(call makePackageTarget,$(t))))

.PHONY: release
release: clean $(foreach t,$(ALL_TARGETS),pkg/$(t).zip)

.PHONY: clean
clean: GOPATH=$(shell go env GOPATH)
clean: ## Remove build artifacts
	@echo "==> Cleaning build artifacts..."
	@rm -rf "$(PROJECT_ROOT)/bin/"
	@rm -rf "$(PROJECT_ROOT)/pkg/"
	@rm -f "$(GOPATH)/bin/zone-update"
