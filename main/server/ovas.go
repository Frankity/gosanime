package server

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/models"
	"xyz.frankity/gosanime/main/utils"
)


type ovaPage struct {
	Data []models.OvaItem `json:"data"`
}

var ovaJSONRE = regexp.MustCompile(`var animes = (\{.+?\});`)

func ovas() ([]models.Anime, error) {
	client := utils.NewHTTPClient()

	res, err := client.R().Get(fmt.Sprintf("%v%s", config.Rooturl, config.Ovasurl))
	if err != nil {
		log.Fatal(err)
	}

	m := ovaJSONRE.FindSubmatch(res.Body())
	if m == nil {
		log.Fatal("ovas: var animes not found in page")
	}

	var page ovaPage
	if err := json.Unmarshal(m[1], &page); err != nil {
		log.Fatal("ovas: failed to parse JSON:", err)
	}

	animes := make([]models.Anime, 0, len(page.Data))
	for _, item := range page.Data {
		animes = append(animes, models.Anime{
			ID:     item.Slug,
			Name:   item.Title,
			Poster: item.Image,
			State:  item.Estado,
			Type:   item.Tipo,
		})
	}

	return animes, nil
}
