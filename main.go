package main

import (
	"flag"
	"fmt"
	"github.com/husobee/vestigo"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"os"
)

// initialize the logger variable
var logger *log.Logger

func init() {
	// make the assets directory if it doesn't already exist

	if err := os.MkdirAll("assets", os.ModePerm); err != nil {
		logger.Println(err)
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
	logFlag := flag.String("logfile", "gotex.log", "log filename\n\t")
	flag.Parse()
	logFlags := log.Ldate | log.Ltime | log.Lshortfile | log.LUTC
	logPrefix := ""
	// open up the logger file
	file, err := os.OpenFile(*logFlag, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	defer file.Close()
	if err != nil {
		logger.Println(err)
	} else {
		// use a multiwriter so STDOUT sees the output, too.
		logger = log.New(io.MultiWriter(os.Stdout, file), logPrefix, logFlags)
	}
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
	vestigo.CustomNotFoundHandlerFunc(indexHandler)
	// if it's the home page or some undefined route
	logger.Printf("----- gotex started: Listening on address %v. -----\n", *addrFlag)
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
