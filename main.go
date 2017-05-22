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
	r.Get("/:domain/:repo", getHandler(db))
	r.Post("/:domain/:repo", postHandler(db))
	http.ListenAndServe("127.0.0.1:80", r)
}
