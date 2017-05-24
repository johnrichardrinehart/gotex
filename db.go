package main

import (
	"database/sql"
	"fmt"
)

type HookInfo struct {
	URL       string // Repo URL
	Commit    string // commit hash
	Committer string // username
	Filename  string // compiled file
	Logname   string // log file
	Diffname  string // diff file
	Date      int    // UNIX timestamp
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
		path VARCHAR NOT NULL,
		url VARCHAR NOT NULL,
		[commit] VARCHAR NOT NULL,
		committer VARCHAR NOT NULL,
		filename VARCHAR NOT NULL,
		logname VARCHAR NOT NULL,
		diffname VARCHAR NOT NULL,
		date UNSIGNED BIG INT NOT NULL
	);
	`

	_, err := db.Exec(sql)

	if err != nil {
		panic(err)
	}
}

func dbRepoInfo(db *sql.DB, d string, u string, p string) []HookInfo {
	// d = domain (github.com)
	// u = user (fuzzybear3965)
	// p = project (gotex)
	fmt.Println("Grabbing rows", d, " ", u, " and", p, ".")
	path := fmt.Sprintf("%v/%v/%v", d, u, p) // maybe %v
	fmt.Println(path)
	stmt, err := db.Prepare(`SELECT url, [commit], committer, filename, logname, diffname, date FROM latex_builds WHERE path = $1`)
	defer stmt.Close()
	if err != nil {
		panic(err)
	}
	rows, err := stmt.Query(path)
	defer rows.Close()
	// make a container of rows that will be returned
	var dbRows []HookInfo
	// make containers for the scanned variables
	var url, commit, committer, filename, logname, diffname string
	var date int
	for rows.Next() {
		err := rows.Scan(&url, &commit, &committer, &filename, &logname, &diffname, &date)
		if err != nil {
			panic(err)
		}
		dbRows = append(dbRows, HookInfo{
			URL:       url,
			Commit:    commit,
			Committer: committer,
			Filename:  filename,
			Logname:   logname,
			Diffname:  diffname,
			Date:      date})
	}
	return dbRows
}

// func addRow(db *sql.DB) []row {
// 	fmt.Println("Adding rows", d, "and", r, ".")
// 	stmt, err := db.Prepare("SELECT cm, url FROM latex_builds WHERE domain = $1 AND repo = $2")
// 	defer stmt.Close()
// 	if err != nil {
// 		panic(err)
// 	}
// 	rows, err := stmt.Query(d, r)
// 	defer rows.Close()
// 	// make a container of rows that will be returned
// 	var dbRows []row
// 	// make containers for the scanned variables
// 	var cm string
// 	var url string
// 	for rows.Next() {
// 		err := rows.Scan(&cm, &url)
// 		if err != nil {
// 			panic(err)
// 		}
// 		dbRows = append(dbRows, row{URL: url, Commit: cm})
// 	}
// 	return dbRows
// }
