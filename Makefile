all: build test check-coverage lint  ## build, test, check coverage and lint
	@if [ -e .git/rebase-merge ]; then git --no-pager log -1 --pretty='%h %s'; fi
	@echo '$(COLOUR_GREEN)Success$(COLOUR_NORMAL)'

clean::  ## Remove generated files

.PHONY: all clean


# --- Build and run

BINARY = covid19-scraper

build:  ## Build covid19-scraper
	go build -o $(BINARY) ./cmd/$(BINARY)

run: build  ## Build and run covid19-scraper with local database
	./$(BINARY) --conn="postgres://postgres:postgres@localhost/?sslmode=disable"

clean::
	rm -f $(BINARY)

.PHONY: build run


# --- Lint

lint:  ## lint the source code
	golangci-lint run

.PHONY: lint


# --- Test

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

.PHONY: test test-all check-coverage cover


# --- DB

postgres:  ## Start Postgres docker container and expose default port 5432
	docker run --rm --name postgres -e POSTGRES_PASSWORD=postgres -p5432:5432 postgres:11.7

psql:  ## Connect to local postgres docker container
	docker exec -it postgres psql -U postgres

.PHONY:   postgres psql


# --- GCP run scraper and access db

GCP_DBNAME = covid19db
GCP_DBHOST = covid19-sars-cov-2:us-central1:covid19
GCP_DBUSER = postgres
GCP_DBPORT = 5555

gcp-proxy: ## Start Google cloud proxy connected to google cloud database
	cloud_sql_proxy -instances=$(GCP_DBHOST)=tcp:$(GCP_DBPORT)

gcp-psql: check-pg-password ## Connect to gcloud postgres database via gcloud-proxy
	docker exec -it -e PGPASSWORD=$(PGPASSWORD) postgres psql \
		-h docker.for.mac.localhost -U $(GCP_DBUSER) -d $(GCP_DBNAME) -p $(GCP_DBPORT)

gcp-run: build check-pg-password  ## Build and run covid19-scraper with gcloud database, see make gcloud-proxy
	./$(BINARY) --conn="postgres://$(GCP_DBUSER)@localhost:$(GCP_DBPORT)/$(GCP_DBNAME)?sslmode=disable"

check-pg-password:
ifeq ($(PGPASSWORD),)
	$(error PGPASSWORD environment variable not set)
endif

.PHONY: check-pg-password gcp-proxy gcp-psql gcp-run


# --- GCP deploy

GCP_RUNTIME = go113
GCP_ENVVARS = PGPASSWORD=$(PGPASSWORD),PGUSER=$(GCP_DBUSER),PGDATABASE=$(GCP_DBNAME),PGHOST=/cloudsql/$(GCP_DBHOST)

deploy:	deploy-http deploy-event deploy-scheduler ## deploy covid19-scraper to GCP

deploy-http: build check-pg-password
	gcloud functions deploy Covid19HTTP \
		--runtime $(GCP_RUNTIME) \
		--trigger-http \
		--allow-unauthenticated \
		--set-env-vars=$(GCP_ENVVARS)

deploy-event: build check-pg-password
	gcloud functions deploy Covid19Event \
		--runtime $(GCP_RUNTIME) \
		--trigger-topic=schedule \
		--allow-unauthenticated \
		--set-env-vars=$(GCP_ENVVARS)

deploy-scheduler: # only once on initial set up
	-gcloud scheduler jobs create pubsub covid19-scrape-job
		--schedule="0 */12 * * *" \
		--topic="schedule" \
		--message-body="go scrape" \
		--description="daily covid19 data scrape trigger"

.PHONY: deploy deploy-event deploy-http deploy-scheduler


# --- Utilities

COLOUR_NORMAL = $(shell tput sgr0 2>/dev/null)
COLOUR_RED    = $(shell tput setaf 1 2>/dev/null)
COLOUR_GREEN  = $(shell tput setaf 2 2>/dev/null)
COLOUR_WHITE  = $(shell tput setaf 7 2>/dev/null)

help:
	@awk -F ':.*## ' 'NF == 2 && $$1 ~ /^[A-Za-z0-9_-]+$$/ { printf "$(COLOUR_WHITE)%-30s$(COLOUR_NORMAL)%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.PHONY: help
