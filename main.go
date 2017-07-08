package main

import (
	"flag"
	"fmt"
	"github.com/husobee/vestigo"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// initialize the logger variable
type loggerStruct struct {
	Info     *log.Logger
	Debug    *log.Logger
	Warning  *log.Logger
	Error    *log.Logger
	Fatal    *log.Logger
	Anything *log.Logger
}

var logger *loggerStruct

var addrFlag = flag.String("address", "127.0.0.1:8080", "listening address and port \n\t(e.g. \"gotex --address 127.0.0.1:8080\" or \"gotex --address :8080\")\n\t")
var logFlag = flag.String("logfile", "gotex.log", "log filename\n\t")
var debugFlag = flag.Bool("debug", false, "debug?")

// grab the working directory for renaming files
var WORKINGDIR, _ = filepath.Abs(filepath.Dir(os.Args[0]))

func main() {
	flag.Parse()
	logger, logfile := setUpLogger(*logFlag, log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	grabAssets()
	// Initialize the database
	db := initDB("gotex.db")
	defer db.Close()
	migrate(db)
	r := vestigo.NewRouter()
	// if it's the home page
	r.Get("/", rootHandler)
	// if it's a repo page
	r.Get("/:domain/:user/:repo", getHandler(db))
	r.Get("/:domain/:user/:repo/:branch", getHandler(db))
	r.Post("/:domain/:user/:repo", postHandler(db))
	// serve asset files
	r.Get("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))).ServeHTTP)
	// serve built files
	r.Get("/builds/*", http.StripPrefix("/builds/", http.FileServer(http.Dir("builds"))).ServeHTTP)
	// catch all
	vestigo.CustomNotFoundHandlerFunc(defaultHandler)
	// if it's the home page or some undefined route
	logger.Info.Printf("----- gotex started: Listening on address %v in directory %v. -----\n", *addrFlag, WORKINGDIR)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Info.Println("----- gotex shutting down, now :(. -----")
		logfile.Close()
		os.Exit(0)
	}()

	if err := http.ListenAndServe(*addrFlag, r); err != nil {
		logger.Fatal.Println(err)
	}
}

func grabAsset(path string) {
	// taken from
	// https://stackoverflow.com/questions/11692860/how-can-i-efficiently-download-a-large-file-using-go
	file := fmt.Sprintf("assets/%v", path)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		out, err := os.Create(file)
		defer out.Close()
		if err != nil {
			logger.Error.Printf("Could not create asset file assets/%v.", path)
		}
		url := fmt.Sprintf("https://raw.githubusercontent.com/fuzzybear3965/gotex/master/%v", file)
		resp, err := http.Get(url)
		defer resp.Body.Close()
		if err != nil {
			logger.Error.Printf("Error downloading asset %v.", url)
		}
		if _, err := io.Copy(out, resp.Body); err != nil {
			logger.Error.Printf("Could not write contents of file %v to %v.", url, path)
		}
	}
}

func grabAssets() {
	// make the assets directory if it doesn't already exist
	if err := os.MkdirAll("assets", os.ModePerm); err != nil {
		logger.Error.Println(err)
	} else {
		logger.Debug.Println("Made the assets directory successfully.")
	}
	assets := []string{
		"repos.html",
		"index.html",
		"custom.css",
		"drawArrows.js",
		"sort.js",
		"default.html",
	}
	for _, asset := range assets {
		grabAsset(asset)
	}
}

func setUpLogger(fn string, flags int) (*loggerStruct, *os.File) {
	// open up the logger file
	file, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	} else {
		// use a multiwriter so STDOUT sees the output, too.
		logger = &loggerStruct{
			Info:     log.New(io.MultiWriter(os.Stdout, file), "INFO: ", flags),
			Warning:  log.New(io.MultiWriter(os.Stdout, file), "WARNING: ", flags),
			Error:    log.New(io.MultiWriter(os.Stdout, file), "ERROR: ", flags),
			Fatal:    log.New(io.MultiWriter(os.Stdout, file), "FATAL: ", flags),
			Anything: log.New(io.MultiWriter(os.Stdout, file), "", flags),
		}
		if *debugFlag {
			logger.Debug = log.New(io.MultiWriter(os.Stdout, file),
				"DEBUG: ", flags)
		} else {
			logger.Debug = log.New(ioutil.Discard, "", flags)
		}
	}
	return logger, file
}
