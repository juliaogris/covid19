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
	sourceURL string    = "https://google.org/crisisresponse/covid19-map"
	out       io.Writer = os.Stdout
)

type data struct {
	date    time.Time
	entries []entry
}

func (d *data) String() string {
	s := make([]string, len(d.entries))
	for i, e := range d.entries {
		s[i] = fmt.Sprintf("%32s %8d %8.1f %8d %8d", e.country, e.cases, e.casesPer1M, e.deaths, e.recoveries)
	}
	return strings.Join(s, "\n")
}

type entry struct {
	country    string
	cases      int
	casesPer1M float64
	deaths     int
	recoveries int
}

func main() {
	data, err := scrape(sourceURL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(out, data)
}
