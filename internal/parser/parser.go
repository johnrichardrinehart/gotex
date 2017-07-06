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

type Commit struct {
	Timestamp string
	ID        string
	URL       string
	Branch    string
	Message   string
	Username  string
	RealName  string
	PDFName   string
	LogName   string
	DiffName  string
	Path      string // path to the directory storing this compilation
	TeXRoot   string // root name of the main LaTeX file
}

func ParseHook(r *http.Request, queries url.Values) []*Commit {
	branches := queries.Get("branches")
	d := strings.Split(r.URL.Path, "/")[1] // github.com/a/b -> github.com
	// Decode the JSON body into the appropriate struct
	if d == "github.com" {
		// Get the push event
		p := new(github.PushEvent)        // GitHub push event container
		json.NewDecoder(r.Body).Decode(p) // decoded push event
		ref := strings.Split(p.Ref, "/")
		branch := ref[len(ref)-1]
		if contains(strings.Split(branches, ","), branch) {
			// get the url to be used as the path column (github.com/a/b)
			u, err := url.Parse(p.Repository.URL)
			if err != nil {
				panic(err)
			}
			// restruct the array of commits into a general purpose container
			h := make([]*Commit, len(p.Commits))
			for idx, c := range p.Commits {
				h[idx] = &Commit{
					Timestamp: c.Timestamp,
					ID:        c.ID,
					URL:       c.URL,
					Branch:    branch,
					Message:   c.Message,
					Username:  c.Author.Username,
					RealName:  c.Author.RealName,
					Path:      u.Hostname() + u.Path,
					TeXRoot:   r.URL.Query().Get("root"), // empty if not in query
				}
			}
			return h
		} else {
			return nil
		}
	} else if d == "gitlab.com" {
		// Get the push event
		p := new(gitlab.PushEvent)
		json.NewDecoder(r.Body).Decode(p)
		// get the url to be used as the path column (github.com/a/b)
		u, err := url.Parse(p.Repository.URL)
		ref := strings.Split(p.Ref, "/")
		branch := ref[len(ref)-1]
		if contains(strings.Split(branches, ","), branch) {
			if err != nil {
				panic(err)
			}
			// restruct the array of commits into a general purpose container
			h := make([]*Commit, len(p.Commits))
			for idx, c := range p.Commits {
				h[idx] = &Commit{
					Timestamp: c.Timestamp,
					ID:        c.ID,
					URL:       c.URL,
					Branch:    ref[len(ref)-1],
					Message:   c.Message,
					Username:  c.Author.Username,
					RealName:  c.Author.Username, // gitlab has no RealName
					Path:      u.Hostname() + u.Path,
					TeXRoot:   r.URL.Query().Get("root"), // empty if not in query
				}
			}
			return h
		} else {
			return nil
		}
	} else if d == "bitbucket.com" {
		//var p bitbucket.PushEvent
		return make([]*Commit, 5)
	} else {
		return make([]*Commit, 5)
		// TODOs
		// 1) log the fact that we accessed an unsupported domain.
		// 2) Return a page to the user that suggests they contact the CI admin to
		// implement this  domain
	}
}

func contains(a []string, v string) bool {
	for _, av := range a {
		if (len(a) == 1 && a[0] == "") || av == v {
			return true
		}
	}
	return false
}
