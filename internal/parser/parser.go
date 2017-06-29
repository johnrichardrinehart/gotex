package parser

import (
	"encoding/json"
	//"github.com/fuzzybear3965/gotex/internal/bitbucket"
	"github.com/fuzzybear3965/gotex/internal/github"
	"github.com/fuzzybear3965/gotex/internal/gitlab"
	"net/http"
	"net/url"
	"strings"
)

// TODO: Change DBRow for another name (it's not in db.go and has non-exported
// fields)
type DBRow struct {
	Timestamp string
	ID        string
	URL       string
	Message   string
	Username  string
	RealName  string
	PDFName   string
	LogName   string
	DiffName  string
	Path      string
	TeXRoot   string // root name of the main LaTeX file
}

func ParseHook(r *http.Request) []*DBRow {
	d := strings.Split(r.URL.Path, "/")[1] // /github.com/a/b -> github.com
	// Decode the JSON body into the appropriate struct
	if d == "github.com" {
		// Get the push event
		p := new(github.PushEvent)
		json.NewDecoder(r.Body).Decode(p)
		// get the url to be used as the path column (github.com/a/b)
		u, err := url.Parse(p.Repository.URL)
		if err != nil {
			panic(err)
		}
		// restruct the array of commits into a general purpose container
		h := make([]*DBRow, len(p.Commits))
		//fmt.Printf("%+v", r.URL.Query())
		for idx, c := range p.Commits {
			h[idx] = &DBRow{
				Timestamp: c.Timestamp,
				ID:        c.ID,
				URL:       c.URL,
				Message:   c.Message,
				Username:  c.Author.Username,
				RealName:  c.Author.RealName,
				Path:      u.Hostname() + u.Path,
				TeXRoot:   r.URL.Query().Get("root"), // in the Query
			}
		}
		return h
	} else if d == "gitlab.com" {
		// Get the push event
		p := new(gitlab.PushEvent)
		json.NewDecoder(r.Body).Decode(p)
		// get the url to be used as the path column (github.com/a/b)
		u, err := url.Parse(p.Repository.URL)
		if err != nil {
			panic(err)
		}
		// restruct the array of commits into a general purpose container
		h := make([]*DBRow, len(p.Commits))
		//fmt.Printf("%+v", r.URL.Query())
		for idx, c := range p.Commits {
			h[idx] = &DBRow{
				Timestamp: c.Timestamp,
				ID:        c.ID,
				URL:       c.URL,
				Message:   c.Message,
				Username:  c.Author.Username,
				RealName:  c.Author.Username, // gitlab has no RealName
				Path:      u.Hostname() + u.Path,
				TeXRoot:   r.URL.Query().Get("root"), // in the Query
			}
		}
		return h
	} else if d == "bitbucket.com" {
		//var p bitbucket.PushEvent
		return make([]*DBRow, 5)
	} else {
		return make([]*DBRow, 5)
		// TODOs
		// 1) log the fact that we accessed an unsupported domain.
		// 2) Return a page to the user that suggests they contact the CI admin to
		// implement this  domain
	}
}
