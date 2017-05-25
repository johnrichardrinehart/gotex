package parser

import (
	"encoding/json"
	//"fmt"
	//"github.com/fuzzybear3965/gotex/internal/bitbucket"
	//"github.com/fuzzybear3965/gotex/internal/commits"
	"github.com/fuzzybear3965/gotex/internal/github"
	//"github.com/fuzzybear3965/gotex/internal/gitlab"
	"net/http"
	"net/url"
)

type DBRow struct {
	Timestamp string
	ID        string
	URL       string
	Message   string
	UserName  string
	RealName  string
	PDFName   string
	LogName   string
	DiffName  string
	Path      string
}

func ParseHook(r *http.Request) []*DBRow {
	d := r.Header["Origin"][0]
	// Decode the JSON body into the appropriate struct
	if d == "https://github.com" {
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
		for idx, c := range p.Commits {
			h[idx] = &DBRow{
				Timestamp: c.Timestamp,
				ID:        c.ID,
				URL:       c.URL,
				Message:   c.Message,
				UserName:  c.Author.UserName,
				RealName:  c.Author.RealName,
				Path:      u.Hostname() + u.Path,
			}
		}
		return h
	} else if d == "gitlab.com" {
		//var p gitlab.PushEvent
		return make([]*DBRow, 5)
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
	//if d == "github.com" {
	//}
	//} else if d == "gitlab.com" {
	//return &hookInfo{}
	//} else if d == "bitbucket.com" {
	//return &hookInfo{}
	//}
}
