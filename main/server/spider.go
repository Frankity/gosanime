package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
	"xyz.frankity/gosanime/main/models"
)

func lastAnimes() ([]models.Anime, error) {
	animes := []models.Anime{}
	resp, err := soup.Get(ROOTURL.url)
	if err != nil {
		os.Exit(1)
	}

	doc := soup.HTMLParse(resp)

	main := doc.Find("div", "class", "trending__anime").FindAll("div", "class", "anime__item")

	for _, p := range main {
		anime := models.Anime{
			ID:     strings.Split(p.Find("h5").Find("a").Attrs()["href"], "/")[3],
			Name:   p.Find("h5").Find("a").Text(),
			Poster: p.Find("a").Find("div", "class", "anime__item__pic").Attrs()["data-setbg"],
			State:  p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[0].Text(),
			Type:   strings.TrimSpace(p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[1].Text()),
		}
		animes = append(animes, anime)
	}
	return animes, nil
}

func ovas() ([]models.Anime, error) {
	animes := []models.Anime{}
	resp, err := soup.Get(fmt.Sprintf("%v%s", ROOTURL.url, OVASURL.url))
	if err != nil {
		os.Exit(1)
	}

	doc := soup.HTMLParse(resp)

	main := doc.Find("div", "class", "anime__page__content").FindAll("div", "class", "anime__item")

	for _, p := range main {
		anime := models.Anime{
			ID:     strings.Split(p.Find("h5").Find("a").Attrs()["href"], "/")[3],
			Name:   p.Find("h5").Find("a").Text(),
			Poster: p.Find("a").Find("div", "class", "anime__item__pic").Attrs()["data-setbg"],
			State:  p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[0].Text(),
			Type:   strings.TrimSpace(p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[1].Text()),
		}
		animes = append(animes, anime)
	}

	return animes, nil

}

func anime(r *http.Request) (interface{}, error) {

	if err := r.ParseForm(); err != nil {
		os.Exit(1)
	}
	x := r.Form.Get("id")

	resp, err := soup.Get(fmt.Sprintf("%v/%s", ROOTURL.url, x))
	if err != nil {
		os.Exit(1)
	}

	print(fmt.Sprintf("%v/%s", ROOTURL.url, x))

	doc := soup.HTMLParse(resp)

	episodes := doc.Find("div", "class", "anime__pagination").FindAll("a", "class", "numbers")

	lastEp := strings.Split(strings.TrimSpace(episodes[len(episodes)-1].Text()), "-")[1]
	lastEpIntVal, err := strconv.Atoi(strings.TrimSpace(lastEp))

	eplist := []models.Episode{}
	for i := 0; i < lastEpIntVal; i++ {
		ep := models.Episode{
			Id:      strconv.Itoa(i),
			Episode: x,
		}
		eplist = append(eplist, ep)
	}

	an := models.Anime{
		ID:       x,
		Name:     doc.Find("div", "class", "anime__details__content").Find("h3").Text(),
		Poster:   doc.Find("div", "class", "anime__details__pic").Attrs()["data-setbg"],
		Type:     strings.TrimSpace(doc.Find("div", "class", "anime__details__widget").Find("ul").FindAll("li")[0].Text()),
		Synopsis: doc.Find("div", "class", "anime__details__content").Find("p").Text(),
		State:    strings.TrimSpace(doc.Find("div", "class", "anime__details__widget").Find("ul").FindAll("li")[6].Find("span", "class", "enemision").Text()),
		Episodes: eplist,
	}

	return an, err

}

func (a *Server) IndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Gosanime Api Running")
	}
}

func (a *Server) GetTopAnimes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		animes, err := lastAnimes()
		if err != nil {
			log.Printf("cant get latest animes err=%v \n", err)
			sendResponse(w, r, nil, http.StatusInternalServerError)
			return
		}

		var resp = make([]models.Anime, len(animes))
		for ifx, anime := range animes {
			resp[ifx] = mapToJson(&anime)
		}

		sendResponse(w, r, makeResponse(resp), http.StatusOK)
	}
}

func (a *Server) GetOvas() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		animes, err := ovas()
		if err != nil {
			log.Printf("cant get latest animes err=%v \n", err)
			sendResponse(w, r, nil, http.StatusInternalServerError)
			return
		}

		var resp = make([]models.Anime, len(animes))
		for ifx, anime := range animes {
			resp[ifx] = mapToJson(&anime)
		}

		sendResponse(w, r, makeResponse(resp), http.StatusOK)
	}
}

func (a *Server) GetAnime() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		animes, err := anime(r)
		if err != nil {
			log.Printf("cant get latest animes err=%v \n", err)
			sendResponse(w, r, nil, http.StatusInternalServerError)
			return
		}

		sendResponse(w, r, animes, http.StatusOK)
	}
}

type Response struct {
	Data    []models.Anime
	Status  string
	Message string
}

func makeResponse(animes []models.Anime) interface{} {
	return Response{
		Data:    animes,
		Status:  "200",
		Message: "Success",
	}
}
