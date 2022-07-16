package server

import (
	"fmt"
	"log"
	"net/http"

	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/models"
)

var bearer string

func (a *Server) IndexHandler() http.HandlerFunc {
	bearer = config.Config().Bearer
	fmt.Println(bearer)

	return func(w http.ResponseWriter, r *http.Request) {

		greet := config.Greet{
			Message: fmt.Sprintf("Gosanime Api vr: %v is Running.", config.Config().Version),
			Status:  "OK",
			Code:    "200",
		}

		SendResponse(w, r, greet, http.StatusOK)
	}
}

func (a *Server) GetMain() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			SendResponse(w, r, config.GetNoBearer(), http.StatusUnauthorized)
		} else {

			animes, err := main()
			if err != nil {
				log.Printf("cant get latest animes err=%v \n", err)
				SendResponse(w, r, nil, http.StatusInternalServerError)
				return
			}

			ovas, err := ovas()

			if err != nil {
				log.Printf("cant get latest animes err=%v \n", err)
				SendResponse(w, r, nil, http.StatusInternalServerError)
				return
			}

			topAnimes, err := top()

			if err != nil {
				log.Printf("cant get latest animes err=%v \n", err)
				SendResponse(w, r, nil, http.StatusInternalServerError)
				return
			}

			var animeResp = make([]models.Anime, len(animes))
			for ifx, anime := range animes {
				animeResp[ifx] = MapToJson(&anime)
			}

			var ovasResp = make([]models.Anime, len(ovas))
			for ifx, ova := range ovas {
				ovasResp[ifx] = MapToJson(&ova)
			}

			var topResp = make([]models.Anime, len(topAnimes))
			for tp, anime := range topAnimes {
				topResp[tp] = MapToJson(&anime)
			}

			p := []models.SlugResponse{}

			var ur = MakeSlugArrayResponse(topResp, "Top Animes")
			var ar = MakeSlugArrayResponse(animeResp, "Ultimos Agregados")
			var or = MakeSlugArrayResponse(ovasResp, "Ovas")

			p = append(p, ur)
			p = append(p, ar)
			p = append(p, or)

			SendResponse(w, r, MakeSlugMainResponse(p), http.StatusOK)
		}
	}
}

func (a *Server) GetOvas() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			SendResponse(w, r, config.GetNoBearer(), http.StatusUnauthorized)
		} else {
			animes, err := ovas()
			if err != nil {
				log.Printf("cant get latest animes err=%v \n", err)
				SendResponse(w, r, nil, http.StatusInternalServerError)
				return
			}

			var resp = make([]models.Anime, len(animes))
			for ifx, anime := range animes {
				resp[ifx] = MapToJson(&anime)
			}

			SendResponse(w, r, MakeArrayResponse(resp), http.StatusOK)
		}
	}
}

func (a *Server) GetTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			SendResponse(w, r, config.GetNoBearer(), http.StatusUnauthorized)
		} else {
			animes, err := tags(r)

			errorResponse := models.SearchAnimeResponse{
				Data:    nil,
				Status:  "401",
				Message: "Not Found",
				Page:    -1,
			}

			if err != nil {
				SendResponse(w, r, errorResponse, http.StatusNotFound)
			} else {
				SendResponse(w, r, animes, http.StatusOK)
			}
		}
	}
}

func (a *Server) GetAnime() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			SendResponse(w, r, config.GetNoBearer(), http.StatusUnauthorized)
		} else {
			anime, err := anime(r)
			if err != nil {
				log.Printf("cant get latest animes err=%v \n", err)
				SendResponse(w, r, nil, http.StatusInternalServerError)
				return
			}

			var response = models.Response{
				Data:    anime,
				Status:  "200",
				Message: "Success",
			}

			SendResponse(w, r, response, http.StatusOK)
		}
	}
}

func (a *Server) GetVideoServers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			SendResponse(w, r, config.GetNoBearer(), http.StatusUnauthorized)
		} else {
			episodes, err := videosByServer(r)

			if err != nil {
				log.Printf("cant get latest animes err=%v \n", err)
				SendResponse(w, r, nil, http.StatusInternalServerError)
				return
			}

			var response = models.Response{
				Data:    episodes,
				Status:  "200",
				Message: "Success",
			}

			SendResponse(w, r, response, http.StatusOK)

		}
	}
}

func (a *Server) SearchAnime() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			SendResponse(w, r, config.GetNoBearer(), http.StatusUnauthorized)
		} else {
			animes, err := searchAnime(r)

			errorResponse := models.SearchAnimeResponse{
				Data:    nil,
				Status:  "401",
				Message: "Not Found",
				Page:    -1,
			}

			if err != nil {
				SendResponse(w, r, errorResponse, http.StatusNotFound)
			} else {
				SendResponse(w, r, animes, http.StatusOK)
			}
		}
	}
}
