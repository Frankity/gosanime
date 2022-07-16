package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"xyz.frankity/gosanime/main/models"
)

func SendResponse(w http.ResponseWriter, _ *http.Request, data interface{}, status int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("Cannot format json. err=%v\n", err)
	}
}

func MapToJson(a *models.Anime) models.Anime {
	return models.Anime{
		ID:       a.ID,
		Name:     a.Name,
		Poster:   a.Poster,
		Type:     a.Type,
		Synopsis: a.Synopsis,
		State:    a.State,
		Episodes: a.Episodes,
	}
}

func GetHiddenUrl(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	return fmt.Sprint(resp.Request.URL)
}

type Url struct {
	File string
}

func MakeArrayResponse(animes []models.Anime) interface{} {
	return models.ArrayResponse{
		Data:    animes,
		Status:  "200",
		Message: "Success",
	}
}

func MakeSlugMainResponse(slugs []models.SlugResponse) models.SlugMainResponse {
	return models.SlugMainResponse{
		Data:    slugs,
		Status:  "200",
		Message: "Success",
	}
}

func MakeSlugArrayResponse(animes []models.Anime, name string) models.SlugResponse {
	return models.SlugResponse{
		Data: animes,
		Name: name,
	}
}
