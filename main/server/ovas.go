package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/anaskhan96/soup"
	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/models"
)

func ovas() ([]models.Anime, error) {
	animes := []models.Anime{}

	client := http.DefaultClient

	http.DefaultClient.Transport = config.AddCloudFlareByPass(http.DefaultClient.Transport)

	res, err := client.Get(fmt.Sprintf("%v%s", config.Rooturl, config.Ovasurl))
	if err != nil {
		log.Fatal(err)
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	responseString := string(responseData)

	doc := soup.HTMLParse(responseString)

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
