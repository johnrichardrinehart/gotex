package main

import (
	"database/sql"
	"fmt"
	"github.com/fuzzybear3965/gotex/internal/parser"
	"os"
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
	url VARCHAR NOT NULL,
	branch VARCHAR,
	message VARCHAR,
	username VARCHAR NOT NULL,
	realname VARCHAR NOT NULL,
	pdfname VARCHAR NOT NULL,
	logname VARCHAR NOT NULL,
	diffname VARCHAR NOT NULL,
	path VARCHAR NOT NULL,
	UNIQUE(id, path)
    );
    `

	_, err := db.Exec(sql)

	if err != nil {
		logger.Fatal.Println(err)
	}
}

// d = domain (e.g. github.com)
// u = user (e.g. fuzzybear3965)
// p = project (e.g. gotex)
// b = branch (e.g. master)
func getRows(db *sql.DB, d string, u string, r string, b string) []*parser.Commit {
	// where are these files stored?
	path := fmt.Sprintf("%v/%v/%v", d, u, r) // maybe %v
	// if they want a specific branch...
	if b != "" {
		logger.Info.Printf("Grabbing rows for %v/%v/%v and branch %v.\n", d, u, r, b)
	} else {
		logger.Info.Printf("Grabbing rows for %v/%v/%v.\n", d, u, r)
	}
	stmt, err := db.Prepare(`SELECT timestamp, id, url, branch, message, username, realname, pdfname, logname, diffname FROM latex_builds WHERE path = $1`)
	defer stmt.Close()
	if err != nil {
		logger.Fatal.Println(err)
	}
	rows, err := stmt.Query(path)
	defer rows.Close()
	// make a container of rows that will be returned
	var dbRows []*parser.Commit
	// make containers for the scanned variables
	var timestamp, id, url, branch, message, username, realname, pdfname, logname, diffname string
	for rows.Next() {
		err := rows.Scan(&timestamp, &id, &url, &branch, &message, &username, &realname, &pdfname, &logname, &diffname)
		if err != nil {
			logger.Fatal.Println(err)
		}
		if b == "" || b == branch {
			dbRows = append(dbRows, &parser.Commit{
				Timestamp: timestamp,
				ID:        id,
				URL:       url,
				Branch:    branch,
				Message:   message,
				Username:  username,
				RealName:  realname,
				PDFName:   pdfname,
				LogName:   logname,
				DiffName:  diffname,
			})
		}
	}
	return dbRows
}

// addRows accepts as arguments 1) db in which to store rows and 2) an array of rows to add
func addRows(db *sql.DB, c chan []*parser.Commit) {
	h := <-c
	logger.Info.Println("\nReceived", len(h), "rows to process.")
	qstmt, err := db.Prepare(`SELECT * FROM latex_builds WHERE id = ? and path = ?`)
	defer qstmt.Close()
	if err != nil {
		logger.Fatal.Println(err)
	}
	istmt, err := db.Prepare(`INSERT INTO latex_builds(timestamp, id, url, branch, message, username, realname, pdfname, logname, diffname, path) values(?,?,?,?,?,?,?,?,?,?,?)`)
	defer istmt.Close()
	if err != nil {
		logger.Fatal.Println(err)
	}
	// Check if we already have any of these rows
	// Loop over the rows to add
	commitNumber := 0
	var removeIdxs []int
	for i, r := range h {
		// Check if the PDF, DIFF, LOG were generated properly
		// Root PDF
		if _, err := os.Stat(fmt.Sprintf("builds/%v/%v/%v.pdf", r.Path, r.ID, r.TeXRoot)); err == nil {
			r.PDFName = fmt.Sprintf("/builds/%v/%v/%v.pdf", r.Path, r.ID, r.TeXRoot)
		} else {
			logger.Error.Printf("File builds/%v/%v/%v.pdf doesn't exist. Maybe compilation failed.\n", r.Path, r.ID, r.TeXRoot)
		}
		// Diff PDF
		if _, err := os.Stat(fmt.Sprintf("builds/%v/%v/%v.diff.pdf", r.Path, r.ID, r.ID)); err == nil {
			r.DiffName = fmt.Sprintf("/builds/%v/%v/%v.diff.pdf", r.Path, r.ID, r.ID)
		} else {
			logger.Error.Printf("File builds/%v/%v/%v.diff.pdf doesn't exist. Maybe compilation failed.\n", r.Path, r.ID, r.ID)
		}
		// Log PDF
		if _, err := os.Stat(fmt.Sprintf("builds/%v/%v/%v.log", r.Path, r.ID, r.TeXRoot)); err == nil {
			r.LogName = fmt.Sprintf("/builds/%v/%v/%v.log", r.Path, r.ID, r.TeXRoot)
		} else {
			logger.Error.Printf("File builds/%v/%v/%v.log doesn't exist. Maybe compilation was never attempted.\n", r.Path, r.ID, r.TeXRoot)
		}

		commitNumber += 1
		logger.Debug.Println("Working through commit #", commitNumber)
		// See if any row has this combination of id and path
		rows, err := qstmt.Query(r.ID, r.Path)
		defer rows.Close()
		if err != nil {
			logger.Fatal.Println(err)
		}
		// if we have matches the store this index to clean up h later
		if rows.Next() {
			// we should only have one match on any ID and Path per the Unique()
			// condition in the scheme for latex_builds
			logger.Warning.Printf("I've already added row with commit hash %v to the database.\n", r.ID)
			removeIdxs = append(removeIdxs, i)
		} else {
			_, err := istmt.Exec(r.Timestamp, r.ID, r.URL, r.Branch, r.Message, r.Username, r.RealName, r.PDFName, r.LogName, r.DiffName, r.Path)
			if err != nil {
				logger.Fatal.Println(err)
			}
		}
	}
	for i := len(removeIdxs) - 1; i >= 0; i-- {
		h = append(h[:i], h[i+1:]...)
	}
	logger.Info.Printf("Added %v rows to the database.\n", len(h))
}
