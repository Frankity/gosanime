package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/models"
)

func top() ([]models.Anime, error) {
	animes := []models.Anime{}

	client := http.DefaultClient

	http.DefaultClient.Transport = config.AddCloudFlareByPass(http.DefaultClient.Transport)

	res, err := client.Get(fmt.Sprintf("%v%s", config.Rooturl, config.TopUrl))
	if err != nil {
		log.Fatal(err)
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	responseString := string(responseData)

	doc := soup.HTMLParse(responseString)

	main := doc.Find("section", "class", "contenido").Find("div", "class", "container").Find("div", "class", "row").Find("div", "class", "col-lg-12").FindAll("div", "class", "list")

	for _, p := range main {
		parent := p.Find("div", "id", "conb")
		anime := models.Anime{
			ID:       strings.Split(parent.Find("a").Attrs()["href"], "/")[3],
			Name:     parent.Find("a").Attrs()["title"],
			Poster:   p.Find("a").Find("img").Attrs()["src"],
			Synopsis: strings.TrimSpace(strings.Split(p.Find("div", "id", "animinfo").Find("span", "class", "title").Text(), "/")[1]),
			State:    strings.TrimSpace(p.Find("div", "id", "animinfo").Find("h2", "class", "portada-title").Text()),
			Type:     strings.TrimSpace(strings.Split(p.Find("div", "id", "animinfo").Find("span", "class", "title").Text(), "/")[0]),
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

	client := http.DefaultClient

	http.DefaultClient.Transport = config.AddCloudFlareByPass(http.DefaultClient.Transport)

	res, err := client.Get(fmt.Sprintf("%v/%s", config.Rooturl, x))
	if err != nil {
		log.Fatal(err)
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	responseString := string(responseData)

	doc := soup.HTMLParse(responseString)

	var episodesData string

	episodes := doc.Find("div", "class", "anime__pagination").FindAll("a", "class", "numbers")

	episodeNumber := strings.TrimSpace(doc.Find("div", "class", "anime__details__widget").Find("ul").FindAll("li")[5].Text())

	lastEp := strings.Split(strings.TrimSpace(episodes[len(episodes)-1].Text()), "-")[1]

	genres := doc.Find("div", "class", "anime__details__widget").Find("ul").FindAll("li")[1].FindAll("a")

	result := []string{}

	for a := range genres {
		result = append(result, strings.ToLower(genres[a].Text()))
	}

	strings.Join(result, ",")

	if episodeNumber == "Desconocido" {
		episodesData = lastEp
	} else {
		episodesData = episodeNumber
	}

	an := models.Anime{
		ID:       x,
		Name:     doc.Find("div", "class", "anime__details__content").Find("h3").Text(),
		Poster:   doc.Find("div", "class", "anime__details__pic").Attrs()["data-setbg"],
		Type:     strings.TrimSpace(doc.Find("div", "class", "anime__details__widget").Find("ul").FindAll("li")[0].Text()),
		Synopsis: doc.Find("div", "class", "anime__details__text").Find("p").Text(),
		Genre:    result,
		State:    doc.Find("div", "class", "anime__details__widget").Find("ul").FindAll("li")[8].Children()[2].Text(),
		Episodes: strings.TrimSpace(episodesData),
	}

	return an, err
}

func searchAnime(r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	anime := r.Form.Get("anime")
	page := r.Form.Get("page")

	url := fmt.Sprintf("%v/buscar/%s/%s/", config.Rooturl, strings.Replace(anime, "-", "_", -1), page)

	client := http.DefaultClient

	http.DefaultClient.Transport = config.AddCloudFlareByPass(http.DefaultClient.Transport)

	res, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	responseString := string(responseData)

	doc := soup.HTMLParse(responseString)

	elements := doc.FindAll("div", "class", "anime__item")

	animes := []models.Anime{}

	for _, p := range elements {
		anime := models.Anime{
			ID:     strings.Split(p.Find("h5").Find("a").Attrs()["href"], "/")[3],
			Name:   p.Find("h5").Find("a").Text(),
			Poster: p.Find("a").Find("div", "class", "anime__item__pic").Attrs()["data-setbg"],
			State:  p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[0].Text(),
			Type:   strings.TrimSpace(p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[1].Text()),
		}
		animes = append(animes, anime)
	}

	pg, err := strconv.Atoi(page)
	if err != nil {
		print(fmt.Sprintf("%v %s", err, "<- Error"))
	}

	if len(elements) == 12 {
		pg = pg + 1
	} else {
		pg = -1
	}

	ar := models.SearchAnimeResponse{
		Data:    animes,
		Status:  "200",
		Message: "Success",
		Page:    pg,
	}

	return ar, err
}
