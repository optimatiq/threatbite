SHELL := /bin/bash
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

.DEFAULT_GOAL: all

GIT_TAG := `git describe --abbrev=0 --tags`
GIT_COMMIT := `git rev-parse HEAD`

LDFLAGS=-ldflags "-s -w -X=main.date=$(shell date +%FT%T%z) -X=main.tag=$(GIT_TAG) -X=main.commit=$(GIT_COMMIT) "

.PHONY: build check clean format format-check git-tag-major git-tag-minor git-tag-patch help test tidy

all: check test build ## Default target: check, test, build,

build: ## Build all excecutables, located under ./bin/
	@echo "[threatbite] Building..."
	@go build -trimpath -o ./bin/threatbite $(LDFLAGS) cmd/threatbite/main.go

clean: ## Remove all artifacts from ./bin/ and ./resources
	@rm -rf ./bin/*

format: ## Format go code with goimports
	@go get golang.org/x/tools/cmd/goimports
	@goimports -l -w .

format-check: ## Check if the code is formatted
	@go get golang.org/x/tools/cmd/goimports
	@for i in $$(goimports -l .); do echo "[ERROR] Code is not formated run 'make format'" && exit 1; done

test: ## Run tests
	@go test -race  ./...

tidy: ## Run go mod tidy
	@go mod tidy

check: format-check ## Linting and static analysis
	@if grep -r --include='*.go' -E "fmt.Print|spew.Dump" *; then \
		echo "code contains fmt.Print* or spew.Dump function"; \
		exit 1; \
	fi

	@if test ! -e ./bin/golangci-lint; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh; \
	fi
	@./bin/golangci-lint run --timeout 180s -E gosec -E stylecheck -E golint -E goimports -E whitespace

git-tag-patch: ## Push new tag to repository with patch number incremented
	$(eval NEW_VERSION=$(shell git describe --tags --abbrev=0 | awk -F'[a-z.]' '{$$4++;print "v" $$2 "." $$3 "." $$4}'))
	@echo Version: $(NEW_VERSION)
	@git tag -a $(NEW_VERSION) -m "new patch release"
	@git push origin $(NEW_VERSION)

git-tag-minor: ## Push new tag to repository with minor number incremented
	$(eval NEW_VERSION=$(shell git describe --tags --abbrev=0 | awk -F'[a-z.]' '{$$3++;print "v" $$2 "." $$3 "." 0}'))
	@echo Version: $(NEW_VERSION)
	@git tag -a $(NEW_VERSION) -m "new minor release"
	@git push origin $(NEW_VERSION)

git-tag-major:  ## Push new tag to repository with major number incremented
	$(eval NEW_VERSION=$(shell git describe --tags --abbrev=0 | awk -F'[a-z.]' '{$$2++;print "v" $$2 "." 0 "." 0}'))
	@echo Version: $(NEW_VERSION)
	@git tag -a $(NEW_VERSION) -m "new major release"
	@git push origin $(NEW_VERSION)

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
