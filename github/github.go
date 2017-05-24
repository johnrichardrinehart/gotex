package github

type PushEvent struct {
	Repository json_repo        `json:"repository"`
	HC         json_head_commit `json:"head_commit"`
}

type json_repo struct {
	URL string `json:"url"`
}

type json_head_commit struct {
	ID        string         `json:"id"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
	Committer committer_repo `json:"committer"`
}

type committer_repo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
