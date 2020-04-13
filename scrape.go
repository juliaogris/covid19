package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/juliaogris/covid19/htmlx"
	"golang.org/x/net/html"
)

type TableScraper struct {
	URL         string
	CSSSelector string

	ColumnDefs []ColumnDef

	HeaderRowIndex int
	HeaderColNames []string
	HeaderRowCount int
	FooterRowCount int

	TargetTableName string
	TargetColNames  []string // must match ColumnDefs[i].TargetName; for rearranging
}

type ColumnDef struct { //nolint:maligned
	Skip bool

	TargetName   string
	Type         reflect.Kind
	ZeroValues   []string // e.g. "-" for numbers
	TruncateFrom string   // e.g. "[" to remove reference in wikipedia "[a]"
	NoTrim       bool     // don't trim whitespace
}

func (t *TableScraper) Scrape() (*Table, error) {
	if err := ValidateTableScraper(t); err != nil {
		return nil, err
	}
	resp, err := http.Get(t.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return t.scrapeFromReader(resp.Body)
}

func ValidateTableScraper(t *TableScraper) error {
	if _, err := url.Parse(t.URL); err != nil {
		return err
	}
	if len(t.HeaderColNames) > 0 {
		if t.HeaderRowIndex >= t.HeaderRowCount {
			return fmt.Errorf("header row index outside header row count")
		}
	}
	return nil
}

func (t *TableScraper) scrapeFromReader(r io.Reader) (*Table, error) {
	node, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	tableContainer, err := htmlx.QuerySelector(node, t.CSSSelector)
	if err != nil {
		return nil, err
	}
	rows := getRows(tableContainer)
	if len(rows) < t.HeaderRowCount+t.FooterRowCount {
		return nil, fmt.Errorf("expected at least %d rows, got %d", t.HeaderRowCount+t.FooterRowCount, len(rows))
	}
	if err := vaildateTableHeader(rows, t.HeaderColNames, t.HeaderRowIndex); err != nil {
		return nil, err
	}
	bodyRows := rows[t.HeaderRowCount : len(rows)-t.FooterRowCount]
	table, err := parseTableBody(bodyRows, t.ColumnDefs)
	if err != nil {
		return nil, err
	}
	table.Name = t.TargetTableName
	if len(t.TargetColNames) != 0 {
		if err := table.RearrangeColumns(t.TargetColNames); err != nil {
			return nil, err
		}
	}
	return table, nil
}

func getRows(table *html.Node) [][]string {
	trs := htmlx.QueryAllNoChildren(table, &htmlx.Selector{Tag: "tr"})
	rows := make([][]string, len(trs))
	for i, tr := range trs {
		rows[i] = getRow(tr)
	}
	return rows
}

func getRow(tr *html.Node) []string {
	th := htmlx.QueryAllNoChildren(tr, &htmlx.Selector{Tag: "th"})
	td := htmlx.QueryAllNoChildren(tr, &htmlx.Selector{Tag: "td"})
	cells := append(th, td...)
	row := make([]string, len(cells))
	for i, cell := range cells {
		row[i] = strings.TrimSpace(getText(cell))
	}
	return row
}

func vaildateTableHeader(rows [][]string, colNames []string, rowIndex int) error {
	if len(colNames) == 0 {
		return nil
	}
	cells := rows[rowIndex]
	if len(cells) != len(colNames) {
		return fmt.Errorf("expected %d columns, got %d", len(colNames), len(cells))
	}
	for i, colName := range colNames {
		s := strings.ToLower(strings.TrimSpace(cells[i]))
		if !strings.Contains(s, strings.ToLower(colName)) {
			return fmt.Errorf("expected header '%s' to contain '%s'", s, colName)
		}
	}
	return nil
}

func parseTableBody(rows [][]string, colDefs []ColumnDef) (*Table, error) {
	cells := make([][]interface{}, len(rows))
	targetColCnt := getTargetColCnt(colDefs)
	var err error
	for i, row := range rows {
		cells[i], err = parseRow(row, colDefs, targetColCnt)
		if err != nil {
			return nil, err
		}
	}
	columns := getTargetColumns(colDefs)
	return &Table{Columns: columns, Cells: cells}, nil
}

func getTargetColCnt(colDefs []ColumnDef) int {
	result := len(colDefs)
	for _, colDef := range colDefs {
		if colDef.Skip {
			result--
		}
	}
	return result
}

func getTargetColumns(colDefs []ColumnDef) []Column {
	cols := []Column{}
	for _, colDef := range colDefs {
		if !colDef.Skip {
			col := Column{Name: colDef.TargetName, Type: colDef.Type}
			cols = append(cols, col)
		}
	}
	return cols
}

func parseRow(row []string, colDefs []ColumnDef, targetColCnt int) ([]interface{}, error) {
	if len(row) != len(colDefs) {
		return nil, fmt.Errorf("expected %d data cells, got %d (%#v)", len(colDefs), len(row), row)
	}
	var err error
	result := make([]interface{}, targetColCnt)
	j := 0
	for i, colDef := range colDefs {
		if colDef.Skip {
			continue
		}
		result[j], err = parseCell(row[i], colDef)
		if err != nil {
			return nil, err
		}
		j++
	}
	return result, nil
}

func parseCell(c string, colDef ColumnDef) (interface{}, error) {
	if colDef.TruncateFrom != "" {
		if i := strings.Index(c, colDef.TruncateFrom); i != -1 {
			c = c[:i]
		}
	}
	if !colDef.NoTrim {
		c = strings.TrimSpace(c)
	}
	if contains(colDef.ZeroValues, c) {
		return zero(colDef.Type)
	}
	switch colDef.Type {
	case reflect.String:
		return c, nil
	case reflect.Int, reflect.Int64:
		c = strings.ReplaceAll(c, ",", "")
		return strconv.Atoi(c)
	case reflect.Float64:
		c = strings.ReplaceAll(c, ",", "")
		return strconv.ParseFloat(c, 64)
	}
	return nil, fmt.Errorf("unknown column type %s", colDef.Type)
}

func zero(k reflect.Kind) (interface{}, error) {
	switch k {
	case reflect.String:
		return "", nil
	case reflect.Int, reflect.Int64:
		return 0, nil
	case reflect.Float64:
		return 0.0, nil
	}
	return nil, fmt.Errorf("unknown kind %s", k)
}

func getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return strings.Trim(n.Data, "\n")
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getText(c))
	}
	return sb.String()
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
