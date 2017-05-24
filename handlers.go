package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/fuzzybear3965/gotex/github"
	"github.com/husobee/vestigo"
	"html/template"
	"net/http"
)

func getHandler(d *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		rows := dbRepoInfo(
			d,
			vestigo.Param(r, "domain"),
			vestigo.Param(r, "user"),
			vestigo.Param(r, "repo"),
		)
		tpl := template.Must(template.New("repos.html").Delims("[[", "]]").ParseFiles("repos.html")) // .Must() panics if err is non-nil
		tpl.Execute(w, rows)
	}
}

func postHandler(d *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		decoder := json.NewDecoder(r.Body)
		var p github.PushEvent
		err := decoder.Decode(&p)

		fmt.Printf("%+v\n", p)
		if err != nil {
			panic(err)
		}
	}
}
