package main

import (
	"database/sql"
	"github.com/fuzzybear3965/gotex/internal/parser"
	"github.com/husobee/vestigo"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

func getHandler(d *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		curPath, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		// Get the directory where this is running
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Error.Print(err)
		}
		os.Chdir(dir)           // change to path to run command
		defer os.Chdir(curPath) // go back to where we started
		//TODO: render the template immediately and serve the rows over WebSockets
		w.WriteHeader(200)
		rows := getRows(
			d,
			vestigo.Param(r, "domain"),
			vestigo.Param(r, "user"),
			vestigo.Param(r, "repo"),
		)
		if len(rows) > 0 {
			logger.Debug.Printf("Number of rows %v.\n", len(rows))
			tpl := template.Must(template.New("repos.html").Delims("[[", "]]").ParseFiles("assets/repos.html")) // .Must() panics if err is non-nil
			data := struct {
				DBRows []*parser.Commit
			}{
				rows,
			}
			tpl.Execute(w, data)
		} else {
			defaultHandler(w, r)
			logger.Debug.Printf("No rows.\n")
		}
	}
}

func postHandler(d *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug.Println("Received a post request.")
		h := parser.ParseHook(r) // parse the webhook into a container h
		ch := make(chan []*parser.Commit)
		go compile(h, ch)
		go addRows(d, ch)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.New("index.html").Delims("[[", "]]").ParseFiles("assets/index.html")) // .Must() panics if err is non-nil
	tpl.Execute(w, nil)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.New("default.html").Delims("[[", "]]").ParseFiles("assets/default.html")) // .Must() panics if err is non-nil
	tpl.Execute(w, nil)
}
