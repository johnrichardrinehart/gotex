package main

import (
	"database/sql"
	"github.com/fuzzybear3965/gotex/internal/parser"
	"github.com/husobee/vestigo"
	"html/template"
	"net/http"
	"path/filepath"
)

func getHandler(d *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: render the template immediately and serve the rows over WebSockets
		w.WriteHeader(200)
		rows := getRows(
			d,
			vestigo.Param(r, "domain"),
			vestigo.Param(r, "user"),
			vestigo.Param(r, "repo"),
			vestigo.Param(r, "branch"),
		)
		if len(rows) > 0 {
			logger.Debug.Printf("Number of rows %v.\n", len(rows))
			tpl := template.Must(template.New("repos.html").Delims("[[", "]]").ParseFiles(filepath.Join(WORKINGDIR, "assets/repos.html"))) // .Must() panics if err is non-nil
			// if the user actually came here to view a single
			// branch's data
			show_branches := false
			// unless it isn't
			if vestigo.Param(r, "branch") == "" {
				show_branches = true
			}
			data := struct {
				DBRows           []*parser.Commit
				ShowBranchesBool bool
			}{
				rows,
				show_branches,
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
		queries := r.URL.Query()
		h := parser.ParseHook(r, queries) // parse webhook obtaining a []*parser.Commit h or nil
		if h != nil {
			ch := make(chan []*parser.Commit)
			go compile(h, ch)
			go addRows(d, ch)
		} else {
			logger.Debug.Println("parser returned nil. Branch possibly isn't supported.")
		}
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.New("index.html").Delims("[[", "]]").ParseFiles("assets/index.html")) // .Must() panics if err is non-nil
	tpl.Execute(w, nil)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	// grab the working directory for renaming files
	tpl := template.Must(template.New("default.html").Delims("[[", "]]").ParseFiles(filepath.Join(WORKINGDIR, "assets/default.html"))) // .Must() panics if err is non-nil
	tpl.Execute(w, nil)
}
