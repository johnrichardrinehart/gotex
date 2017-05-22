package main

import (
	"database/sql"
	"fmt"
)

type dbRows struct {
	urls []string // Git URL
	cms  []string // commit hash
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

func dbRepoInfo(db *sql.DB, d string, r string) dbRows {
	fmt.Println("Grabbing rows", d, "and", r, ".")
	stmt, err := db.Prepare("SELECT cm, url FROM latex_builds WHERE domain = $1 AND repo = $2")
	defer stmt.Close()
	if err != nil {
		panic(err)
	}
	rows, err := stmt.Query(d, r)
	defer rows.Close()
	// make a container of rows that will be returned
	//var rowsarr []row
	// make containers for the scanned variables
	var cm string
	var url string
	// make a container for the data to be returned
	var dbrows dbRows
	for rows.Next() {
		err := rows.Scan(&cm, &url)
		if err != nil {
			panic(err)
		}
		dbrows.urls = append(dbrows.urls, url)
		dbrows.cms = append(dbrows.cms, cm)
	}
	return dbrows
}
