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
GO_PACKAGES := $(shell go list ./... | grep -vE "vendor|tools|sweep|acctest")
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
.PHONY: help docs

help: ## Show this help
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

docs: ## Generate docs
	${call print, "Generating docs"}
	@go generate

#-----------------------------------------------------------------------------------------------------------------------
# Dependencies
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: deps deps-dev deps-rm

deps: ## Download dependencies
	${call print, "Downloading dependencies"}
	@go mod vendor -v

deps-dev: ## Download development dependencies
	${call print, "Installing golangci-lint"}
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	${call print, "Installing go vulnerability checker"}
	@go install golang.org/x/vuln/cmd/govulncheck@latest

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
	@go build -v -ldflags "-X github.com/auth0/terraform-provider-auth0/internal/provider.version=${VERSION}" -o "${BUILD_DIR}/${BINARY}_v$(VERSION)"

install: build ## Install the provider as a terraform plugin. Usage: "make install VERSION=0.2.0"
	@mkdir -p "${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/$(VERSION)/${GO_OS}_${GO_ARCH}"
	@mv "${BUILD_DIR}/${BINARY}_v$(VERSION)" "${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/$(VERSION)/${GO_OS}_${GO_ARCH}"

clean: ## Clean up locally installed provider binaries
	${call print_warning, "Cleaning locally installed provider binaries"}
	@rm -rf "${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}"

#-----------------------------------------------------------------------------------------------------------------------
# Checks
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: lint check-docs check-vuln

lint: ## Run go linter checks
	@if ! command -v golangci-lint &> /dev/null; \
	then \
		make deps-dev; \
 	fi
	${call print, "Running golangci-lint over project"}
	@golangci-lint run -v -c .golangci.yml ./...

check-docs: ## Check that documentation was generated correctly
	${call print, "Checking that documentation was generated correctly"}
	@go generate
	@if [ -n "$$(git status --porcelain)" ]; \
	then \
		echo "Go generate resulted in changed files:"; \
		echo "$$(git diff)"; \
		echo "Please run \`make docs\` to regenerate docs."; \
		exit 1; \
	fi
	@echo "Documentation is generated correctly."

check-vuln: ## Check go vulnerabilities
	@if ! command -v govulncheck &> /dev/null; \
	then \
		make deps-dev; \
 	fi
	${call print, "Running govulncheck over project"}
	@govulncheck -v ./...

#-----------------------------------------------------------------------------------------------------------------------
# Testing
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: test test-unit test-acc test-sweep

test-unit: ## Run unit tests. To run a specific test, pass the FILTER var. Usage `make test-unit FILTER="TestAccResourceServer`
	${call print, "Running unit tests"}
	@TF_ACC= \
		go test \
		-v \
		-run "$(FILTER)" \
		-timeout 30s \
		-coverprofile="${GO_TEST_COVERAGE_FILE}" \
		${GO_PACKAGES}

test-acc: ## Run acceptance tests with http recordings. To run a specific test, pass the FILTER var. Usage `make test-acc FILTER="TestAccResourceServer`
	${call print, "Running acceptance tests with http recordings"}
	@AUTH0_HTTP_RECORDINGS=on \
		AUTH0_DOMAIN=terraform-provider-auth0-dev.eu.auth0.com \
		TF_ACC=1 \
		go test \
		-v \
		-run "$(FILTER)" \
		-timeout 120m \
		-coverprofile="${GO_TEST_COVERAGE_FILE}" \
		${GO_PACKAGES}

test-acc-record: ## Run acceptance tests and record http interactions. To run a specific test, pass the FILTER var. Usage `make test-acc-record FILTER="TestAccResourceServer`
	${call print, "Running acceptance tests and recording http interactions"}
	@AUTH0_HTTP_RECORDINGS=on \
		TF_ACC=1 \
		go test \
		-v \
		-run "$(FILTER)" \
		-timeout 120m \
		${GO_PACKAGES}

test-acc-e2e: ## Run acceptance tests without http recordings. To run a specific test, pass the FILTER var. Usage `make test-acc-e2e FILTER="TestAccResourceServer`
	${call print, "Running acceptance tests against a real Auth0 tenant"}
	@TF_ACC=1 \
		go test \
		-v \
		-run "$(FILTER)" \
		-timeout 120m \
		-parallel 1 \
		-coverprofile="${GO_TEST_COVERAGE_FILE}" \
		${GO_PACKAGES}

test-sweep: ## Clean up test tenant
	${call print_warning, "WARNING: This will destroy infrastructure. Use only in development accounts."}
	@read -p "Continue? [y/N] " ans && ans=$${ans:-N} ; \
	if [ $${ans} = y ] || [ $${ans} = Y ]; then \
		go test ./internal/acctest/sweep -v -sweep="phony" $(SWEEPARGS) ; \
	fi

#-----------------------------------------------------------------------------------------------------------------------
# Helpers
#-----------------------------------------------------------------------------------------------------------------------
define print
	@printf "${TEXT_INVERSE}${COLOR_WHITE} :: ${COLOR_BLUE} %-75s ${COLOR_WHITE} ${RESET}\n" $(1)
endef

define print_warning
	@printf "${TEXT_INVERSE}${COLOR_WHITE} ! ${COLOR_YELLOW} %-75s ${COLOR_WHITE} ${RESET}\n" $(1)
endef
