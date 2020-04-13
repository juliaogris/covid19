# Covid19 data scraper

Scrapes data from Wikipedia's [coronavirus pandemic by country and territory](https://en.wikipedia.org/wiki/2019%E2%80%9320_coronavirus_pandemic_by_country_and_territory) and write it to database.

Pre-requites

    - go 1.14
    - gnu make
    - docker

Run local DB with

    make postgres

and in other terminal run scraper and check results with

    make run-local
    make psql

Write data to cloud and check results with

    gcloud-proxy
    make run
    gcloud-psql

Ask @juliaogris for db password if needed.

Find further options with

    make help
