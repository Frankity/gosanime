package server

import (
	"encoding/json"
	"log"
	"net/http"

	"xyz.frankity/gosanime/main/models"
)

func sendResponse(w http.ResponseWriter, _ *http.Request, data interface{}, status int) {
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

func mapToJson(a *models.Anime) models.Anime {
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
