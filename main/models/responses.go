package models

type ArrayResponse struct {
	Data    []Anime
	Status  string
	Message string
}

type SearchAnimeResponse struct {
	Data    []Anime
	Status  string
	Message string
	Page    interface{}
}

type Response struct {
	Data    interface{}
	Status  string
	Message string
}
