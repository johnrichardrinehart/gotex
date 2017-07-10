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

var wsClients = make(clients)

func wsHandler(d *sql.DB) http.HandlerFunc {
	n_clients := 0
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Fatal.Println(err)
		} else {
			conn.WriteMessage(websocket.TextMessage, []byte("gotex + WebSockets == \"glorious configuration\""))

			// connection isn't registered
			if _, ok := wsClients[conn]; !ok {
				n_clients += 1
				wsClients[conn] = &client{[]string{}, n_clients}
			}
		}
		// wait for client messages
		go func() {
			jsonmsg := make(map[string]string)
			for {
				defer conn.Close()
				if err := conn.ReadJSON(&jsonmsg); err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
						logger.Fatal.Printf("Unexpected WebSocket error: %v", err)
					} else {
						if jsonmsg["loc"] != "" {
							logger.Info.Printf("WebSocket client departing %v.\n", jsonmsg["loc"])
							n_clients -= 1
							logger.Debug.Printf("%v clients remaining.", n_clients)
							delete(wsClients, conn)

						}
					}
					break
				}
				if jsonmsg["loc"] != "" {
					logger.Debug.Printf("WebSocket client %v accessing %v.\n", wsClients[conn].id, jsonmsg["loc"])
					//wsClients[cIdx].urls = append(wsClients[cIdx].urls, jsonmsg["loc"])
					wsClients[conn].urls = append(wsClients[conn].urls, jsonmsg["loc"])

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
