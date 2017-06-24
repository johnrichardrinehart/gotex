package main

import (
	"fmt"
	"github.com/husobee/vestigo"
	_ "github.com/mattn/go-sqlite3"
	"io"
	//"log"
	"flag"
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
	if _, err := os.Stat("assets/custom.css"); os.IsNotExist(err) {
		if err := os.MkdirAll("assets", os.ModePerm); err != nil {
			fmt.Println(err)
		} else {
			out, _ := os.Create("assets/custom.css")
			defer out.Close()
			resp, _ := http.Get("https://raw.githubusercontent.com/fuzzybear3965/gotex/master/assets/custom.css")
			defer resp.Body.Close()
			io.Copy(out, resp.Body)
		}
	}
}

func main() {
	addrFlag := flag.String("address", "127.0.0.1:8080", "listening address and port")
	flag.Parse()
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
	fmt.Printf("Listening on address %v.\n", *addrFlag)
	http.ListenAndServe(*addrFlag, r)
}
