package models

type Anime struct {
	ID       string    `json:"ID"`
	Name     string    `json:"name"`
	Poster   string    `json:"poster"`
	Type     string    `json:"type"`
	Synopsis string    `json:"synopsis"`
	State    string    `json:"state"`
	Episodes []Episode `json:"episodes"`
}

type Episode struct {
	Id      string
	Episode string
}
