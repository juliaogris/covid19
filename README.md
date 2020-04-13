# Covid19 data scraper

Scrapes data from Wikipedia's [coronavirus pandemic by country and territory](https://en.wikipedia.org/wiki/2019%E2%80%9320_coronavirus_pandemic_by_country_and_territory) and write it to database.

## Development

Pre-requisites

    - go 1.14
    - gnu make
    - docker

Run local DB with

    make postgres

and in other terminal run scraper and check results with

    make run
    make psql

Find further options with

    make help

## Google Cloud access and deployment

Pre-requisites

    - cloud_sql_proxy
    - gcloud CLI

Execute scraper locally, write data to cloud and check results.

Start gcp-proxy

    make gcp-proxy

and in other terminal

    make gcp-run
    make gcp-psql

Deploy cloudfunctions with

    make deploy

Scraper will be executed twice daily via gcp scheduler and can be
triggered as http endpoint at:

    https://<REGION>-<GCP-PROJECT-ID>.cloudfunctions.net/Covid19HTTP
