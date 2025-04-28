package server

import (
	"log"
	"strings"

	"github.com/anaskhan96/soup"
	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/models"
	"xyz.frankity/gosanime/main/utils"
)

func main() ([]models.Anime, error) {
	var err error
	animes := []models.Anime{}

	client := utils.NewHTTPClient()

	resp, err := client.R().Get(config.Rooturl)
	if err != nil {
		log.Fatal(err)
	}

	responseString := resp.String()

	doc := soup.HTMLParse(responseString)
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
