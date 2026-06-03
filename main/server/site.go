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
	client := utils.NewHTTPClient()

	resp, err := client.R().Get(config.Rooturl)
	if err != nil {
		log.Fatal(err)
	}

	doc := soup.HTMLParse(resp.String())

	section := doc.Find("div", "class", "trending_div")
	thumbs := section.FindAll("div", "class", "custom_thumb_home")
	bodies := section.FindAll("div", "class", "card-body-home")

	animes := make([]models.Anime, 0, len(thumbs))

	for i, thumb := range thumbs {
		if i >= len(bodies) {
			break
		}
		a := thumb.Find("a")
		img := thumb.Find("img")
		info := bodies[i].Find("div", "class", "card-info").FindAll("p")

		href := a.Attrs()["href"]
		parts := strings.Split(strings.TrimRight(href, "/"), "/")
		id := parts[len(parts)-1]

		state := ""
		typ := ""
		if len(info) > 0 {
			state = strings.TrimSpace(info[0].Text())
		}
		if len(info) > 1 {
			typ = strings.TrimSpace(info[1].Text())
		}

		animes = append(animes, models.Anime{
			ID:     id,
			Name:   strings.TrimSpace(bodies[i].Find("h5").Find("a").Text()),
			Poster: img.Attrs()["src"],
			State:  state,
			Type:   typ,
		})
	}

	return animes, nil
}
