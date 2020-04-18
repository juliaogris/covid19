package table

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Table struct {
	Name    string
	Columns []Column
	Cells   [][]interface{} // string, int64, float32 TODO: better: Rows []Row
}

type Column struct {
	Name string
	Type reflect.Kind
}

func (t *Table) GetColumnNames() []string {
	colNames := make([]string, len(t.Columns))
	for i, c := range t.Columns {
		colNames[i] = c.Name
	}
	return colNames
}

func (t *Table) String() string {
	format := t.getFormat()

	colRow := make([]interface{}, len(t.Columns))
	for i, col := range t.Columns {
		colRow[i] = col.Name
	}

	rows := append([][]interface{}{colRow}, t.Cells...)
	s := make([]string, len(rows))
	for i, row := range rows {
		s[i] = fmt.Sprintf(format, row...)
	}
	return strings.Join(s, "\n")
}

func (t *Table) getFormat() string {
	p := make([]int, len(t.Columns))
	for i, c := range t.Columns {
		p[i] = len(c.Name)
	}
	for _, row := range t.Cells {
		for i, cell := range row {
			l := len(fmt.Sprint(cell))
			if l > p[i] {
				p[i] = l
			}
		}
	}
	formats := make([]string, len(t.Columns))
	for i, pad := range p {
		formats[i] = "%" + strconv.Itoa(pad) + "v"
	}
	return strings.Join(formats, " ")
}

func (t *Table) RearrangeColumns(targetCols []string) error {
	colNames := t.GetColumnNames()
	if err := validateTargetCols(colNames, targetCols); err != nil {
		return err
	}
	colCnt := len(t.Columns)
	columns := make([]Column, colCnt)
	cells := make([][]interface{}, len(t.Cells))
	for i := range cells {
		cells[i] = make([]interface{}, colCnt)
	}

	for i, c := range targetCols {
		idx := index(colNames, c)
		columns[i] = t.Columns[idx]
		for j, row := range t.Cells {
			cells[j][i] = row[idx]
		}
	}
	t.Columns = columns
	t.Cells = cells
	return nil
}

func validateTargetCols(currentCols, targetCols []string) error {
	ccols := append([]string(nil), currentCols...)
	sort.Strings(ccols)
	tcols := append([]string(nil), targetCols...)
	sort.Strings(tcols)
	if !reflect.DeepEqual(ccols, tcols) {
		return fmt.Errorf("target cols don't equal current columns - %+v, %+v", tcols, ccols)
	}
	return nil
}

func index(slice []string, s string) int {
	for i := range slice {
		if slice[i] == s {
			return i
		}
	}
	return -1
}
