package server

import (
	"fmt"
	"log"
	"net/http"

	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/models"
)

// bearer is a package-level variable storing the bearer token for API authorization.
// It is initialized in IndexHandler from config.
// Note: Global variables can introduce complexities in testing and concurrent access
// if not managed carefully.
var bearer string

// IndexHandler handles requests to the root ("/") endpoint of the API.
// It initializes the package-level bearer token from the application configuration.
// It returns a welcome message indicating the API version and status.
func (a *Server) IndexHandler() http.HandlerFunc {
	bearer = config.Config().Bearer // Initializes the global bearer token.
	fmt.Println(bearer)             // Logs the bearer token, consider removing for production.

	return func(w http.ResponseWriter, r *http.Request) {
		greet := config.Greet{
			Message: fmt.Sprintf("Gosanime Api v: %v is Running.", config.Config().Version),
			Status:  "OK",
			Code:    "200",
		}
		SendResponse(w, r, greet, http.StatusOK)
	}
}

// GetMain handles requests to the "/api/v1/main" endpoint.
// It requires Bearer token authorization.
// It fetches data for the main page, including top anime, latest additions, and OVAs,
// then aggregates them into a structured response.
// Returns an HTTP 500 error if any of the data fetching operations fail.
func (a *Server) GetMain() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r) // Logs the request, consider removing or using a more structured logger.
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			SendResponse(w, r, config.GetNoBearer(), http.StatusUnauthorized)
			return
		}

		animes, err := main() // Fetches latest animes
		if err != nil {
			log.Printf("cant get latest animes err=%v \n", err)
			SendResponse(w, r, nil, http.StatusInternalServerError)
			return
		}

		ovasResult, err := ovas() // Fetches OVAs
		if err != nil {
			log.Printf("cant get latest ovas err=%v \n", err) // Corrected log message
			SendResponse(w, r, nil, http.StatusInternalServerError)
			return
		}

		topAnimes, err := top() // Fetches top animes
		if err != nil {
			log.Printf("cant get top animes err=%v \n", err) // Corrected log message
			SendResponse(w, r, nil, http.StatusInternalServerError)
			return
		}

		// The MapToJson function is not defined in the provided code, assuming it exists elsewhere
		// and prepares anime data for the response.
		var animeResp = make([]models.Anime, len(animes))
		for ifx, animeItem := range animes { // Renamed 'anime' to 'animeItem' to avoid conflict
			animeResp[ifx] = MapToJson(&animeItem)
		}

		var ovasResp = make([]models.Anime, len(ovasResult))
		for ifx, ovaItem := range ovasResult { // Renamed 'ova' to 'ovaItem'
			ovasResp[ifx] = MapToJson(&ovaItem)
		}

		var topResp = make([]models.Anime, len(topAnimes))
		for tp, animeItem := range topAnimes { // Renamed 'anime' to 'animeItem'
			topResp[tp] = MapToJson(&animeItem)
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

// GetOvas handles requests to the "/api/v1/ovas" endpoint.
// It requires Bearer token authorization.
// It fetches a list of OVAs and returns them in a structured response.
// Returns an HTTP 500 error if fetching OVAs fails.
func (a *Server) GetOvas() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			SendResponse(w, r, config.GetNoBearer(), http.StatusUnauthorized)
			return
		}
		ovasResult, err := ovas() // Fetches OVAs
		if err != nil {
			log.Printf("cant get ovas err=%v \n", err) // Corrected log message
			SendResponse(w, r, nil, http.StatusInternalServerError)
			return
		}

		var resp = make([]models.Anime, len(ovasResult))
		for ifx, animeItem := range ovasResult { // Renamed 'anime' to 'animeItem'
			resp[ifx] = MapToJson(&animeItem)
		}

		SendResponse(w, r, MakeArrayResponse(resp), http.StatusOK)
	}
}

// GetTag handles requests to the "/api/v1/tags" endpoint.
// It requires Bearer token authorization.
// It fetches anime based on a tag specified in the request (details of how tag is passed are in `tags(r)`).
// Returns a list of anime or an error response if fetching fails or no anime are found.
func (a *Server) GetTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			SendResponse(w, r, config.GetNoBearer(), http.StatusUnauthorized)
			return
		}
		// The `tags(r)` function is responsible for parsing parameters from `r` (e.g., tag name, page).
		animesResult, err := tags(r)

		errorResponse := models.SearchAnimeResponse{
			Data:    nil,
			Status:  "401", // Should probably be 404 if not found, or 500 for server error.
			Message: "Not Found",
			Page:    -1,
		}

		if err != nil {
			// Consider logging the actual error from `tags(r)`
			SendResponse(w, r, errorResponse, http.StatusNotFound)
		} else {
			SendResponse(w, r, animesResult, http.StatusOK)
		}
	}
}

// GetAnime handles requests to the "/api/v1/anime" endpoint.
// It requires Bearer token authorization.
// It fetches detailed information for a specific anime based on parameters in the request (details in `anime(r)`).
// Returns the anime details or an error response if fetching fails.
func (a *Server) GetAnime() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			SendResponse(w, r, config.GetNoBearer(), http.StatusUnauthorized)
			return
		}
		// The `anime(r)` function is responsible for parsing parameters from `r` (e.g., anime ID).
		animeResult, err := anime(r)
		if err != nil {
			log.Printf("cant get anime details err=%v \n", err) // Corrected log message
			SendResponse(w, r, nil, http.StatusInternalServerError)
			return
		}

		var response = models.Response{
			Data:    animeResult,
			Status:  "200",
			Message: "Success",
		}

		SendResponse(w, r, response, http.StatusOK)
	}
}

// GetVideoServers handles requests to the "/api/v1/video" endpoint.
// It requires Bearer token authorization.
// It fetches video server links for a specific anime episode based on parameters in the request (details in `videosByServer(r)`).
// Returns a list of video servers or an error response if fetching fails.
func (a *Server) GetVideoServers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			SendResponse(w, r, config.GetNoBearer(), http.StatusUnauthorized)
			return
		}
		// The `videosByServer(r)` function parses parameters from `r` (e.g., anime slug, episode number).
		episodes, err := videosByServer(r)

		if err != nil {
			log.Printf("cant get video servers err=%v \n", err) // Corrected log message
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

// SearchAnime handles requests to the "/api/v1/search" endpoint.
// It requires Bearer token authorization.
// It searches for anime based on query parameters in the request (details in `searchAnime(r)`).
// Returns a list of matching anime or an error response if the search fails or no results are found.
func (a *Server) SearchAnime() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			SendResponse(w, r, config.GetNoBearer(), http.StatusUnauthorized)
			return
		}
		// The `searchAnime(r)` function parses parameters from `r` (e.g., search query, page number).
		animesResult, err := searchAnime(r)

		errorResponse := models.SearchAnimeResponse{
			Data:    nil,
			Status:  "401", // Should be 404 if not found, or 500 for server error.
			Message: "Not Found",
			Page:    -1,
		}

		if err != nil {
			// Consider logging the actual error from `searchAnime(r)`
			SendResponse(w, r, errorResponse, http.StatusNotFound)
		} else {
			SendResponse(w, r, animesResult, http.StatusOK)
		}
	}
}
