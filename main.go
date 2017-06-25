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
	// make the assets directory if it doesn't already exist

	if err := os.MkdirAll("assets", os.ModePerm); err != nil {
		fmt.Println(err)
	}
	// taken from
	// https://stackoverflow.com/questions/11692860/how-can-i-efficiently-download-a-large-file-using-go
	grabAsset("repos.html")
	grabAsset("index.html")
	grabAsset("custom.css")
	grabAsset("drawArrows.js")
	grabAsset("sort.js")
}

func main() {
	addrFlag := flag.String("address", "127.0.0.1:8080", "listening address and port \n\t(e.g. \"gotex --address 127.0.0.1:8080\" or \"gotex --address :8080\")\n\t")
	flag.Parse()
	// Initialize the database
	db := initDB("gotex.db")
	defer db.Close()
	migrate(db)
	r := vestigo.NewRouter()
	// if it's a repo page
	r.Get("/:domain/:user/:repo", getHandler(db))
	r.Post("/:domain/:user/:repo", postHandler(db))
	// serve asset files
	r.Get("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))).ServeHTTP)
	// serve built files
	r.Get("/builds/*", http.StripPrefix("/builds/", http.FileServer(http.Dir("builds"))).ServeHTTP)
	// catch all
	r.Get("/*", indexHandler)
	// if it's the home page or some undefined route
	fmt.Printf("Listening on address %v.\n", *addrFlag)
	http.ListenAndServe(*addrFlag, r)
}

func grabAsset(path string) {
	file := fmt.Sprintf("assets/%v", path)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		out, _ := os.Create(file)
		defer out.Close()
		resp, _ := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/fuzzybear3965/gotex/master/%v", file))
		defer resp.Body.Close()
		io.Copy(out, resp.Body)
	}
}
