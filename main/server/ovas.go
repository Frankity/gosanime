package server

import (
	"fmt"
	"log"
	"strings"

	"github.com/anaskhan96/soup"
	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/models"
	"xyz.frankity/gosanime/main/utils"
)

// ovas fetches a list of OVA (Original Video Animation) series from the source.
// It scrapes the configured Ovasurl to gather information about OVAs.
// Returns a slice of models.Anime, where each Anime struct represents an OVA,
// and an error if any issues occur during the scraping or processing.
// Note: Similar to other scraping functions, errors leading to log.Fatal will terminate
// the application instead of returning an error to the caller.
func ovas() ([]models.Anime, error) {
	var err error // err is declared but not meaningfully used if log.Fatal is called.
	animes := []models.Anime{}

	client := utils.NewHTTPClient()

	res, err := client.R().Get(fmt.Sprintf("%v%s", config.Rooturl, config.Ovasurl))
	if err != nil {
		log.Fatal(err) // Critical error, application will exit.
	}

	responseString := res.String()

	doc := soup.HTMLParse(responseString)

	main := doc.Find("div", "class", "anime__page__content").FindAll("div", "class", "anime__item")

	for _, p := range main {
		anime := models.Anime{
			ID:     strings.Split(p.Find("h5").Find("a").Attrs()["href"], "/")[3],
			Name:   p.Find("h5").Find("a").Text(),
			Poster: p.Find("a").Find("div", "class", "anime__item__pic").Attrs()["data-setbg"],
			State:  p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[0].Text(),
			Type:   strings.TrimSpace(p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[1].Text()),
			// Synopsis, Genre, Episodes are not populated here, consider if they should be.
		}
		animes = append(animes, anime)
	}

	return animes, nil // Error will always be nil here if log.Fatal is used.
}
