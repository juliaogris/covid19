# Covid19 data scraper

Scrapes data from https://google.org/crisisresponse/covid19-map and prints it.

Next steps:

-   dump data into Google Cloud SQL
-   run in cloudfunction as Chron job
-   scrape data from Wikipedia source
-   scrape country population stats
-   scrape country hospital / ICU stats

Setting up work with gcloud

    cloud_sql_proxy -instances=covid19-sars-cov-2:us-central1:covid19=tcp:5432
    psql "host=127.0.0.1 port=5432 sslmode=disable dbname=covid19db user=postgres"

Local postgres DB:

    make postgres

and for the psql CLI

    make psql

Ask @juliaogris for db password if needed.
