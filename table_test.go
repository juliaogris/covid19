package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func tableFixture() *Table {
	return &Table{
		Columns: []Column{
			{"c1", reflect.String},
			{"c2", reflect.String},
			{"c3", reflect.String},
			{"c4", reflect.String},
		},
		Cells: [][]interface{}{
			{"a1", "a2", "a3", "a4"},
			{"b1", "b2", "b3", "b4"},
		},
	}
}

func TestTableString(t *testing.T) {
	table := tableFixture()
	wantStr := strings.TrimSpace(`
c1 c2 c3 c4
a1 a2 a3 a4
b1 b2 b3 b4
`)
	require.Equal(t, wantStr, table.String())
}

func TestTableRearrange(t *testing.T) {
	table := tableFixture()
	targetCols := []string{"c3", "c1", "c2", "c4"}
	err := table.RearrangeColumns(targetCols)
	require.NoError(t, err)
	wantStr := strings.TrimSpace(`
c3 c1 c2 c4
a3 a1 a2 a4
b3 b1 b2 b4
`)
	require.Equal(t, wantStr, table.String())
	require.Equal(t, targetCols, table.GetColumnNames())
}
