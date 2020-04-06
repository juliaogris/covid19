package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"
)

func persistTable(connStr string, t *Table) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	if err := setupSchema(db, t); err != nil {
		return err
	}
	return insertRows(db, t)
}

var identifierRe = regexp.MustCompile("^[_a-zA-Z]+[_a-zA-Z0-9]*$")

func setupSchema(db *sql.DB, t *Table) error {
	if err := createSchema(db, t); err != nil {
		return err
	}
	return validateSchema(db, t)
}

func createSchema(db *sql.DB, t *Table) error {
	if !identifierRe.MatchString(t.Name) {
		return fmt.Errorf("invalid table name, must be SQL identifier")
	}
	cols, err := getPQCols(t.Columns)
	if err != nil {
		return err
	}
	stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
   		id serial PRIMARY KEY,
	   	date timestamp NOT NULL,
	   	%s
	)`, t.Name, cols)
	_, err = db.Exec(stmt)
	return err
}

func getPQCols(cols []Column) (string, error) {
	s := make([]string, len(cols))
	var err error
	for i, col := range cols {
		s[i], err = getPQCol(col)
		if err != nil {
			return "", err
		}
	}
	return strings.Join(s, ",\n\t\t"), nil
}

func getPQCol(col Column) (string, error) {
	if !identifierRe.MatchString(col.Name) {
		return "", fmt.Errorf("invalid column name, must be SQL identifier")
	}

	pqType, err := getPQType(col.Type)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s", col.Name, pqType), nil
}

func getPQType(k reflect.Kind) (string, error) {
	switch k {
	case reflect.String:
		return "text", nil
	case reflect.Int, reflect.Int64:
		return "bigint", nil
	case reflect.Float64:
		return "double precision", nil
	}
	return "", fmt.Errorf("unknown kind %s", k)
}

func validateSchema(db *sql.DB, t *Table) error {
	types, err := getPQTypeMap(t)
	if err != nil {
		return err
	}

	q := "SELECT column_name, data_type, column_default FROM information_schema.columns WHERE table_name = $1;"
	rows, err := db.Query(q, t.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	cnt := 0
	for rows.Next() {
		if err := validateSchemaRow(rows, types); err != nil {
			return err
		}
		cnt++
	}
	if cnt != len(types) {
		return fmt.Errorf("expeceted %d columns, got %d", len(types), cnt)
	}
	return nil
}

func validateSchemaRow(rows *sql.Rows, types map[string]string) error {
	var colName, dataType string
	var colDefault *string

	if err := rows.Scan(&colName, &dataType, &colDefault); err != nil {
		return err
	}
	if types[colName] != dataType {
		return fmt.Errorf("expected pq type '%s' got '%s' for column '%s'", types[colName], dataType, colName)
	}
	if colName == "id" && (colDefault == nil || !strings.HasPrefix(*colDefault, "nextval")) {
		return fmt.Errorf("id is not serial type")
	}
	return nil
}

func getPQTypeMap(t *Table) (map[string]string, error) {
	m := map[string]string{"id": "integer", "date": "timestamp without time zone"}
	for _, col := range t.Columns {
		t, err := getPQType(col.Type)
		if err != nil {
			return nil, err
		}
		m[col.Name] = t
	}
	return m, nil
}

func insertRows(db *sql.DB, t *Table) error {
	date := time.Now().UTC()

	txn, err := db.Begin()
	if err != nil {
		return err
	}
	colNames := t.GetColumnNames()
	s := append([]string{"date"}, colNames...)
	stmt, err := txn.Prepare(pq.CopyIn(t.Name, s...))
	if err != nil {
		return err
	}

	for _, row := range t.Cells {
		vals := append([]interface{}{date}, row...)
		_, err = stmt.Exec(vals...)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	err = txn.Commit()
	if err != nil {
		return err
	}

	return nil
}
