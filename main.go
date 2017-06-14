package main

import (
	//"fmt"
	"github.com/husobee/vestigo"
	//"log"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
	"os"
)

func init() {
	// taken from
	// https://stackoverflow.com/questions/11692860/how-can-i-efficiently-download-a-large-file-using-go
	if _, err := os.Stat("repos.html"); os.IsNotExist(err) {
		out, _ := os.Create("repos.html")
		defer out.Close()
		resp, _ := http.Get("https://raw.githubusercontent.com/fuzzybear3965/gotex/master/repos.html")
		defer resp.Body.Close()
		io.Copy(out, resp.Body)
	}
}

func main() {
	// Initialize the database
	db := initDB("gotex.db")
	defer db.Close()
	migrate(db)
	r := vestigo.NewRouter()
	// serve static assets
	r.Get("/:domain/:user/:repo", getHandler(db))
	r.Post("/:domain/:user/:repo", postHandler(db))
	r.Get("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))).ServeHTTP)
	r.Get("/builds/*", http.StripPrefix("/builds/", http.FileServer(http.Dir("builds"))).ServeHTTP)
	http.ListenAndServe("0.0.0.0:8000", r)
}
