package main

import (
	"database/sql"
	"fmt"
	"github.com/fuzzybear3965/gotex/internal/parser"
)

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
		n INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		timestamp VARCHAR NOT NULL,
		id VARCHAR NOT NULL,
		message VARCHAR,
		url VARCHAR NOT NULL,
		username VARCHAR NOT NULL,
		realname VARCHAR NOT NULL,
		pdfname VARCHAR NOT NULL,
		diffname VARCHAR NOT NULL,
		logname VARCHAR NOT NULL,
		path VARCHAR NOT NULL,
		UNIQUE(id, path)
	);
	`

	_, err := db.Exec(sql)

	if err != nil {
		panic(err)
	}
}

func dbRepoInfo(db *sql.DB, d string, u string, p string) []*parser.DBRow {
	// d = domain (github.com)
	// u = user (fuzzybear3965)
	// p = project (gotex)
	fmt.Println("Grabbing rows", d, " ", u, " and", p, ".")
	path := fmt.Sprintf("%v/%v/%v", d, u, p) // maybe %v
	fmt.Println(path)
	stmt, err := db.Prepare(`SELECT timestamp, id, message, url, username, realname, pdfname, logname, diffname FROM latex_builds WHERE path = $1`)
	defer stmt.Close()
	if err != nil {
		panic(err)
	}
	rows, err := stmt.Query(path)
	defer rows.Close()
	// make a container of rows that will be returned
	var dbRows []*parser.DBRow
	// make containers for the scanned variables
	var timestamp, id, message, url, username, realname, pdfname, logname, diffname string
	for rows.Next() {
		err := rows.Scan(&timestamp, &id, &message, &url, &username, &realname, &pdfname, &logname, &diffname)
		if err != nil {
			panic(err)
		}
		dbRows = append(dbRows, &parser.DBRow{
			Timestamp: timestamp,
			ID:        id,
			Message:   message,
			URL:       url,
			UserName:  username,
			RealName:  realname,
			PDFName:   pdfname,
			LogName:   logname,
			DiffName:  diffname,
		})
	}
	return dbRows
}

func addRows(db *sql.DB, h []*parser.DBRow) {
	fmt.Println("Going to add some rows.")
	qstmt, err := db.Prepare(`SELECT * FROM latex_builds WHERE id = ? and path = ?`)
	istmt, err := db.Prepare(`INSERT INTO latex_builds(timestamp, id, message, url, username, realname, pdfname, logname, diffname, path) values(?,?,?,?,?,?,?,?,?,?)`)
	defer qstmt.Close()
	defer istmt.Close()
	if err != nil {
		panic(err)
	}

	// Check if we already have any of these rows
	// Loop over the rows to add
	commitNumber := 0
	for i, r := range h {
		commitNumber += 1
		fmt.Println("Working through commit #", commitNumber)
		// See if any row has this combination of id and path
		res, err := qstmt.Query(r.ID, r.Path)
		if err != nil {
			panic(err)
		}
		// if we have matches the store this index to clean up h later
		var removeIdxs []int
		if res.Next() {
			fmt.Println("We have a match for that row.")
			// we should only have one match on any ID and Path per the Unique()
			// condition in the scheme for latex_builds
			removeIdxs = append(removeIdxs, i)
		} else {
			fmt.Println("No rows found like that, yet.")
			res.Close()
			// Compile the document and, if successful, log the row to the database
			_, err := istmt.Exec(r.Timestamp, r.ID, r.Message, r.URL, r.UserName, r.RealName, r.PDFName, r.LogName, r.DiffName, r.Path)
			if err != nil {
				panic(err)
			}
		}
	}
}
