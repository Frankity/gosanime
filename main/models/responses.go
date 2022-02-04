package models

type ArrayResponse struct {
	Data    []Anime `json:"data"`
	Status  string  `json:"status"`
	Message string  `json:"message"`
}

type SearchAnimeResponse struct {
	Data    []Anime     `json:"data"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Page    interface{} `json:"page"`
}

type Response struct {
	Data    interface{} `json:"data"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
}

type SlugResponse struct {
	Data interface{} `json:"animes"`
	Name string      `json:"name"`
}

type SlugMainResponse struct {
	Data    []SlugResponse `json:"data"`
	Status  string         `json:"status"`
	Message string         `json:"message"`
}
