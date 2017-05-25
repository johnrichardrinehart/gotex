package main

import (
	"database/sql"
	//"encoding/json"
	"fmt"
	"github.com/fuzzybear3965/gotex/internal/parser"
	"github.com/husobee/vestigo"
	"html/template"
	"net/http"
)

type templateContainer struct {
	DBRows []*parser.DBRow
}

func getHandler(d *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: render the template immediately and serve the rows over WebSockets
		w.WriteHeader(200)
		rows := dbRepoInfo(
			d,
			vestigo.Param(r, "domain"),
			vestigo.Param(r, "user"),
			vestigo.Param(r, "repo"),
		)
		tpl := template.Must(template.New("repos.html").Delims("[[", "]]").ParseFiles("repos.html")) // .Must() panics if err is non-nil
		fmt.Printf("%+v", rows)
		tpl.Execute(w, templateContainer{DBRows: rows})
	}
}

func postHandler(d *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received a post request.")
		h := parser.ParseHook(r)
		ch := make(chan []*parser.DBRow)
		go addRows(d, h, ch)
		go compile(ch)
	}
}
