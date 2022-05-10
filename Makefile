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
