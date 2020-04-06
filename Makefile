all: build test check-coverage lint  ## build, test, check coverage and lint
	@if [ -e .git/rebase-merge ]; then git --no-pager log -1 --pretty='%h %s'; fi
	@echo '$(COLOUR_GREEN)Success$(COLOUR_NORMAL)'

clean::  ## Remove generated files

.PHONY: all clean

# Get the first directory in GOPATH
GOPATH1 = $(firstword $(subst :, ,$(GOPATH)))

# -- Build and run -------------------------------------------------------------

BINARY = covid19-scraper
build:  ## Build covid19-scraper
	go build -o $(BINARY) .

run: build  ## Build and run covid19-scraper
	./$(BINARY)

clean::
	rm -f $(BINARY)

.PHONY: build run

# -- Lint ----------------------------------------------------------------------

lint:  ## lint the source code
	golangci-lint run

.PHONY: lint

# -- Test ----------------------------------------------------------------------
COVERFILE = coverage.out
COVERAGE = 43

test:  ## Run tests and generate a coverage file
	go test -coverprofile=$(COVERFILE) -short ./...

test-all: ## Run all tests, including some against DB
	go test ./...

check-coverage: test  ## Check that test coverage meets the required level
	@go tool cover -func=$(COVERFILE) | $(CHECK_COVERAGE) || $(FAIL_COVERAGE)

cover: test  ## Show test coverage in your browser
	go tool cover -html=$(COVERFILE)

clean::
	rm -f $(COVERFILE)

CHECK_COVERAGE = awk -F '[ \t%]+' '/^total:/ {print; if ($$3 < $(COVERAGE)) exit 1}'
FAIL_COVERAGE = { echo '$(COLOUR_RED)FAIL - Coverage below $(COVERAGE)%$(COLOUR_NORMAL)'; exit 1; }

.PHONY: test check-coverage cover

# -- DB ------------------------------------------------------------------------

postgres:
	docker run --rm --name postgres -e POSTGRES_PASSWORD=postgres -p5555:5432 postgres:11.7

psql:
	docker exec -it postgres psql -U postgres

.PHONY: postgres psql

# --- Utilities ---------------------------------------------------------------

COLOUR_NORMAL = $(shell tput sgr0 2>/dev/null)
COLOUR_RED    = $(shell tput setaf 1 2>/dev/null)
COLOUR_GREEN  = $(shell tput setaf 2 2>/dev/null)
COLOUR_WHITE  = $(shell tput setaf 7 2>/dev/null)

help:
	@awk -F ':.*## ' 'NF == 2 && $$1 ~ /^[A-Za-z0-9_-]+$$/ { printf "$(COLOUR_WHITE)%-30s$(COLOUR_NORMAL)%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.PHONY: help
