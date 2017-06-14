package main

import (
	//"fmt"
	"github.com/husobee/vestigo"
	//"log"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
)

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
