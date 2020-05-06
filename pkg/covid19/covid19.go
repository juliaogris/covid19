package covid19

import (
	"reflect"

	"github.com/juliaogris/covid19/pkg/table"
)

const WikiURL = "https://en.wikipedia.org/wiki/2019%E2%80%9320_coronavirus_pandemic_by_country_and_territory"

func newScraper(url string) *table.Scraper {
	dashes := []string{"-", "—", "–"}
	return &table.Scraper{
		URL:         url,
		CSSSelector: "div#covid19-container table.wikitable",
		ColumnDefs: []table.ColumnDef{
			{Skip: true},
			{TargetName: "country", Type: reflect.String, TruncateFrom: "["},
			{TargetName: "cases", Type: reflect.Int, ZeroValues: dashes},
			{TargetName: "deaths", Type: reflect.Int, ZeroValues: dashes},
			{TargetName: "recoveries", Type: reflect.Int, ZeroValues: dashes},
			{Skip: true},
		},
		HeaderRowIndex:  0,
		HeaderColNames:  []string{"locations", "cases", "deaths", "recov", "ref"},
		HeaderRowCount:  2,
		FooterRowCount:  2,
		TargetTableName: "entries",
		ContinueOnError: true,
	}
}

func ScrapeWiki(url, conn string) (*table.Table, error) {
	t, err := newScraper(url).Scrape()
	if err != nil {
		return nil, err
	}
	if err := table.Persist(conn, t); err != nil {
		return nil, err
	}
	return t, nil
}
