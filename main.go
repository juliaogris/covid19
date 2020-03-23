package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var (
	sourceURL string = "https://google.org/crisisresponse/covid19-map"
	dbConnStr string = "postgres://postgres:postgres@localhost:5432/?sslmode=disable"
	// use with cloud_sql_proxy to write to google cloudSQL
	//dbConnStr string    = "postgres://postgres:<PWD>@localhost:5432/covid19db?sslmode=disable"
	out io.Writer = os.Stdout
)

type data struct {
	date    time.Time
	entries []entry
}

func (d *data) String() string {
	s := make([]string, len(d.entries))
	for i, e := range d.entries {
		s[i] = fmt.Sprintf("%32s %8d %8.1f %8d %8d", e.country, e.cases, e.cases1m, e.deaths, e.recoveries)
	}
	return strings.Join(s, "\n")
}

type entry struct {
	country    string
	cases      int
	cases1m    float32
	deaths     int
	recoveries int
}

func makeSnapshot(url string, connStr string) error {
	data, err := scrape(url)
	if err != nil {
		return (err)
	}
	fmt.Fprintln(out, data)
	return writeData(connStr, data)
}

func main() {
	if err := makeSnapshot(sourceURL, dbConnStr); err != nil {
		log.Fatal(err)
	}
}
