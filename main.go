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

func init() {
	// make the assets directory if it doesn't already exist
	if err := os.MkdirAll("assets", os.ModePerm); err != nil {
		logger.Error.Println(err)
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
	flag.Parse()
	logger, logfile := setUpLogger(*logFlag, log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
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
	logger.Info.Printf("----- gotex started: Listening on address %v. -----\n", *addrFlag)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Info.Println("----- gotex shutting down, now :(. -----")
		logfile.Close()
		os.Exit(0)
	}()

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
