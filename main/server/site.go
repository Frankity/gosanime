package server

import (
	"log"
	"strings"

	"github.com/anaskhan96/soup"
	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/models"
	"xyz.frankity/gosanime/main/utils"
)

// main (consider renaming for clarity, e.g., fetchMainSiteData) fetches the list of anime
// displayed on the main page of the source site, typically trending or new anime.
// It scrapes the configured Rooturl (base URL of the site).
// Returns a slice of models.Anime, where each Anime struct contains information
// about an anime series, and an error if any issues occur during scraping or processing.
// Note: Errors causing log.Fatal will terminate the application.
func main() ([]models.Anime, error) {
	var err error // err is declared but not meaningfully used if log.Fatal is called.
	animes := []models.Anime{}

	client := utils.NewHTTPClient()

	resp, err := client.R().Get(config.Rooturl)
	if err != nil {
		log.Fatal(err) // Critical error, application will exit.
	}

	responseString := resp.String()

	doc := soup.HTMLParse(responseString)
	// Finds the section assumed to contain trending anime items.
	mainContent := doc.Find("div", "class", "trending__anime").FindAll("div", "class", "anime__item")

	for _, p := range mainContent {
		anime := models.Anime{
			ID:     strings.Split(p.Find("h5").Find("a").Attrs()["href"], "/")[3],
			Name:   p.Find("h5").Find("a").Text(),
			Poster: p.Find("a").Find("div", "class", "anime__item__pic").Attrs()["data-setbg"],
			State:  p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[0].Text(),
			Type:   strings.TrimSpace(p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[1].Text()),
			// Synopsis, Genre, Episodes are not populated here.
		}
		animes = append(animes, anime)
	}

	return animes, nil // Error will always be nil here if log.Fatal is used.
}
