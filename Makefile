all: build test check-coverage lint  ## build, test, check coverage and lint
	@if [ -e .git/rebase-merge ]; then git --no-pager log -1 --pretty='%h %s'; fi
	@echo '$(COLOUR_GREEN)Success$(COLOUR_NORMAL)'

clean::  ## Remove generated files

.PHONY: all clean

# Get the first directory in GOPATH
GOPATH1 = $(firstword $(subst :, ,$(GOPATH)))

# -- Build and run -------------------------------------------------------------

BINARY = covid19-scraper
#	dbConnStr string    = "postgres://postgres@localhost:5432/covid19db?sslmode=disable"
# $(PGPASSWORD
build:  ## Build covid19-scraper
	go build -o $(BINARY) .

run: build check-pg-password  ## Build and run covid19-scraper with gcloud database, see make gcloud-proxy
	./$(BINARY) --conn="postgres://postgres@localhost:5555/covid19db?sslmode=disable"

run-local: build  ## Build and run covid19-scraper with local database
	./$(BINARY) --conn="postgres://postgres:postgres@localhost:5432/?sslmode=disable"

clean::
	rm -f $(BINARY)

.PHONY: build run run-local

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

GCLOUD_PORT = 5555

postgres:  ## Start Postgres docker container on port 5432
	docker run --rm --name postgres -e POSTGRES_PASSWORD=postgres -p5432:5432 postgres:11.7

psql:  ## Connect to local postgres docker container
	docker exec -it postgres psql -U postgres

gcloud-proxy: ## Start Google cloud proxy connected to google cloud database
	cloud_sql_proxy -instances=covid19-sars-cov-2:us-central1:covid19=tcp:$(GCLOUD_PORT)

gcloud-psql:  check-pg-password ## Connect to gcloud postgres database via gcloud-proxy
	docker exec -it -e PGPASSWORD=$(PGPASSWORD) postgres psql \
		-h docker.for.mac.localhost -U postgres -d covid19db -p $(GCLOUD_PORT)

check-pg-password:
ifeq ($(PGPASSWORD),)
	$(error PGPASSWORD environement varaible not set)
endif

.PHONY: check-pg-password gcloud-proxy gcloud-psql postgres psql

# --- Utilities ---------------------------------------------------------------

COLOUR_NORMAL = $(shell tput sgr0 2>/dev/null)
COLOUR_RED    = $(shell tput setaf 1 2>/dev/null)
COLOUR_GREEN  = $(shell tput setaf 2 2>/dev/null)
COLOUR_WHITE  = $(shell tput setaf 7 2>/dev/null)

help:
	@awk -F ':.*## ' 'NF == 2 && $$1 ~ /^[A-Za-z0-9_-]+$$/ { printf "$(COLOUR_WHITE)%-30s$(COLOUR_NORMAL)%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.PHONY: help
