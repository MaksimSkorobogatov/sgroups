export GOSUMDB=off
export GO111MODULE=on

$(value $(shell [ ! -d "$(CURDIR)/bin" ] && mkdir -p "$(CURDIR)/bin"))
export GOBIN=$(CURDIR)/bin

GO?=$(shell which go)
GIT_TAG:=$(shell git describe --exact-match --abbrev=0 --tags 2> /dev/null)
GIT_HASH:=$(shell git log --format="%h" -n 1 2> /dev/null)
GIT_BRANCH:=$(shell git branch 2> /dev/null | grep '*' | cut -f2 -d' ')
GO_VERSION:=$(shell go version | sed -E 's/.* go(.*) .*/\1/g')
BUILD_TS:=$(shell date +%FT%T%z)
VERSION:=$(shell cat ./VERSION 2> /dev/null | sed -n "1p")

.NOTPARALLEL:

.PHONY: help
help: ##display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

GOLANGCI_BIN:=$(GOBIN)/golangci-lint
GOLANGCI_REPO=https://github.com/golangci/golangci-lint
GOLANGCI_VERSION:=v2.11.4
ifneq ($(wildcard $(GOLANGCI_BIN)),)
	GOLANGCI_CUR_VERSION=$(strip v$(shell $(GOLANGCI_BIN) --version|sed -E 's/.*version (.*) built.*/\1/g'))
else
	GOLANGCI_CUR_VERSION=
endif

.PHONY: .install-linter
.install-linter:
ifneq ($(GOLANGCI_CUR_VERSION), $(GOLANGCI_VERSION))
	$(info Installing GOLANGCI-LINT $(GOLANGCI_VERSION)...)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_VERSION)
	@chmod +x $(GOLANGCI_BIN)
else
	@echo 1 >/dev/null
endif

.PHONY: lint
lint: ##run full lint
	@echo full lint... && \
	$(MAKE) .install-linter && \
	$(GOLANGCI_BIN) cache clean && \
	$(GOLANGCI_BIN) run --timeout=120s --config=$(CURDIR)/.golangci.yaml -v $(CURDIR)/... &&\
	echo -=OK=-


.PHONY: go-deps
go-deps: ##install golang dependencies
	@echo check go modules dependencies ... && \
	$(GO) mod tidy && \
 	GOWORK=off $(GO) mod vendor && \
	$(GO) mod verify && \
	echo -=OK=-

.PHONY: test
test: ##run tests
	@echo running tests... && \
	$(GO) clean -testcache && \
	$(GO) test -v -race ./... && \
	echo -=OK=-

os?=$(shell $(GO) env GOOS)
arch?=$(shell $(GO) env GOARCH)

PROJECT:=SGROUPS
APP?=sg-server
OUT?=$(CURDIR)/bin/$(APP)
APP_NAME?=$(PROJECT)/$(APP)
APP_VERSION:=$(if $(VERSION),$(VERSION),$(if $(GIT_TAG),$(GIT_TAG),$(GIT_BRANCH)))
APP_IDENTITY?=github.com/H-BF/corlib/app/identity
LDFLAGS?=-X '$(APP_IDENTITY).Name=$(APP_NAME)'\
		 -X '$(APP_IDENTITY).Version=$(APP_VERSION)'\
		 -X '$(APP_IDENTITY).BuildTS=$(BUILD_TS)'\
		 -X '$(APP_IDENTITY).BuildBranch=$(GIT_BRANCH)'\
		 -X '$(APP_IDENTITY).BuildHash=$(GIT_HASH)'\
		 -X '$(APP_IDENTITY).BuildTag=$(GIT_TAG)'\


.PHONY: sg-server
sg-server: ##build SGROUPS server. Usage: make sg-server [os=<linux|darwin>] [arch=<amd64|arm64>]
sg-server:
ifeq ($(filter linux darwin,$(os)),)
	$(error os=$(os) but must be in [linux|darwin])
endif
ifeq ($(filter amd64 arm64,$(arch)),)
	$(error arch=$(arch) but must be in [amd64|arm64])
endif
	@$(MAKE) go-deps && \
	echo build '$(APP)' for OS/ARCH='$(os)'/'$(arch)' ... && \
	echo into '$(OUT)' && \
	env GOOS=$(os) GOARCH=$(arch) CGO_ENABLED=0 \
	$(GO) build -ldflags="$(LDFLAGS)" -o $(OUT) $(CURDIR)/cmd/sg-server &&\
	echo -=OK=-

.PHONY: sg-agent
sg-agent: ##build SGROUPS agent. Usage: make sg-agent [platform=linux/<amd64|arm64>]
sg-agent: APP=sg-agent
sg-agent:
ifeq ($(filter amd64 arm64,$(arch)),)
	$(error arch=$(arch) but must be in [amd64|arm64])
endif
ifneq ($(os),linux)
	@$(MAKE) $@ os=linux
else
	@$(MAKE) go-deps &&\
	echo build \"$(APP)\" for OS/ARCH=\"$(os)/$(arch)\" ... && \
	echo into \"$(OUT)\" && \
	env GOOS=$(os) GOARCH=$(arch) CGO_ENABLED=0 \
	$(GO) build -ldflags="$(LDFLAGS)" -o $(OUT) $(CURDIR)/cmd/$(APP) &&\
	echo -=OK=-
endif


GOOSE_REPO:=https://github.com/Morwran/goose
GOOSE_VERSION:=v3.24.3-fork.1
GOOSE:=$(GOBIN)/goose

GOOSE_ARCH   := $(if $(filter $(arch),amd64),x86_64,$(arch))
GOOSE_ASSET  := goose_$(os)_$(GOOSE_ARCH)
GOOSE_BASEURL := $(GOOSE_REPO)/releases/download/$(GOOSE_VERSION)
GOOSE_URL := $(GOOSE_BASEURL)/$(GOOSE_ASSET)

.PHONY: .install-goose
.install-goose: ##build goose. Usage: make goose [os=<linux|darwin>] [arch=<amd64|arm64>]
ifneq ($(wildcard $(GOOSE)),)
	@echo >/dev/null
else
ifeq ($(filter linux darwin,$(os)),)
	$(error os=$(os) but must be in [linux|darwin])
endif
ifeq ($(filter amd64 arm64,$(arch)),)
	$(error arch=$(arch) but must be in [amd64|arm64])
endif
	@echo "Downloading $(GOOSE_ASSET) ($(GOOSE_VERSION))..."
	@mkdir -p $(GOBIN)
	@wget -q -O $(GOBIN)/goose $(GOOSE_URL)
	@chmod +x $(GOBIN)/goose
	@echo "goose installed to $(GOBIN)/goose"
endif


PG_MIGRATIONS?=$(CURDIR)/internal/sg-server/repository/pg/migrations

PG_URI?=
.PHONY: pg-migrations
pg-migrations: ##run Postgres migrations. Needed PG_URI environment variable. Example: make pg-migrations PG_URI="postgres://user_admin:qwerty@localhost:15432/sgroups?sslmode=disable"
ifneq ($(PG_URI),)
	@$(MAKE) .install-goose
	@echo "Generating combined init migration (gen-migration.sh)..."
	@cd $(PG_MIGRATIONS) && ./gen-migration.sh
	@echo "Running Postgres migrations..."
	@$(GOOSE) -table=sg_db_ver -dir=$(PG_MIGRATIONS) -v postgres $(PG_URI) up
else
	$(error need define PG_URI environment variable)
endif