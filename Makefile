# KUBECTL ARTILLERY PLUGIN
# ------------------------

THIS_FILE := $(lastword $(MAKEFILE_LIST))

# ARTILLERY_DISABLE_TELEMETRY defines whether telemetry should be enabled or not.
# By default, telemetry is disabled for all builds except public builds
ARTILLERY_DISABLE_TELEMETRY ?= true

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

##@ Build

## TO HELP KUBECTL DISCOVER AND RUN YOUR PLUGIN:
## After you build the binary, ensure it is available on your path.
# e.g.: export PATH="$GOPATH/src/github.com/artilleryio/kubectl-artillery/bin:$PATH"
# SEE using a plugin: https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/#using-a-plugin

build: ## Build kubectl-artillery plugin binary.
	go build -o bin/kubectl-artillery cmd/kubectl-artillery/main.go

run: ## Run kubectl-artillery plugin.
	ARTILLERY_DISABLE_TELEMETRY=true go run cmd/kubectl-artillery/main.go

##@ Release

# Create a draft release distribution of the kubectl-artillery plugin.
# The draft release is created using goreleaser - automatically downloaded locally if necessary.
# It creates a draft release page on Github and requires an access token stored in a .goreleaser-github-token file.
#
# Regarding the Github draft release page:
# ** ENSURE the draft release CHANGELOG only contains kubectl plugin features.
# ** ENSURE the repo is NOT in a dirty state before running the task.
.PHONY: kubeplugin-release, check-release-tag-version, check-release-tag-msg

#check-release-tag-version:
#ifndef KUBEPLUGIN_TAG_VERSION
#	$(error KUBEPLUGIN_TAG_VERSION a tag version is required to tag the kubectl-artillery release)
#endif
#
#check-release-tag-msg:
#ifndef KUBEPLUGIN_TAG_MSG
#	$(error KUBEPLUGIN_TAG_MSG a message is required to tag the kubectl-artillery release)
#endif

#kubeplugin-release: check-release-tag-version check-release-tag-msg goreleaser ## Creates a draft release of the artillery kubectl plugin on Github see .goreleaser.yaml.
#	git tag -a "v$(KUBEPLUGIN_TAG_VERSION)" -m "$(KUBEPLUGIN_TAG_MSG)"
#	git push --tags
#	$(GORELEASER) release --config .goreleaser.yaml --rm-dist

release: goreleaser ## Creates a draft release of the artillery kubectl plugin on Github see .goreleaser.yaml.
	$(GORELEASER) release --config .goreleaser.yaml --rm-dist

#clean-release-tags: check-release-tag-version ## Removes draft release tags to abort a release.
#	git tag -d "v$(KUBEPLUGIN_TAG_VERSION)"
#	git push origin ":refs/tags/v$(KUBEPLUGIN_TAG_VERSION)"


## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN): ## Ensure that the directory exists
	mkdir -p $(LOCALBIN)

## Tool Binaries
GORELEASER ?= $(LOCALBIN)/goreleaser

## Tool Versions
.PHONY: goreleaser
goreleaser: $(GORELEASER) ## Download goreleaser locally if necessary.
	GOBIN=$(LOCALBIN) go install github.com/goreleaser/goreleaser@latest
