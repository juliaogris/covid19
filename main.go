package main

import (
	"flag"
	"fmt"
	"log"
	"reflect"
)

const (
	defaultConn = "postgres://postgres:postgres@localhost:5432/?sslmode=disable"
	defaultURL  = "https://en.wikipedia.org/wiki/2019%E2%80%9320_coronavirus_pandemic_by_country_and_territory"
)

var (
	conn      = flag.String("conn", defaultConn, "postgres db connection string")
	scrapeURL = flag.String("url", defaultURL, "url to scrap table from")
)

func newWikiCovid19Scraper() *TableScraper {
	dashes := []string{"-", "—", "–"}
	return &TableScraper{
		URL:         *scrapeURL,
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
	flag.Parse()
	table, err := newWikiCovid19Scraper().Scrape()
	if err != nil {
		log.Fatal(err)
	}
	if err := persistTable(*conn, table); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully added", len(table.Cells), "rows.")
}
