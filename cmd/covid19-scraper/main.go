package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/juliaogris/covid19/pkg/covid19"
)

var (
	conn      = flag.String("conn", "postgres://postgres:postgres@localhost:5432/?sslmode=disable", "postgres db connection string")
	scrapeURL = covid19.WikiURL
)

func main() {
	flag.Parse()
	t, err := covid19.ScrapeWiki(scrapeURL, *conn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully added", len(t.Cells), "rows.")
}
