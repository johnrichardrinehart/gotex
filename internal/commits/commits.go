package commits

type Commits []*commit

type commit struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	URL       string `json:"url"`
	Author    author `json:"author"`
}

type author struct {
	UserName string `json:"username"`
	RealName string `json:"realname"`
}

//type json_repo struct {
//URL string `json:"url"`
//}

//type json_head_commit struct {
//ID        string         `json:"id"`
//Message   string         `json:"message"`
//Timestamp string         `json:"timestamp"`
//Committer committer_repo `json:"committer"`
//CommitURL string         `json:"url"`
//Message   string         `json:"message"`
//}

//type committer_repo struct {
//Name  string `json:"name"`
//Email string `json:"email"`
//}
