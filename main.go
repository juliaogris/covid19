package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
)

var (
	dbConnStr string = "postgres://postgres:postgres@localhost:5555/?sslmode=disable"
	// use with cloud_sql_proxy to write to google cloudSQL
	// dbConnStr string    = "postgres://postgres:<redacted>@localhost:5432/covid19db?sslmode=disable"
	out       io.Writer = os.Stdout
	scrapeURL           = "https://en.wikipedia.org/wiki/2019%E2%80%9320_coronavirus_pandemic_by_country_and_territory"
)

func newWikiCovid19Scraper() *TableScraper {
	dashes := []string{"-", "—", "–"}
	return &TableScraper{
		URL:         scrapeURL,
		CSSSelector: "div#covid19-container table.wikitable",
		ColumnDefs: []ColumnDef{
			{Skip: true},
			{TargetName: "country", Type: reflect.String, TruncateFrom: "["},
			{TargetName: "cases", Type: reflect.Int, ZeroValues: dashes},
			{TargetName: "deaths", Type: reflect.Int, ZeroValues: dashes},
			{TargetName: "recoveries", Type: reflect.Int, ZeroValues: dashes},
			{Skip: true},
		},
		HeaderRowIndex:  0,
		HeaderColNames:  []string{"countries", "cases", "deaths", "recov", "ref"},
		HeaderRowCount:  2,
		FooterRowCount:  2,
		TargetTableName: "entries",
	}
}

func main() {
	table, err := newWikiCovid19Scraper().Scrape()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(out, table)
	if err := persistTable(dbConnStr, table); err != nil {
		log.Fatal(err)
	}
}
