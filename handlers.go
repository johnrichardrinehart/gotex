package main

import (
	"database/sql"
	"fmt"
	"github.com/husobee/vestigo"
	"html/template"
	"net/http"
	"os"
)

func getHandler(d *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		dbrows := dbRepoInfo(d, vestigo.Param(r, "domain"), vestigo.Param(r, "repo")) // struct
		var tmplvars = struct {
			Urls []string
			Cms  []string
		}{
			Urls: dbrows.urls,
			Cms:  dbrows.cms,
		}
		tpl := template.Must(template.New("").Delims("[[", "]]").ParseFiles("repos.html")) // .Must() panics if err is non-nil
		tpl.Execute(os.Stdout, tmplvars)
		tpl.Execute(w, tmplvars)
	}
}

func postHandler(d *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Println("PUT")
		fmt.Println(vestigo.Param(r, "domain"))
		fmt.Println(vestigo.Param(r, "repo"))
	}
}
