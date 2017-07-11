package main

import (
	"database/sql"
	"github.com/fuzzybear3965/gotex/internal/parser"
	"github.com/gorilla/websocket"
	"github.com/husobee/vestigo"
	"html/template"
	"net/http"
	"path/filepath"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var curClients = make(clients)

func wsHandler(d *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Fatal.Println(err)
		} else {
			conn.WriteMessage(websocket.TextMessage, []byte("gotex + WebSockets == \"glorious configuration\""))
			// connection hasn't been registered
			if _, iscurclient := curClients[conn]; !iscurclient {
				// have we seen this dude, before?
				curClients[conn] = &client{url: "", id: len(curClients) + 1}
			} else {
				logger.Fatal.Println("Somehow a user has been registered having had sent a new request. This doesn't make any sense.")
			}
		}
		// wait for client messages
		go func() {
			defer conn.Close()
			jsonmsg := make(map[string]string)
			for {
				if err := conn.ReadJSON(&jsonmsg); err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
						logger.Fatal.Printf("Unexpected WebSocket error: %v", err)
					} else {
						logger.Debug.Printf("Departing: WebSocket client %v at %v.\n", curClients[conn].id, curClients[conn].url)

					}
					delete(curClients, conn)
					logger.Debug.Printf("%v clients remaining.", len(curClients))
					break
				}
				if jsonmsg["loc"] != "" {
					curClients[conn].url = jsonmsg["loc"]
					logger.Debug.Printf("Arriving: WebSocket client %v at %v.\n", curClients[conn].id, curClients[conn].url)

				}
			}
		}()
	}
}

func getHandler(d *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: render the template immediately and serve the rows over WebSockets
		w.WriteHeader(200)
		rows := getRows(
			d,
			vestigo.Param(r, "domain"),
			vestigo.Param(r, "user"),
			vestigo.Param(r, "repo"),
			vestigo.Param(r, "branch"),
		)
		if len(rows) > 0 {
			logger.Debug.Printf("Number of rows %v.\n", len(rows))
			tpl := template.Must(template.New("repos.html").Delims("[[", "]]").ParseFiles(filepath.Join(WORKINGDIR, "assets/repos.html"))) // .Must() panics if err is non-nil
			// if the user actually came here to view a single
			// branch's data
			show_branches := false
			// unless it isn't
			if vestigo.Param(r, "branch") == "" {
				show_branches = true
			}
			data := struct {
				DBRows           []*parser.Commit
				ShowBranchesBool bool
			}{
				rows,
				show_branches,
			}
			tpl.Execute(w, data)
		} else {
			defaultHandler(w, r)
			logger.Debug.Printf("No rows.\n")
		}
	}
}

func postHandler(d *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug.Println("Received a post request.")
		queries := r.URL.Query()
		h := parser.ParseHook(r, queries) // parse webhook obtaining a []*parser.Commit h or nil
		if h != nil {
			ch := make(chan []*parser.Commit)
			go compile(h, ch)
			go addRows(d, ch)
		} else {
			logger.Debug.Println("parser returned nil. Branch possibly isn't supported.")
		}
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.New("index.html").Delims("[[", "]]").ParseFiles("assets/index.html")) // .Must() panics if err is non-nil
	tpl.Execute(w, nil)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	// grab the working directory for renaming files
	tpl := template.Must(template.New("default.html").Delims("[[", "]]").ParseFiles(filepath.Join(WORKINGDIR, "assets/default.html"))) // .Must() panics if err is non-nil
	tpl.Execute(w, nil)
}
