package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

const wikiFile = "wikipedia_2020-04-05.htm"
const mapFile = "coronavirus-map-2020-03-22.html"

func wikiTableScraper() *TableScraper {
	return &TableScraper{
		URL:         "https://en.wikipedia.org/wiki/2019%E2%80%9320_coronavirus_pandemic_by_country_and_territory",
		CSSSelector: "div#covid19-container table.wikitable",
		ColumnDefs: []ColumnDef{
			{Skip: true},
			{TargetName: "country", Type: reflect.String, TruncateFrom: "["},
			{TargetName: "cases", Type: reflect.Int, ZeroValue: "–"},
			{TargetName: "deaths", Type: reflect.Int, ZeroValue: "–"},
			{TargetName: "recoveries", Type: reflect.Int, ZeroValue: "–"},
			{Skip: true},
		},
		HeaderRowIndex:  0,
		HeaderColNames:  []string{"countries", "cases", "deaths", "recov", "ref"},
		HeaderRowCount:  2,
		FooterRowCount:  2,
		TargetTableName: "wiki_entries",
	}
}

func mapTableScraper() *TableScraper {
	return &TableScraper{
		URL:         "https://google.com/covid19-map",
		CSSSelector: "div.table_container div.table_scroll.table_height table",
		ColumnDefs: []ColumnDef{
			{TargetName: "country", Type: reflect.String},
			{TargetName: "cases", Type: reflect.Int, ZeroValue: "—"},
			{TargetName: "cases1m", Type: reflect.Float64, ZeroValue: "—"},
			{TargetName: "recovered", Type: reflect.Int, ZeroValue: "—"},
			{TargetName: "deaths", Type: reflect.Int, ZeroValue: "—"},
		},
		HeaderRowIndex:  0,
		HeaderColNames:  []string{"Location", "Confirmed cases", "Cases per 1M people", "Recovered", "Deaths"},
		HeaderRowCount:  1,
		FooterRowCount:  0,
		TargetTableName: "map_entries",
	}
}

func TestScrapeTable(t *testing.T) {
	tableScraper := wikiTableScraper()
	require.NoError(t, ValidateTableScraper(tableScraper))

	rearrangedWikiScraper := wikiTableScraper()
	rearrangedWikiScraper.TargetColNames = []string{"deaths", "country", "cases", "recoveries"}

	tests := map[string]struct {
		inputFile    string
		scraper      *TableScraper
		wantRowCnt   int
		wantColNames []string
		wantCells0   []interface{}
	}{
		"wiki": {
			inputFile:    wikiFile,
			scraper:      wikiTableScraper(),
			wantRowCnt:   223,
			wantColNames: []string{"country", "cases", "deaths", "recoveries"},
			wantCells0:   []interface{}{"United States", 311616, 8489, 14943},
		},
		"map": {
			inputFile:    mapFile,
			scraper:      mapTableScraper(),
			wantRowCnt:   168,
			wantColNames: []string{"country", "cases", "cases1m", "recovered", "deaths"},
			wantCells0:   []interface{}{"Worldwide", 303594, 43.09, 94625, 12964},
		},
		"rearranged_wiki": {
			inputFile:    wikiFile,
			scraper:      rearrangedWikiScraper,
			wantRowCnt:   223,
			wantColNames: []string{"deaths", "country", "cases", "recoveries"},
			wantCells0:   []interface{}{8489, "United States", 311616, 14943},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			fpath := filepath.Join("testdata", tc.inputFile)
			r, err := os.Open(fpath)
			require.NoError(t, err)

			table, err := tc.scraper.scrapeFromReader(r)
			require.NoError(t, err)
			require.NotNil(t, table)

			require.Equal(t, tc.scraper.TargetTableName, table.Name)
			require.Equal(t, tc.wantRowCnt, len(table.Cells))
			require.Equal(t, tc.wantColNames, table.GetColumnNames())
			require.Equal(t, tc.wantCells0, table.Cells[0])
		})
	}
}
