package models

type Anime struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Poster   string    `json:"poster"`
	Type     string    `json:"type"`
	Synopsis string    `json:"synopsis"`
	State    string    `json:"state"`
	Episodes []Episode `json:"episodes"`
}

type Episode struct {
	Id      string `json:"id"`
	Episode string `json:"episode"`
}

type Server struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Slug struct {
	Name   string  `json:"name"`
	Animes []Anime `json:"animes"`
}
