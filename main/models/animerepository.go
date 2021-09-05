package models

type AnimeRepository interface {
	GetTopAnimes() ([]*Anime, error)
}
