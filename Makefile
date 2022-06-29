#-----------------------------------------------------------------------------------------------------------------------
# Variables (https://www.gnu.org/software/make/manual/html_node/Using-Variables.html#Using-Variables)
#-----------------------------------------------------------------------------------------------------------------------
.DEFAULT_GOAL := help

HOSTNAME = registry.terraform.io
NAMESPACE = auth0
NAME = auth0
BINARY = terraform-provider-${NAME}

BUILD_DIR ?= $(CURDIR)/out

GO_OS ?= $(shell go env GOOS)
GO_ARCH ?= $(shell go env GOARCH)
GO_PACKAGES := $(shell go list ./... | grep -v vendor)
GO_FILES := $(shell find . -name '*.go' | grep -v vendor)
GO_LINT_SCRIPT ?= $(CURDIR)/scripts/golangci-lint.sh
GO_TEST_COVERAGE_FILE ?= "coverage.out"

# Colors for the printf
RESET = $(shell tput sgr0)
COLOR_WHITE = $(shell tput setaf 7)
COLOR_BLUE = $(shell tput setaf 4)
COLOR_YELLOW = $(shell tput setaf 3)
TEXT_INVERSE = $(shell tput smso)

#-----------------------------------------------------------------------------------------------------------------------
# Rules (https://www.gnu.org/software/make/manual/html_node/Rule-Introduction.html#Rule-Introduction)
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: help docs-gen

help: ## Show this help
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

docs-gen: ## Generate docs for a specific resource. Usage: "make docs-gen RESOURCE=client
	@if [ -z "$(RESOURCE)" ]; \
	then \
	  echo "Please provide a resource. Example: make docs-gen RESOURCE=client" && exit 1; \
 	fi
	${call print, "Generating docs for resource: $(RESOURCE)"}
	@go run scripts/gendocs.go -resource "auth0_$(RESOURCE)"

#-----------------------------------------------------------------------------------------------------------------------
# Dependencies
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: deps deps-rm

deps: ## Download dependencies
	${call print, "Downloading dependencies"}
	@go mod vendor -v

deps-rm: ## Remove the dependencies folder
	${call print, "Removing the dependencies folder"}
	@rm -rfv vendor

#-----------------------------------------------------------------------------------------------------------------------
# Building & Installing
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: build install clean

build: ## Build the provider binary. Usage: "make build VERSION=0.2.0"
	${call print, "Building the provider binary"}
	@if [ -z "$(VERSION)" ]; \
	then \
	  echo "Please provide a version. Example: make build VERSION=0.2.0" && exit 1; \
 	fi
	@go build -v -o "${BUILD_DIR}/${BINARY}_v$(VERSION)"

install: build ## Install the provider as a terraform plugin. Usage: "make install VERSION=0.2.0"
	@mkdir -p "${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/$(VERSION)/${GO_OS}_${GO_ARCH}"
	@mv "${BUILD_DIR}/${BINARY}_v$(VERSION)" "${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/$(VERSION)/${GO_OS}_${GO_ARCH}"

clean: ## Clean up locally installed provider binaries."
	${call print_warning, "Cleaning locally installed provider binaries"}
	@rm -rf "${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}"

#-----------------------------------------------------------------------------------------------------------------------
# Code Style
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: lint

lint: ## Run go linter checks
	${call print, "Running golangci-lint over project"}
	@sh -c "${GO_LINT_SCRIPT}"

#-----------------------------------------------------------------------------------------------------------------------
# Testing
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: test test-unit test-acc test-sweep

test: test-unit test-acc ## Run all tests

test-unit: ## Run unit tests
	${call print, "Running unit tests"}
	@go test ${GO_PACKAGES} || exit 1
	@echo ${GO_PACKAGES} | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

test-acc: ## Run acceptance tests with http recordings
	${call print, "Running acceptance tests"}
	@AUTH0_HTTP_RECORDINGS=on AUTH0_DOMAIN=terraform-provider-auth0-dev.eu.auth0.com TF_ACC=1 \
		go test ${GO_PACKAGES} -v $(TESTARGS) -timeout 120m -coverprofile="${GO_TEST_COVERAGE_FILE}"

test-acc-e2e: ## Run acceptance tests end to end
	${call print, "Running acceptance tests E2E"}
	@TF_ACC=1 go test ${GO_PACKAGES} -v $(TESTARGS) -timeout 120m -coverprofile="${GO_TEST_COVERAGE_FILE}"

test-sweep: ## Clean up test tenant
	${call print_warning, "WARNING: This will destroy infrastructure. Use only in development accounts."}
	@go test ./auth0 -v -sweep="phony" $(SWEEPARGS)

#-----------------------------------------------------------------------------------------------------------------------
# Helpers
#-----------------------------------------------------------------------------------------------------------------------
define print
	@printf "${TEXT_INVERSE}${COLOR_WHITE} :: ${COLOR_BLUE} %-75s ${COLOR_WHITE} ${RESET}\n" $(1)
endef

define print_warning
	@printf "${TEXT_INVERSE}${COLOR_WHITE} ! ${COLOR_YELLOW} %-75s ${COLOR_WHITE} ${RESET}\n" $(1)
endef
