package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/anaskhan96/soup"
	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/models"
	"xyz.frankity/gosanime/main/utils"
)

func top() ([]models.Anime, error) {
	client := utils.NewHTTPClient()

	res, err := client.R().Get(fmt.Sprintf("%v%s", config.Rooturl, config.TopUrl))
	if err != nil {
		log.Fatal(err)
	}

	doc := soup.HTMLParse(res.String())
	items := doc.FindAll("div", "class", "toplist")

	animes := make([]models.Anime, 0, len(items))
	for _, p := range items {
		href := p.Find("a").Attrs()["href"]
		parts := strings.Split(strings.TrimRight(href, "/"), "/")
		id := parts[len(parts)-1]

		animes = append(animes, models.Anime{
			ID:       id,
			Name:     strings.TrimSpace(p.Find("h5").Text()),
			Poster:   p.Find("img").Attrs()["src"],
			Synopsis: strings.TrimSpace(p.Find("div", "class", "card-synopsis").Text()),
		})
	}

	return animes, nil
}

func anime(r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	id := r.Form.Get("id")

	client := utils.NewHTTPClient()

	res, err := client.R().Get(fmt.Sprintf("%v/%s", config.Rooturl, id))
	if err != nil {
		log.Fatal(err)
	}

	doc := soup.HTMLParse(res.String())

	name := strings.TrimSpace(doc.Find("div", "class", "anime_info").Find("h3").Text())
	poster := doc.Find("div", "class", "anime_pic").Find("img").Attrs()["src"]
	synopsis := strings.TrimSpace(doc.Find("div", "class", "anime_info").Find("p", "class", "scroll").Text())

	cardBod := doc.Find("div", "class", "card-bod")

	typ := strings.TrimSpace(cardBod.Find("li", "rel", "tipo").Text())

	genres := []string{}
	episodes := ""
	state := ""

	for _, li := range cardBod.Find("ul").FindAll("li") {
		label := strings.TrimSpace(li.Find("span").Text())
		switch {
		case strings.Contains(label, "Generos"):
			for _, a := range li.FindAll("a") {
				genres = append(genres, strings.ToLower(strings.TrimSpace(a.Text())))
			}
		case strings.Contains(label, "Episodios"):
			episodes = strings.TrimSpace(li.Text())
		case strings.Contains(label, "Estado"):
			state = strings.TrimSpace(li.Find("div", "class", "enemision").Text())
		}
	}

	an := models.Anime{
		ID:       id,
		Name:     name,
		Poster:   poster,
		Type:     typ,
		Synopsis: synopsis,
		Genre:    genres,
		State:    state,
		Episodes: episodes,
	}

	return an, nil
}

func searchAnime(r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	animeQuery := r.Form.Get("anime")

	url := fmt.Sprintf("%v/buscar/%s/", config.Rooturl, strings.Replace(animeQuery, "-", "_", -1))

	client := utils.NewHTTPClient()

	res, err := client.R().Get(url)
	if err != nil {
		log.Fatal(err)
	}

	doc := soup.HTMLParse(res.String())
	elements := doc.FindAll("div", "class", "anime__item")

	animes := []models.Anime{}
	for _, p := range elements {
		href := p.Find("h5").Find("a").Attrs()["href"]
		parts := strings.Split(strings.TrimRight(href, "/"), "/")
		animes = append(animes, models.Anime{
			ID:     parts[len(parts)-1],
			Name:   p.Find("h5").Find("a").Text(),
			Poster: p.Find("a").Find("div", "class", "anime__item__pic").Attrs()["data-setbg"],
			State:  p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[0].Text(),
			Type:   strings.TrimSpace(p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[1].Text()),
		})
	}

	return models.SearchAnimeResponse{
		Data:    animes,
		Status:  "200",
		Message: "Success",
		Page:    -1,
	}, err
}
