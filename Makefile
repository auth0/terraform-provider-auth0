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
GO_FMT_SCRIPT ?= $(CURDIR)/scripts/gofmtcheck.sh
GO_ERR_CHECK_SCRIPT ?= $(CURDIR)/scripts/errcheck.sh
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

build: fmt-check ## Build the provider binary. Usage: "make build VERSION=0.2.0"
	${call print, "Building the provider binary"}
	@if [ -z "$(VERSION)" ]; \
	then \
	  echo "Please provide a version. Example: make build VERSION=0.2.0" && exit 1; \
 	fi
	@go build -v -o "${BUILD_DIR}/${BINARY}_v$(VERSION)"

install: build ## Install the provider as a terraform plugin. Usage: "make install VERSION=0.2.0"
	@mkdir -p "${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/$(VERSION)/${GO_OS}_${GO_ARCH}"
	@mv "${BUILD_DIR}/${BINARY}_v$(VERSION)" "${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/$(VERSION)/${GO_OS}_${GO_ARCH}"

clean: ## Clean up installed provider binary. Usage: "make clean VERSION=0.2.0"
	${call print_warning, "Cleaning installed provider binary: ${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/$(VERSION)"}
	@if [ -z "$(VERSION)" ]; \
    	then \
    	  echo "Please provide a version. Example: make clean VERSION=0.2.0" && exit 1; \
     	fi
	@if [ -d "${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/$(VERSION)" ]; \
	then \
	  rm -rf "${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/$(VERSION)"; \
 	fi

#-----------------------------------------------------------------------------------------------------------------------
# Code Style
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: fmt fmt-check err-check

fmt: ## Format go files
	${call print, "Formatting go files"}
	@gofmt -w ${GO_FILES}

fmt-check: ## Check gofmt formatting
	${call print, "Checking that code complies with gofmt requirements"}
	@sh -c "${GO_FMT_SCRIPT}"

err-check: ## Check for unchecked errors
	${call print, "Checking for unchecked errors"}
	@sh -c "${GO_ERR_CHECK_SCRIPT}"

#-----------------------------------------------------------------------------------------------------------------------
# Testing
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: test test-unit test-acc test-sweep

test: test-unit test-acc ## Run all tests

test-unit: fmt-check ## Run unit tests
	${call print, "Running unit tests"}
	@go test ${GO_PACKAGES} || exit 1
	@echo ${GO_PACKAGES} | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

test-acc: fmt-check dev-up ## Run acceptance tests
	${call print, "Running acceptance tests"}
	@TF_ACC=1 go test ${GO_PACKAGES} -v $(TESTARGS) -timeout 120m -coverprofile="${GO_TEST_COVERAGE_FILE}"

test-sweep: ## Clean up test tenant
	${call print_warning, "WARNING: This will destroy infrastructure. Use only in development accounts."}
	@go test ./auth0 -v -sweep="phony" $(SWEEPARGS)

#-----------------------------------------------------------------------------------------------------------------------
# Development
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: dev-up dev-down dev-stop dev-clean

dev-up: ## Bootstrap the development containers
	${call print, "Starting development containers"}
	@docker-compose up -d

dev-down: ## Bring down the development containers
	${call print, "Bringing the development containers down"}
	@docker-compose down

dev-stop: ## Stop the development containers
	${call print, "Stopping the development containers"}
	@docker-compose stop

dev-rm: ## Delete the development containers
	${call print, "Deleting the development containers"}
	@docker-compose rm -f

#-----------------------------------------------------------------------------------------------------------------------
# Helpers
#-----------------------------------------------------------------------------------------------------------------------
define print
	@printf "${TEXT_INVERSE}${COLOR_WHITE} :: ${COLOR_BLUE} %-75s ${COLOR_WHITE} ${RESET}\n" $(1)
endef

define print_warning
	@printf "${TEXT_INVERSE}${COLOR_WHITE} ! ${COLOR_YELLOW} %-75s ${COLOR_WHITE} ${RESET}\n" $(1)
endef
