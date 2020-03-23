package main

import (
	"database/sql"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func writeData(connStr string, d *data) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	if err := setupSchema(db); err != nil {
		return err
	}
	return insertEntries(db, d)
}

func setupSchema(db *sql.DB) error {
	stmt := `CREATE TABLE IF NOT EXISTS entries (
   		id serial PRIMARY KEY,
	   	date TIMESTAMP NOT NULL,
	   	country VARCHAR(64) NOT NULL,
	   	cases INTEGER NOT NULL,
	   	cases1m FLOAT(32) NOT NULL,
	   	deaths INTEGER NOT NULL,
	   	recoveries INTEGER NOT NULL
	)`
	_, err := db.Exec(stmt)
	return err
}

func insertEntries(db *sql.DB, d *data) error {
	txn, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := txn.Prepare(pq.CopyIn("entries", "date", "country", "cases", "cases1m", "deaths", "recoveries"))
	if err != nil {
		return err
	}

	for _, e := range d.entries {
		_, err = stmt.Exec(d.date, e.country, e.cases, e.cases1m, e.deaths, e.recoveries)
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
