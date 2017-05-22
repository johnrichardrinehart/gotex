package main

import (
	"database/sql"
	"fmt"
)

type row struct {
	URL    string // Repo URL
	Commit string // commit hash
}

func initDB(fpath string) *sql.DB {
	db, err := sql.Open("sqlite3", fpath)

	if err != nil {
		panic(err)
	}

	if db == nil {
		panic("db == nil")
	}

	return db
}

func migrate(db *sql.DB) {
	sql := `
	CREATE TABLE IF NOT EXISTS latex_builds(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		domain VARCHAR NOT NULL,
		repo VARCHAR NOT NULL,
		cm VARCHAR NOT NULL,
		url VARCHAR NOT NULL
	);
	`

	_, err := db.Exec(sql)

	if err != nil {
		panic(err)
	}
}

func dbRepoInfo(db *sql.DB, d string, r string) []row {
	fmt.Println("Grabbing rows", d, "and", r, ".")
	stmt, err := db.Prepare("SELECT cm, url FROM latex_builds WHERE domain = $1 AND repo = $2")
	defer stmt.Close()
	if err != nil {
		panic(err)
	}
	rows, err := stmt.Query(d, r)
	defer rows.Close()
	// make a container of rows that will be returned
	var dbRows []row
	// make containers for the scanned variables
	var cm string
	var url string
	for rows.Next() {
		err := rows.Scan(&cm, &url)
		if err != nil {
			panic(err)
		}
		dbRows = append(dbRows, row{URL: url, Commit: cm})
	}
	return dbRows
}
